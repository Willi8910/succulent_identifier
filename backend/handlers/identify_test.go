package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"succulent-identifier-backend/models"
	"succulent-identifier-backend/services"
	"succulent-identifier-backend/utils"
	"testing"
)

// mockMLClient simulates ML service responses
type mockMLClient struct {
	response *models.MLInferenceResponse
	err      error
}

func (m *mockMLClient) Infer(imagePath string) (*models.MLInferenceResponse, error) {
	return m.response, m.err
}

func (m *mockMLClient) HealthCheck() error {
	return nil
}

// mockCareDataService simulates care data retrieval
type mockCareDataService struct {
	care models.CareInstructions
	err  error
}

func (m *mockCareDataService) GetCareInstructions(species, genus string) (models.CareInstructions, error) {
	return m.care, m.err
}

func TestIdentifyHandlerHandle(t *testing.T) {
	// Setup test environment
	uploadDir := "../testdata/uploads_handler_test"
	os.MkdirAll(uploadDir, 0755)
	defer os.RemoveAll(uploadDir)

	fileUploader, _ := utils.NewFileUploader(uploadDir, 5*1024*1024, []string{".jpg", ".png"})

	tests := []struct {
		name               string
		method             string
		mlResponse         *models.MLInferenceResponse
		mlError            error
		careInstructions   models.CareInstructions
		careError          error
		speciesThreshold   float64
		expectedStatus     int
		expectSpecies      bool
		setupRequest       func() *http.Request
	}{
		{
			name:   "Successful identification with high confidence",
			method: http.MethodPost,
			mlResponse: &models.MLInferenceResponse{
				Predictions: []models.MLPrediction{
					{Label: "haworthia_zebrina", Confidence: 0.85},
				},
			},
			careInstructions: models.CareInstructions{
				Sunlight: "Bright indirect light",
				Watering: "Water when dry",
				Soil:     "Well-draining",
				Notes:    "Easy care",
			},
			speciesThreshold: 0.4,
			expectedStatus:   http.StatusOK,
			expectSpecies:    true,
			setupRequest: func() *http.Request {
				return createMultipartRequest(t, "test.jpg", []byte("fake image"))
			},
		},
		{
			name:   "Low confidence returns genus only",
			method: http.MethodPost,
			mlResponse: &models.MLInferenceResponse{
				Predictions: []models.MLPrediction{
					{Label: "haworthia_zebrina", Confidence: 0.25},
				},
			},
			careInstructions: models.CareInstructions{
				Sunlight: "Bright indirect light",
				Watering: "Water when dry",
				Soil:     "Well-draining",
				Notes:    "Easy care",
			},
			speciesThreshold: 0.4,
			expectedStatus:   http.StatusOK,
			expectSpecies:    false,
			setupRequest: func() *http.Request {
				return createMultipartRequest(t, "test.jpg", []byte("fake image"))
			},
		},
		{
			name:           "Method not allowed",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "/identify", nil)
				return req
			},
		},
		{
			name:           "No image file provided",
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest,
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodPost, "/identify", bytes.NewBuffer([]byte{}))
				req.Header.Set("Content-Type", "multipart/form-data")
				return req
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mlClient := &mockMLClient{
				response: tt.mlResponse,
				err:      tt.mlError,
			}

			careService := &mockCareDataService{
				care: tt.careInstructions,
				err:  tt.careError,
			}

			// Create handler
			handler := NewIdentifyHandler(
				mlClient,
				careService,
				fileUploader,
				tt.speciesThreshold,
			)

			// Create request
			req := tt.setupRequest()

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.Handle(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v, expected %v",
					rr.Code, tt.expectedStatus)
			}

			// For successful requests, verify response body
			if tt.expectedStatus == http.StatusOK {
				var response models.IdentifyResponse
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
					return
				}

				// Check if genus is present
				if response.Plant.Genus == "" {
					t.Error("Response missing genus")
				}

				// Check species based on confidence
				if tt.expectSpecies && response.Plant.Species == "" {
					t.Error("Expected species in response, got empty")
				}

				if !tt.expectSpecies && response.Plant.Species != "" {
					t.Errorf("Expected no species in response, got %v", response.Plant.Species)
				}

				// Check care instructions
				if response.Care.Sunlight == "" {
					t.Error("Response missing care sunlight")
				}
			}
		})
	}
}

// createMultipartRequest creates a multipart form request with an image file
func createMultipartRequest(t *testing.T, filename string, content []byte) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("image", filename)
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	_, err = part.Write(content)
	if err != nil {
		t.Fatalf("Failed to write content: %v", err)
	}

	writer.Close()

	req, err := http.NewRequest(http.MethodPost, "/identify", body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}

func TestProcessMLResponse(t *testing.T) {
	// Setup care data service with test data
	careService, err := services.NewCareDataService("../testdata/care_data_test.json")
	if err != nil {
		t.Fatalf("Failed to create care service: %v", err)
	}

	// Setup file uploader (not used in this test but required for handler)
	fileUploader, _ := utils.NewFileUploader("../testdata/uploads", 5*1024*1024, []string{".jpg"})

	// Mock ML client (not used in this test but required for handler)
	mlClient := &mockMLClient{}

	tests := []struct {
		name             string
		mlResponse       *models.MLInferenceResponse
		speciesThreshold float64
		expectSpecies    bool
		expectedGenus    string
	}{
		{
			name: "High confidence shows species",
			mlResponse: &models.MLInferenceResponse{
				Predictions: []models.MLPrediction{
					{Label: "test_genus_species", Confidence: 0.85},
				},
			},
			speciesThreshold: 0.4,
			expectSpecies:    true,
			expectedGenus:    "Test_genus",
		},
		{
			name: "Low confidence shows genus only",
			mlResponse: &models.MLInferenceResponse{
				Predictions: []models.MLPrediction{
					{Label: "test_genus_species", Confidence: 0.25},
				},
			},
			speciesThreshold: 0.4,
			expectSpecies:    false,
			expectedGenus:    "Test_genus",
		},
		{
			name: "Threshold boundary - exactly at threshold",
			mlResponse: &models.MLInferenceResponse{
				Predictions: []models.MLPrediction{
					{Label: "test_genus_species", Confidence: 0.4},
				},
			},
			speciesThreshold: 0.4,
			expectSpecies:    true,
			expectedGenus:    "Test_genus",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewIdentifyHandler(
				mlClient,
				careService,
				fileUploader,
				tt.speciesThreshold,
			)

			response, err := handler.processMLResponse(tt.mlResponse)

			if err != nil {
				t.Errorf("processMLResponse() unexpected error: %v", err)
				return
			}

			if response.Plant.Genus == "" {
				t.Error("processMLResponse() genus is empty")
			}

			if tt.expectSpecies && response.Plant.Species == "" {
				t.Error("processMLResponse() expected species, got empty")
			}

			if !tt.expectSpecies && response.Plant.Species != "" {
				t.Errorf("processMLResponse() expected no species, got %v", response.Plant.Species)
			}

			if response.Plant.Confidence != tt.mlResponse.Predictions[0].Confidence {
				t.Errorf("processMLResponse() confidence = %v, expected %v",
					response.Plant.Confidence, tt.mlResponse.Predictions[0].Confidence)
			}
		})
	}
}
