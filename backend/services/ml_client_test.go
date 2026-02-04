package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"succulent-identifier-backend/models"
	"testing"
)

func TestNewMLClient(t *testing.T) {
	baseURL := "http://localhost:8000"
	client := NewMLClient(baseURL)

	if client == nil {
		t.Fatal("NewMLClient() returned nil")
	}

	if client.baseURL != baseURL {
		t.Errorf("NewMLClient() baseURL = %v, expected %v", client.baseURL, baseURL)
	}

	if client.httpClient == nil {
		t.Error("NewMLClient() httpClient is nil")
	}
}

func TestInfer(t *testing.T) {
	tests := []struct {
		name           string
		imagePath      string
		serverResponse models.MLInferenceResponse
		serverStatus   int
		wantErr        bool
		expectedLabel  string
	}{
		{
			name:      "Successful inference",
			imagePath: "/test/path/image.jpg",
			serverResponse: models.MLInferenceResponse{
				Predictions: []models.MLPrediction{
					{Label: "echeveria_elegans", Confidence: 0.85},
					{Label: "echeveria_perle", Confidence: 0.10},
				},
			},
			serverStatus:  http.StatusOK,
			wantErr:       false,
			expectedLabel: "echeveria_elegans",
		},
		{
			name:      "ML service returns error",
			imagePath: "/test/path/image.jpg",
			serverResponse: models.MLInferenceResponse{
				Predictions: []models.MLPrediction{},
			},
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
		{
			name:      "Empty predictions",
			imagePath: "/test/path/image.jpg",
			serverResponse: models.MLInferenceResponse{
				Predictions: []models.MLPrediction{},
			},
			serverStatus: http.StatusOK,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %v", r.Method)
				}
				if r.URL.Path != "/infer" {
					t.Errorf("Expected /infer path, got %v", r.URL.Path)
				}

				// Verify request body
				var req models.MLInferenceRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Errorf("Failed to decode request body: %v", err)
				}

				if req.ImagePath != tt.imagePath {
					t.Errorf("Request image_path = %v, expected %v", req.ImagePath, tt.imagePath)
				}

				// Send response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			// Create client with mock server URL
			client := NewMLClient(server.URL)

			// Call Infer
			response, err := client.Infer(tt.imagePath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Infer() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Infer() unexpected error: %v", err)
				return
			}

			if response == nil {
				t.Fatal("Infer() returned nil response")
			}

			if len(response.Predictions) == 0 {
				t.Fatal("Infer() returned empty predictions")
			}

			if response.Predictions[0].Label != tt.expectedLabel {
				t.Errorf("Infer() label = %v, expected %v",
					response.Predictions[0].Label, tt.expectedLabel)
			}
		})
	}
}

func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name         string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "Healthy service",
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "Unhealthy service",
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
		{
			name:         "Service unavailable",
			serverStatus: http.StatusServiceUnavailable,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/health" {
					t.Errorf("Expected /health path, got %v", r.URL.Path)
				}
				w.WriteHeader(tt.serverStatus)
			}))
			defer server.Close()

			// Create client with mock server URL
			client := NewMLClient(server.URL)

			// Call HealthCheck
			err := client.HealthCheck()

			if tt.wantErr {
				if err == nil {
					t.Errorf("HealthCheck() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("HealthCheck() unexpected error: %v", err)
			}
		})
	}
}

func TestInferServerDown(t *testing.T) {
	// Create client with invalid URL
	client := NewMLClient("http://localhost:99999")

	// Call Infer - should fail to connect
	_, err := client.Infer("/test/image.jpg")

	if err == nil {
		t.Error("Infer() expected error when server is down, got nil")
	}
}
