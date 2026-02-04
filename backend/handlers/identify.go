package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"succulent-identifier-backend/models"
	"succulent-identifier-backend/utils"
)

// IdentifyHandler handles plant identification requests
type IdentifyHandler struct {
	mlClient         MLClientInterface
	careDataService  CareDataServiceInterface
	fileUploader     FileUploaderInterface
	speciesThreshold float64
}

// NewIdentifyHandler creates a new identify handler
func NewIdentifyHandler(
	mlClient MLClientInterface,
	careDataService CareDataServiceInterface,
	fileUploader FileUploaderInterface,
	speciesThreshold float64,
) *IdentifyHandler {
	return &IdentifyHandler{
		mlClient:         mlClient,
		careDataService:  careDataService,
		fileUploader:     fileUploader,
		speciesThreshold: speciesThreshold,
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
	response, err := h.processMLResponse(mlResponse)
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
func (h *IdentifyHandler) processMLResponse(mlResponse *models.MLInferenceResponse) (*models.IdentifyResponse, error) {
	// Get top prediction
	topPrediction := mlResponse.Predictions[0]

	// Parse label to extract genus and species
	genus, species := utils.ParseLabel(topPrediction.Label)

	// Initialize response
	response := &models.IdentifyResponse{
		Plant: models.PlantInfo{
			Genus:      utils.FormatGenus(genus),
			Confidence: topPrediction.Confidence,
		},
	}

	// Apply confidence threshold logic
	if topPrediction.Confidence >= h.speciesThreshold && species != "" {
		// High confidence: show species
		response.Plant.Species = utils.FormatSpecies(topPrediction.Label)
	}

	// Get care instructions (try species first, fall back to genus)
	care, err := h.careDataService.GetCareInstructions(species, genus)
	if err != nil {
		// If no care data found, return generic care instructions
		log.Printf("No care data found: %v", err)
		care = models.CareInstructions{
			Sunlight: "Information not available",
			Watering: "Information not available",
			Soil:     "Information not available",
			Notes:    "Care information is not available for this plant.",
		}
	}

	response.Care = care

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
