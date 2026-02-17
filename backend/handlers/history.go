package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"succulent-identifier-backend/db"
	"succulent-identifier-backend/models"
)

// HistoryHandler handles history-related requests
type HistoryHandler struct {
	identificationRepo IdentificationRepositoryInterface
	chatRepo           ChatRepositoryInterface
}

// NewHistoryHandler creates a new history handler
func NewHistoryHandler(
	identificationRepo IdentificationRepositoryInterface,
	chatRepo ChatRepositoryInterface,
) *HistoryHandler {
	return &HistoryHandler{
		identificationRepo: identificationRepo,
		chatRepo:           chatRepo,
	}
}

// HandleList returns paginated list of identifications
func (h *HistoryHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	// Only accept GET requests
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // default
	offset := 0  // default

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 100 {
				limit = 100 // max limit
			}
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get identifications from database
	identifications, err := h.identificationRepo.GetAll(limit, offset)
	if err != nil {
		log.Printf("Failed to get identifications: %v", err)
		h.sendError(w, http.StatusInternalServerError, "Failed to retrieve history")
		return
	}

	// Get total count
	total, err := h.identificationRepo.Count()
	if err != nil {
		log.Printf("Failed to count identifications: %v", err)
		// Continue without total count
		total = 0
	}

	// Convert to response format
	items := make([]models.HistoryItem, 0, len(identifications))
	for _, ident := range identifications {
		// Extract filename from full path for the API response
		imagePath := ident.ImagePath
		if idx := strings.LastIndex(imagePath, "/"); idx != -1 {
			imagePath = imagePath[idx+1:]
		}

		items = append(items, models.HistoryItem{
			ID:         ident.ID,
			Genus:      ident.Genus,
			Species:    ident.Species,
			Confidence: ident.Confidence,
			ImagePath:  imagePath,
			CreatedAt:  ident.CreatedAt,
		})
	}

	response := models.HistoryListResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleGetByID returns detailed information about a specific identification
func (h *HistoryHandler) HandleGetByID(w http.ResponseWriter, r *http.Request) {
	// Only accept GET requests
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract ID from URL path
	// Expecting /history/:id
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		h.sendError(w, http.StatusBadRequest, "Missing identification ID")
		return
	}
	id := pathParts[1]

	// Get identification from database
	identification, err := h.identificationRepo.GetByID(id)
	if err != nil {
		log.Printf("Failed to get identification: %v", err)
		h.sendError(w, http.StatusNotFound, "Identification not found")
		return
	}

	// Convert care guide to response format
	var careGuide *models.CareInstructions
	if identification.CareGuide != nil {
		careGuide = &models.CareInstructions{
			Sunlight: identification.CareGuide.Sunlight,
			Watering: identification.CareGuide.Watering,
			Soil:     identification.CareGuide.Soil,
			Notes:    identification.CareGuide.Notes,
		}
	}

	// Extract filename from full path for the API response
	imagePath := identification.ImagePath
	if idx := strings.LastIndex(imagePath, "/"); idx != -1 {
		imagePath = imagePath[idx+1:]
	}

	response := models.HistoryDetailResponse{
		ID:         identification.ID,
		Genus:      identification.Genus,
		Species:    identification.Species,
		Confidence: identification.Confidence,
		ImagePath:  imagePath,
		CareGuide:  careGuide,
		CreatedAt:  identification.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleGetWithChat returns identification with its chat history
func (h *HistoryHandler) HandleGetWithChat(w http.ResponseWriter, r *http.Request) {
	// Only accept GET requests
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract ID from URL path
	// Expecting /history/:id/with-chat
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		h.sendError(w, http.StatusBadRequest, "Missing identification ID")
		return
	}
	id := pathParts[1]

	// Get identification from database
	identification, err := h.identificationRepo.GetByID(id)
	if err != nil {
		log.Printf("Failed to get identification: %v", err)
		h.sendError(w, http.StatusNotFound, "Identification not found")
		return
	}

	// Get chat messages
	chatMessages, err := h.chatRepo.GetByIdentificationID(id)
	if err != nil {
		log.Printf("Failed to get chat messages: %v", err)
		// Continue with empty chat history
		chatMessages = []db.ChatMessage{}
	}

	// Convert to response format
	var careGuide *models.CareInstructions
	if identification.CareGuide != nil {
		careGuide = &models.CareInstructions{
			Sunlight: identification.CareGuide.Sunlight,
			Watering: identification.CareGuide.Watering,
			Soil:     identification.CareGuide.Soil,
			Notes:    identification.CareGuide.Notes,
		}
	}

	// Extract filename from full path for the API response
	imagePath := identification.ImagePath
	if idx := strings.LastIndex(imagePath, "/"); idx != -1 {
		imagePath = imagePath[idx+1:]
	}

	messages := make([]models.ChatMessageResponse, 0, len(chatMessages))
	for _, msg := range chatMessages {
		messages = append(messages, models.ChatMessageResponse{
			ID:        msg.ID,
			Message:   msg.Message,
			Sender:    msg.Sender,
			CreatedAt: msg.CreatedAt,
		})
	}

	response := models.HistoryWithChatResponse{
		Identification: models.HistoryDetailResponse{
			ID:         identification.ID,
			Genus:      identification.Genus,
			Species:    identification.Species,
			Confidence: identification.Confidence,
			ImagePath:  imagePath,
			CareGuide:  careGuide,
			CreatedAt:  identification.CreatedAt,
		},
		ChatMessages: messages,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleGetChatHistory returns chat messages for a specific identification
func (h *HistoryHandler) HandleGetChatHistory(w http.ResponseWriter, r *http.Request) {
	// Only accept GET requests
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract ID from URL path
	// Expecting /chat/:identification_id
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		h.sendError(w, http.StatusBadRequest, "Missing identification ID")
		return
	}
	identificationID := pathParts[1]

	// Get chat messages
	chatMessages, err := h.chatRepo.GetByIdentificationID(identificationID)
	if err != nil {
		log.Printf("Failed to get chat messages: %v", err)
		h.sendError(w, http.StatusInternalServerError, "Failed to retrieve chat history")
		return
	}

	// Convert to response format
	messages := make([]models.ChatMessageResponse, 0, len(chatMessages))
	for _, msg := range chatMessages {
		messages = append(messages, models.ChatMessageResponse{
			ID:        msg.ID,
			Message:   msg.Message,
			Sender:    msg.Sender,
			CreatedAt: msg.CreatedAt,
		})
	}

	response := models.ChatHistoryResponse{
		IdentificationID: identificationID,
		Messages:         messages,
		Total:            len(messages),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// sendError sends an error response
func (h *HistoryHandler) sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
}
