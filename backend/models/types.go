package models

import "time"

// MLPrediction represents a single prediction from the ML service
type MLPrediction struct {
	Label      string  `json:"label"`
	Confidence float64 `json:"confidence"`
}

// MLInferenceRequest represents the request to ML service
type MLInferenceRequest struct {
	ImagePath string `json:"image_path"`
}

// MLInferenceResponse represents the response from ML service
type MLInferenceResponse struct {
	Predictions []MLPrediction `json:"predictions"`
}

// CareInstructions represents plant care information
type CareInstructions struct {
	Sunlight string `json:"sunlight"`
	Watering string `json:"watering"`
	Soil     string `json:"soil"`
	Notes    string `json:"notes"`
	Trivia   string `json:"trivia,omitempty"`
}

// PlantInfo represents identified plant information
type PlantInfo struct {
	Genus      string  `json:"genus"`
	Species    string  `json:"species,omitempty"`
	Confidence float64 `json:"confidence"`
}

// IdentifyResponse represents the response to the client
type IdentifyResponse struct {
	ID    string           `json:"id"`
	Plant PlantInfo        `json:"plant"`
	Care  CareInstructions `json:"care"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// ChatRequest represents a chat request from the client
type ChatRequest struct {
	IdentificationID string `json:"identification_id"`
	Message          string `json:"message"`
}

// ChatResponse represents a chat response to the client
type ChatResponse struct {
	Message   string    `json:"message"`
	MessageID string    `json:"message_id"`
	Timestamp time.Time `json:"timestamp"`
}

// HistoryItem represents a single identification in the history list
type HistoryItem struct {
	ID         string    `json:"id"`
	Genus      string    `json:"genus"`
	Species    string    `json:"species,omitempty"`
	Confidence float64   `json:"confidence"`
	ImagePath  string    `json:"image_path"`
	CreatedAt  time.Time `json:"created_at"`
}

// HistoryListResponse represents the paginated history list response
type HistoryListResponse struct {
	Items  []HistoryItem `json:"items"`
	Total  int           `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

// HistoryDetailResponse represents detailed information about an identification
type HistoryDetailResponse struct {
	ID         string              `json:"id"`
	Genus      string              `json:"genus"`
	Species    string              `json:"species,omitempty"`
	Confidence float64             `json:"confidence"`
	ImagePath  string              `json:"image_path"`
	CareGuide  *CareInstructions   `json:"care_guide,omitempty"`
	CreatedAt  time.Time           `json:"created_at"`
}

// ChatMessageResponse represents a single chat message
type ChatMessageResponse struct {
	ID        string    `json:"id"`
	Message   string    `json:"message"`
	Sender    string    `json:"sender"` // "user" or "llm"
	CreatedAt time.Time `json:"created_at"`
}

// ChatHistoryResponse represents the chat history for an identification
type ChatHistoryResponse struct {
	IdentificationID string                `json:"identification_id"`
	Messages         []ChatMessageResponse `json:"messages"`
	Total            int                   `json:"total"`
}

// HistoryWithChatResponse represents identification with its chat history
type HistoryWithChatResponse struct {
	Identification HistoryDetailResponse `json:"identification"`
	ChatMessages   []ChatMessageResponse `json:"chat_messages"`
}
