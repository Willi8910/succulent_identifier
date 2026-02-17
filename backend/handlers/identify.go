package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"succulent-identifier-backend/db"
	"succulent-identifier-backend/models"
	"succulent-identifier-backend/utils"
)

// IdentifyHandler handles plant identification requests
type IdentifyHandler struct {
	mlClient               MLClientInterface
	chatService            ChatServiceInterface
	careRepo               CareInstructionsRepositoryInterface
	fileUploader           FileUploaderInterface
	identificationRepo     IdentificationRepositoryInterface
	speciesThreshold       float64
}

// NewIdentifyHandler creates a new identify handler
func NewIdentifyHandler(
	mlClient MLClientInterface,
	chatService ChatServiceInterface,
	careRepo CareInstructionsRepositoryInterface,
	fileUploader FileUploaderInterface,
	identificationRepo IdentificationRepositoryInterface,
	speciesThreshold float64,
) *IdentifyHandler {
	return &IdentifyHandler{
		mlClient:               mlClient,
		chatService:            chatService,
		careRepo:               careRepo,
		fileUploader:           fileUploader,
		identificationRepo:     identificationRepo,
		speciesThreshold:       speciesThreshold,
	}
}

// Handle processes the identify request
func (h *IdentifyHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		h.sendError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// Get uploaded file
	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "No image file provided")
		return
	}
	defer file.Close()

	// Save uploaded file
	imagePath, err := h.fileUploader.SaveFile(file, fileHeader)
	if err != nil {
		log.Printf("File upload error: %v", err)
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Optional: Clean up file after processing (can be configured)
	// defer h.fileUploader.DeleteFile(imagePath)

	// Call ML service for inference
	mlResponse, err := h.mlClient.Infer(imagePath)
	if err != nil {
		log.Printf("ML inference error: %v", err)
		h.sendError(w, http.StatusInternalServerError, "Failed to identify plant")
		return
	}

	// Process predictions with confidence threshold logic
	response, err := h.processMLResponse(mlResponse, imagePath)
	if err != nil {
		log.Printf("Processing error: %v", err)
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Send successful response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// processMLResponse processes ML predictions and applies confidence threshold logic
func (h *IdentifyHandler) processMLResponse(mlResponse *models.MLInferenceResponse, imagePath string) (*models.IdentifyResponse, error) {
	// Get top prediction
	topPrediction := mlResponse.Predictions[0]

	// Parse label to extract genus and species
	genus, species := utils.ParseLabel(topPrediction.Label)

	// Apply confidence threshold logic
	var displaySpecies string
	if topPrediction.Confidence >= h.speciesThreshold && species != "" {
		// High confidence: show species
		displaySpecies = utils.FormatSpecies(topPrediction.Label)
	}

	// Get care instructions with caching strategy
	var careGuide *db.CareGuide

	// Check cache first
	cachedCare, err := h.careRepo.GetBySpecies(genus, species)
	if err != nil {
		log.Printf("Error checking care cache: %v", err)
	}

	if cachedCare != nil {
		// Use cached care instructions
		log.Printf("Using cached care instructions for %s %s", genus, species)
		careGuide = cachedCare.CareGuide
	} else {
		// Generate new care instructions with LLM
		log.Printf("Generating new care instructions for %s %s", genus, species)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		careGuide, err = h.chatService.GenerateCareInstructions(ctx, genus, species)
		if err != nil {
			log.Printf("Failed to generate care instructions: %v", err)
			// Fallback to generic instructions if LLM fails
			careGuide = &db.CareGuide{
				Sunlight: "Provide bright, indirect light for most succulents.",
				Watering: "Water when soil is completely dry. Succulents prefer infrequent, deep watering.",
				Soil:     "Use well-draining cactus or succulent mix.",
				Notes:    "Care information could not be generated. These are general succulent care guidelines.",
			}
		} else {
			// Save to cache for future use
			cacheEntry := &db.CareInstructionsCache{
				ID:        uuid.New().String(),
				Genus:     genus,
				Species:   species,
				CareGuide: careGuide,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if err := h.careRepo.Create(cacheEntry); err != nil {
				log.Printf("Failed to cache care instructions: %v", err)
				// Don't fail the request, just log the error
			} else {
				log.Printf("Care instructions cached for %s %s", genus, species)
			}
		}
	}

	// Convert to response format
	care := models.CareInstructions{
		Sunlight: careGuide.Sunlight,
		Watering: careGuide.Watering,
		Soil:     careGuide.Soil,
		Notes:    careGuide.Notes,
	}

	// Generate UUID for identification
	identificationID := uuid.New().String()

	// Create identification record for database
	identification := &db.Identification{
		ID:         identificationID,
		Genus:      genus,
		Species:    species,
		Confidence: topPrediction.Confidence,
		ImagePath:  imagePath,
		CareGuide:  careGuide,
		CreatedAt:  time.Now(),
	}

	// Save to database
	if err := h.identificationRepo.Create(identification); err != nil {
		log.Printf("Failed to save identification to database: %v", err)
		// Note: We don't fail the request if DB save fails, just log the error
		// The user still gets their identification result
	} else {
		log.Printf("Identification saved to database with ID: %s", identificationID)
	}

	// Build response
	response := &models.IdentifyResponse{
		ID: identificationID,
		Plant: models.PlantInfo{
			Genus:      utils.FormatGenus(genus),
			Species:    displaySpecies,
			Confidence: topPrediction.Confidence,
		},
		Care: care,
	}

	return response, nil
}

// sendError sends an error response
func (h *IdentifyHandler) sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
}
