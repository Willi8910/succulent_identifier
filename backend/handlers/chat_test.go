package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"succulent-identifier-backend/db"
	"succulent-identifier-backend/models"
	"succulent-identifier-backend/services"
	"testing"
	"time"
)

// mockChatService simulates chat service responses
type mockChatService struct {
	response *services.ChatResponse
	err      error
}

func (m *mockChatService) Chat(ctx context.Context, req services.ChatRequest) (*services.ChatResponse, error) {
	return m.response, m.err
}

func TestChatHandlerHandle(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		requestBody        interface{}
		identification     *db.Identification
		identificationErr  error
		chatHistory        []db.ChatMessage
		chatHistoryErr     error
		chatResponse       *services.ChatResponse
		chatErr            error
		expectedStatus     int
		expectUserMessage  bool
		expectLLMMessage   bool
	}{
		{
			name:   "Successful chat with plant context",
			method: http.MethodPost,
			requestBody: models.ChatRequest{
				IdentificationID: "plant-id-1",
				Message:          "How often should I water this plant?",
			},
			identification: &db.Identification{
				ID:         "plant-id-1",
				Genus:      "Haworthia",
				Species:    "zebrina",
				Confidence: 0.95,
				ImagePath:  "/uploads/test.jpg",
				CareGuide: &db.CareGuide{
					Sunlight: "Bright indirect light",
					Watering: "Water when dry",
					Soil:     "Well-draining",
					Notes:    "Easy care",
				},
				CreatedAt: time.Now(),
			},
			chatHistory: []db.ChatMessage{},
			chatResponse: &services.ChatResponse{
				Message: "Based on the care instructions, water this Haworthia zebrina when the soil is dry.",
			},
			expectedStatus:    http.StatusOK,
			expectUserMessage: true,
			expectLLMMessage:  true,
		},
		{
			name:   "Chat with conversation history",
			method: http.MethodPost,
			requestBody: models.ChatRequest{
				IdentificationID: "plant-id-2",
				Message:          "Can it survive in low light?",
			},
			identification: &db.Identification{
				ID:         "plant-id-2",
				Genus:      "Aloe",
				Species:    "vera",
				Confidence: 0.90,
				ImagePath:  "/uploads/aloe.jpg",
				CareGuide: &db.CareGuide{
					Sunlight: "Full sun",
					Watering: "Infrequent",
					Soil:     "Sandy",
				},
				CreatedAt: time.Now(),
			},
			chatHistory: []db.ChatMessage{
				{
					ID:               "msg-1",
					IdentificationID: "plant-id-2",
					Message:          "What is this plant?",
					Sender:           "user",
					CreatedAt:        time.Now().Add(-5 * time.Minute),
				},
				{
					ID:               "msg-2",
					IdentificationID: "plant-id-2",
					Message:          "This is Aloe vera.",
					Sender:           "llm",
					CreatedAt:        time.Now().Add(-4 * time.Minute),
				},
			},
			chatResponse: &services.ChatResponse{
				Message: "Aloe vera prefers full sun but can tolerate some shade.",
			},
			expectedStatus:    http.StatusOK,
			expectUserMessage: true,
			expectLLMMessage:  true,
		},
		{
			name:           "Method not allowed",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "Missing identification_id",
			method: http.MethodPost,
			requestBody: models.ChatRequest{
				Message: "Test message",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Missing message",
			method: http.MethodPost,
			requestBody: models.ChatRequest{
				IdentificationID: "plant-id-1",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Identification not found",
			method: http.MethodPost,
			requestBody: models.ChatRequest{
				IdentificationID: "non-existent",
				Message:          "Test message",
			},
			identificationErr: db.ErrNotFound,
			expectedStatus:    http.StatusNotFound,
		},
		{
			name:   "Chat service error",
			method: http.MethodPost,
			requestBody: models.ChatRequest{
				IdentificationID: "plant-id-1",
				Message:          "Test message",
			},
			identification: &db.Identification{
				ID:    "plant-id-1",
				Genus: "Test",
			},
			chatHistory:       []db.ChatMessage{},
			chatErr:           db.ErrNotFound,
			expectedStatus:    http.StatusInternalServerError,
			expectUserMessage: true,
			expectLLMMessage:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockIdentRepo := &mockIdentificationRepository{
				getByIDResult: tt.identification,
				getByIDErr:    tt.identificationErr,
			}

			mockChatRepo := &mockChatRepository{
				getAllResult: tt.chatHistory,
				getAllErr:    tt.chatHistoryErr,
			}

			mockChatSvc := &mockChatService{
				response: tt.chatResponse,
				err:      tt.chatErr,
			}

			// Create handler
			handler := NewChatHandler(mockChatSvc, mockIdentRepo, mockChatRepo)

			// Create request
			var req *http.Request
			if tt.method == http.MethodPost && tt.requestBody != nil {
				body, _ := json.Marshal(tt.requestBody)
				req = httptest.NewRequest(tt.method, "/chat", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, "/chat", nil)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.Handle(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v, expected %v",
					rr.Code, tt.expectedStatus)
			}

			// For successful requests, verify response
			if tt.expectedStatus == http.StatusOK {
				var response models.ChatResponse
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
					return
				}

				if response.Message == "" {
					t.Error("Response missing message")
				}

				if response.MessageID == "" {
					t.Error("Response missing message_id")
				}

				// Verify messages were saved
				if tt.expectUserMessage && !mockChatRepo.createCalled {
					t.Error("Expected user message to be saved")
				}

				if tt.expectLLMMessage && mockChatRepo.createCallCount < 2 {
					t.Errorf("Expected both user and LLM messages to be saved, got %d calls",
						mockChatRepo.createCallCount)
				}
			}
		})
	}
}

// mockChatRepository simulates chat repository operations
type mockChatRepository struct {
	createCalled     bool
	createCallCount  int
	lastCreated      *db.ChatMessage
	createErr        error
	getAllResult     []db.ChatMessage
	getAllErr        error
	getLatestResult  []db.ChatMessage
	getLatestErr     error
	countResult      int
	countErr         error
}

func (m *mockChatRepository) Create(message *db.ChatMessage) error {
	m.createCalled = true
	m.createCallCount++
	m.lastCreated = message
	return m.createErr
}

func (m *mockChatRepository) GetByIdentificationID(identificationID string) ([]db.ChatMessage, error) {
	return m.getAllResult, m.getAllErr
}

func (m *mockChatRepository) GetLatestMessages(identificationID string, limit int) ([]db.ChatMessage, error) {
	return m.getLatestResult, m.getLatestErr
}

func (m *mockChatRepository) CountByIdentificationID(identificationID string) (int, error) {
	return m.countResult, m.countErr
}

func TestChatHandlerIntegration(t *testing.T) {
	tests := []struct {
		name              string
		identification    *db.Identification
		userMessage       string
		chatResponse      string
		expectBothSaved   bool
	}{
		{
			name: "Full chat flow with plant context",
			identification: &db.Identification{
				ID:         "plant-123",
				Genus:      "Echeveria",
				Species:    "elegans",
				Confidence: 0.98,
				CareGuide: &db.CareGuide{
					Sunlight: "Bright light",
					Watering: "When soil is dry",
					Soil:     "Well-draining cactus mix",
				},
			},
			userMessage:     "Is this plant pet-safe?",
			chatResponse:    "Echeveria elegans is generally non-toxic to pets.",
			expectBothSaved: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockIdentRepo := &mockIdentificationRepository{
				getByIDResult: tt.identification,
			}

			mockChatRepo := &mockChatRepository{
				getAllResult: []db.ChatMessage{},
			}

			mockChatSvc := &mockChatService{
				response: &services.ChatResponse{
					Message: tt.chatResponse,
				},
			}

			// Create handler
			handler := NewChatHandler(mockChatSvc, mockIdentRepo, mockChatRepo)

			// Create request
			reqBody := models.ChatRequest{
				IdentificationID: tt.identification.ID,
				Message:          tt.userMessage,
			}
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.Handle(rr, req)

			// Verify success
			if rr.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %v", rr.Code)
			}

			// Verify both messages saved
			if tt.expectBothSaved {
				if mockChatRepo.createCallCount != 2 {
					t.Errorf("Expected 2 messages saved (user + LLM), got %d",
						mockChatRepo.createCallCount)
				}
			}

			// Verify response
			var response models.ChatResponse
			json.NewDecoder(rr.Body).Decode(&response)

			if response.Message != tt.chatResponse {
				t.Errorf("Expected response '%s', got '%s'",
					tt.chatResponse, response.Message)
			}
		})
	}
}
