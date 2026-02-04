package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"succulent-identifier-backend/models"
	"time"
)

// MLClient handles communication with the ML inference service
type MLClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewMLClient creates a new ML service client
func NewMLClient(baseURL string) *MLClient {
	return &MLClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Infer sends an image to the ML service for inference
func (c *MLClient) Infer(imagePath string) (*models.MLInferenceResponse, error) {
	// Prepare request
	reqBody := models.MLInferenceRequest{
		ImagePath: imagePath,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request to ML service
	url := fmt.Sprintf("%s/infer", c.baseURL)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call ML service: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML service returned error: status %d", resp.StatusCode)
	}

	// Parse response
	var mlResponse models.MLInferenceResponse
	if err := json.NewDecoder(resp.Body).Decode(&mlResponse); err != nil {
		return nil, fmt.Errorf("failed to decode ML response: %w", err)
	}

	if len(mlResponse.Predictions) == 0 {
		return nil, fmt.Errorf("ML service returned no predictions")
	}

	return &mlResponse, nil
}

// HealthCheck checks if the ML service is available
func (c *MLClient) HealthCheck() error {
	url := fmt.Sprintf("%s/health", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("ML service not reachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ML service unhealthy: status %d", resp.StatusCode)
	}

	return nil
}
