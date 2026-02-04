package utils

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Save original env vars
	origPort := os.Getenv("SERVER_PORT")
	origMLURL := os.Getenv("ML_SERVICE_URL")
	origMaxSize := os.Getenv("MAX_FILE_SIZE")
	origThreshold := os.Getenv("SPECIES_THRESHOLD")

	// Restore env vars after test
	defer func() {
		os.Setenv("SERVER_PORT", origPort)
		os.Setenv("ML_SERVICE_URL", origMLURL)
		os.Setenv("MAX_FILE_SIZE", origMaxSize)
		os.Setenv("SPECIES_THRESHOLD", origThreshold)
	}()

	tests := []struct {
		name                  string
		envVars               map[string]string
		expectedPort          string
		expectedMLURL         string
		expectedMaxFileSize   int64
		expectedThreshold     float64
	}{
		{
			name:                "Default configuration",
			envVars:             map[string]string{},
			expectedPort:        "8080",
			expectedMLURL:       "http://localhost:8000",
			expectedMaxFileSize: 5242880,
			expectedThreshold:   0.4,
		},
		{
			name: "Custom configuration",
			envVars: map[string]string{
				"SERVER_PORT":        "9000",
				"ML_SERVICE_URL":     "http://ml-service:8000",
				"MAX_FILE_SIZE":      "10485760",
				"SPECIES_THRESHOLD":  "0.5",
			},
			expectedPort:        "9000",
			expectedMLURL:       "http://ml-service:8000",
			expectedMaxFileSize: 10485760,
			expectedThreshold:   0.5,
		},
		{
			name: "Partial custom configuration",
			envVars: map[string]string{
				"SERVER_PORT": "3000",
			},
			expectedPort:        "3000",
			expectedMLURL:       "http://localhost:8000",
			expectedMaxFileSize: 5242880,
			expectedThreshold:   0.4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars
			os.Unsetenv("SERVER_PORT")
			os.Unsetenv("ML_SERVICE_URL")
			os.Unsetenv("MAX_FILE_SIZE")
			os.Unsetenv("SPECIES_THRESHOLD")

			// Set test env vars
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Load config
			config := LoadConfig()

			// Verify values
			if config.ServerPort != tt.expectedPort {
				t.Errorf("LoadConfig() ServerPort = %v, expected %v",
					config.ServerPort, tt.expectedPort)
			}

			if config.MLServiceURL != tt.expectedMLURL {
				t.Errorf("LoadConfig() MLServiceURL = %v, expected %v",
					config.MLServiceURL, tt.expectedMLURL)
			}

			if config.MaxFileSize != tt.expectedMaxFileSize {
				t.Errorf("LoadConfig() MaxFileSize = %v, expected %v",
					config.MaxFileSize, tt.expectedMaxFileSize)
			}

			if config.SpeciesThreshold != tt.expectedThreshold {
				t.Errorf("LoadConfig() SpeciesThreshold = %v, expected %v",
					config.SpeciesThreshold, tt.expectedThreshold)
			}

			// Verify allowed extensions
			if len(config.AllowedExtensions) != 3 {
				t.Errorf("LoadConfig() AllowedExtensions length = %v, expected 3",
					len(config.AllowedExtensions))
			}
		})
	}
}
