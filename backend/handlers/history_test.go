package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"succulent-identifier-backend/db"
	"succulent-identifier-backend/models"
	"testing"
	"time"
)

func TestHistoryHandlerHandleList(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		identifications []db.Identification
		repoErr        error
		totalCount     int
		countErr       error
		expectedStatus int
		expectedItems  int
	}{
		{
			name:        "Successful list with default pagination",
			queryParams: "",
			identifications: []db.Identification{
				{
					ID:         "id-1",
					Genus:      "Haworthia",
					Species:    "zebrina",
					Confidence: 0.95,
					ImagePath:  "/uploads/1.jpg",
					CreatedAt:  time.Now(),
				},
				{
					ID:         "id-2",
					Genus:      "Aloe",
					Species:    "vera",
					Confidence: 0.90,
					ImagePath:  "/uploads/2.jpg",
					CreatedAt:  time.Now(),
				},
			},
			totalCount:     2,
			expectedStatus: http.StatusOK,
			expectedItems:  2,
		},
		{
			name:           "Empty list",
			queryParams:    "",
			identifications: []db.Identification{},
			totalCount:     0,
			expectedStatus: http.StatusOK,
			expectedItems:  0,
		},
		{
			name:        "Custom pagination",
			queryParams: "?limit=5&offset=10",
			identifications: []db.Identification{
				{
					ID:         "id-11",
					Genus:      "Echeveria",
					Species:    "elegans",
					Confidence: 0.88,
					ImagePath:  "/uploads/11.jpg",
					CreatedAt:  time.Now(),
				},
			},
			totalCount:     50,
			expectedStatus: http.StatusOK,
			expectedItems:  1,
		},
		{
			name:           "Database error",
			queryParams:    "",
			repoErr:        db.ErrNotFound,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "Method not allowed",
			queryParams: "",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockIdentRepo := &mockIdentificationRepository{
				getAllResult: tt.identifications,
				getAllErr:    tt.repoErr,
				countResult:  tt.totalCount,
				countErr:     tt.countErr,
			}

			mockChatRepo := &mockChatRepository{}

			// Create handler
			handler := NewHistoryHandler(mockIdentRepo, mockChatRepo)

			// Create request
			method := http.MethodGet
			if tt.name == "Method not allowed" {
				method = http.MethodPost
			}
			req := httptest.NewRequest(method, "/history"+tt.queryParams, nil)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.HandleList(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v, expected %v",
					rr.Code, tt.expectedStatus)
			}

			// For successful requests, verify response
			if tt.expectedStatus == http.StatusOK {
				var response models.HistoryListResponse
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
					return
				}

				if len(response.Items) != tt.expectedItems {
					t.Errorf("Expected %d items, got %d", tt.expectedItems, len(response.Items))
				}

				if response.Total != tt.totalCount {
					t.Errorf("Expected total %d, got %d", tt.totalCount, response.Total)
				}

				// Verify item structure
				for _, item := range response.Items {
					if item.ID == "" {
						t.Error("Item missing ID")
					}
					if item.Genus == "" {
						t.Error("Item missing genus")
					}
					if item.ImagePath == "" {
						t.Error("Item missing image path")
					}
				}
			}
		})
	}
}

func TestHistoryHandlerHandleGetByID(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		identification *db.Identification
		repoErr        error
		expectedStatus int
	}{
		{
			name: "Successful get by ID",
			path: "/history/plant-id-1",
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
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Not found",
			path:           "/history/non-existent",
			repoErr:        db.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Method not allowed",
			path:           "/history/plant-id-1",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockIdentRepo := &mockIdentificationRepository{
				getByIDResult: tt.identification,
				getByIDErr:    tt.repoErr,
			}

			mockChatRepo := &mockChatRepository{}

			// Create handler
			handler := NewHistoryHandler(mockIdentRepo, mockChatRepo)

			// Create request
			method := http.MethodGet
			if tt.name == "Method not allowed" {
				method = http.MethodPost
			}
			req := httptest.NewRequest(method, tt.path, nil)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.HandleGetByID(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v, expected %v",
					rr.Code, tt.expectedStatus)
			}

			// For successful requests, verify response
			if tt.expectedStatus == http.StatusOK {
				var response models.HistoryDetailResponse
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
					return
				}

				if response.ID != tt.identification.ID {
					t.Errorf("Expected ID %s, got %s", tt.identification.ID, response.ID)
				}

				if response.Genus != tt.identification.Genus {
					t.Errorf("Expected genus %s, got %s", tt.identification.Genus, response.Genus)
				}

				if tt.identification.CareGuide != nil && response.CareGuide == nil {
					t.Error("Expected care guide in response")
				}

				if response.CareGuide != nil {
					if response.CareGuide.Sunlight == "" {
						t.Error("Care guide missing sunlight")
					}
				}
			}
		})
	}
}

func TestHistoryHandlerHandleGetChatHistory(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		chatMessages   []db.ChatMessage
		repoErr        error
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "Successful get chat history",
			path: "/chat/plant-id-1",
			chatMessages: []db.ChatMessage{
				{
					ID:               "msg-1",
					IdentificationID: "plant-id-1",
					Message:          "How often should I water?",
					Sender:           "user",
					CreatedAt:        time.Now().Add(-10 * time.Minute),
				},
				{
					ID:               "msg-2",
					IdentificationID: "plant-id-1",
					Message:          "Water when the soil is dry.",
					Sender:           "llm",
					CreatedAt:        time.Now().Add(-9 * time.Minute),
				},
				{
					ID:               "msg-3",
					IdentificationID: "plant-id-1",
					Message:          "How much sunlight does it need?",
					Sender:           "user",
					CreatedAt:        time.Now().Add(-5 * time.Minute),
				},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "Empty chat history",
			path:           "/chat/plant-id-2",
			chatMessages:   []db.ChatMessage{},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "Database error",
			path:           "/chat/plant-id-3",
			repoErr:        db.ErrNotFound,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Method not allowed",
			path:           "/chat/plant-id-1",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockIdentRepo := &mockIdentificationRepository{}

			mockChatRepo := &mockChatRepository{
				getAllResult: tt.chatMessages,
				getAllErr:    tt.repoErr,
			}

			// Create handler
			handler := NewHistoryHandler(mockIdentRepo, mockChatRepo)

			// Create request
			method := http.MethodGet
			if tt.name == "Method not allowed" {
				method = http.MethodPost
			}
			req := httptest.NewRequest(method, tt.path, nil)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.HandleGetChatHistory(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v, expected %v",
					rr.Code, tt.expectedStatus)
			}

			// For successful requests, verify response
			if tt.expectedStatus == http.StatusOK {
				var response models.ChatHistoryResponse
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
					return
				}

				if len(response.Messages) != tt.expectedCount {
					t.Errorf("Expected %d messages, got %d", tt.expectedCount, len(response.Messages))
				}

				if response.Total != tt.expectedCount {
					t.Errorf("Expected total %d, got %d", tt.expectedCount, response.Total)
				}

				// Verify message structure
				for _, msg := range response.Messages {
					if msg.ID == "" {
						t.Error("Message missing ID")
					}
					if msg.Message == "" {
						t.Error("Message missing content")
					}
					if msg.Sender != "user" && msg.Sender != "llm" {
						t.Errorf("Invalid sender: %s", msg.Sender)
					}
				}
			}
		})
	}
}

func TestHistoryHandlerHandleGetWithChat(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		identification *db.Identification
		chatMessages   []db.ChatMessage
		identErr       error
		chatErr        error
		expectedStatus int
	}{
		{
			name: "Successful get with chat",
			path: "/history/plant-id-1/with-chat",
			identification: &db.Identification{
				ID:         "plant-id-1",
				Genus:      "Echeveria",
				Species:    "elegans",
				Confidence: 0.98,
				ImagePath:  "/uploads/test.jpg",
				CareGuide: &db.CareGuide{
					Sunlight: "Bright light",
					Watering: "When soil is dry",
					Soil:     "Well-draining",
				},
				CreatedAt: time.Now(),
			},
			chatMessages: []db.ChatMessage{
				{
					ID:               "msg-1",
					IdentificationID: "plant-id-1",
					Message:          "Is this plant pet-safe?",
					Sender:           "user",
					CreatedAt:        time.Now().Add(-5 * time.Minute),
				},
				{
					ID:               "msg-2",
					IdentificationID: "plant-id-1",
					Message:          "Yes, Echeveria is generally non-toxic to pets.",
					Sender:           "llm",
					CreatedAt:        time.Now().Add(-4 * time.Minute),
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Get with empty chat",
			path: "/history/plant-id-2/with-chat",
			identification: &db.Identification{
				ID:         "plant-id-2",
				Genus:      "Aloe",
				Species:    "vera",
				Confidence: 0.90,
				ImagePath:  "/uploads/aloe.jpg",
				CreatedAt:  time.Now(),
			},
			chatMessages:   []db.ChatMessage{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Identification not found",
			path:           "/history/non-existent/with-chat",
			identErr:       db.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Chat fetch error (continues with empty)",
			path: "/history/plant-id-3/with-chat",
			identification: &db.Identification{
				ID:    "plant-id-3",
				Genus: "Test",
			},
			chatErr:        db.ErrNotFound,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockIdentRepo := &mockIdentificationRepository{
				getByIDResult: tt.identification,
				getByIDErr:    tt.identErr,
			}

			mockChatRepo := &mockChatRepository{
				getAllResult: tt.chatMessages,
				getAllErr:    tt.chatErr,
			}

			// Create handler
			handler := NewHistoryHandler(mockIdentRepo, mockChatRepo)

			// Create request
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.HandleGetWithChat(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v, expected %v",
					rr.Code, tt.expectedStatus)
			}

			// For successful requests, verify response
			if tt.expectedStatus == http.StatusOK {
				var response models.HistoryWithChatResponse
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
					return
				}

				if response.Identification.ID != tt.identification.ID {
					t.Errorf("Expected identification ID %s, got %s",
						tt.identification.ID, response.Identification.ID)
				}

				expectedMsgCount := len(tt.chatMessages)
				if tt.chatErr != nil {
					expectedMsgCount = 0 // Empty on error
				}

				if len(response.ChatMessages) != expectedMsgCount {
					t.Errorf("Expected %d chat messages, got %d",
						expectedMsgCount, len(response.ChatMessages))
				}
			}
		})
	}
}
