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
	"succulent-identifier-backend/services"
)

// ChatHandler handles chat requests
type ChatHandler struct {
	chatService        ChatServiceInterface
	identificationRepo IdentificationRepositoryInterface
	chatRepo           ChatRepositoryInterface
}

// NewChatHandler creates a new chat handler
func NewChatHandler(
	chatService ChatServiceInterface,
	identificationRepo IdentificationRepositoryInterface,
	chatRepo ChatRepositoryInterface,
) *ChatHandler {
	return &ChatHandler{
		chatService:        chatService,
		identificationRepo: identificationRepo,
		chatRepo:           chatRepo,
	}
}

// Handle processes chat requests
func (h *ChatHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse request body
	var req models.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.IdentificationID == "" {
		h.sendError(w, http.StatusBadRequest, "identification_id is required")
		return
	}

	if req.Message == "" {
		h.sendError(w, http.StatusBadRequest, "message is required")
		return
	}

	// Get identification from database
	identification, err := h.identificationRepo.GetByID(req.IdentificationID)
	if err != nil {
		log.Printf("Failed to get identification: %v", err)
		h.sendError(w, http.StatusNotFound, "Identification not found")
		return
	}

	// Get chat history
	chatHistory, err := h.chatRepo.GetByIdentificationID(req.IdentificationID)
	if err != nil {
		log.Printf("Failed to get chat history: %v", err)
		// Continue even if history fetch fails
		chatHistory = []db.ChatMessage{}
	}

	// Save user message to database
	userMessageID := uuid.New().String()
	userMessage := &db.ChatMessage{
		ID:               userMessageID,
		IdentificationID: req.IdentificationID,
		Message:          req.Message,
		Sender:           "user",
		CreatedAt:        time.Now(),
	}

	if err := h.chatRepo.Create(userMessage); err != nil {
		log.Printf("Failed to save user message: %v", err)
		// Continue even if save fails
	}

	// Call chat service
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	chatReq := services.ChatRequest{
		UserMessage:         req.Message,
		Identification:      identification,
		ConversationHistory: chatHistory,
	}

	chatResp, err := h.chatService.Chat(ctx, chatReq)
	if err != nil {
		log.Printf("Chat service error: %v", err)
		h.sendError(w, http.StatusInternalServerError, "Failed to get response from assistant")
		return
	}

	// Save LLM response to database
	llmMessageID := uuid.New().String()
	llmMessage := &db.ChatMessage{
		ID:               llmMessageID,
		IdentificationID: req.IdentificationID,
		Message:          chatResp.Message,
		Sender:           "llm",
		CreatedAt:        time.Now(),
	}

	if err := h.chatRepo.Create(llmMessage); err != nil {
		log.Printf("Failed to save LLM message: %v", err)
		// Continue even if save fails - user still gets response
	}

	// Send response
	response := models.ChatResponse{
		Message:   chatResp.Message,
		MessageID: llmMessageID,
		Timestamp: llmMessage.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// sendError sends an error response
func (h *ChatHandler) sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
}
