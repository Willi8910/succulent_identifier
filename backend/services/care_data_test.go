package services

import (
	"testing"
)

func TestNewCareDataService(t *testing.T) {
	tests := []struct {
		name        string
		careDataPath string
		wantErr     bool
	}{
		{
			name:        "Valid care data file",
			careDataPath: "../testdata/care_data_test.json",
			wantErr:     false,
		},
		{
			name:        "Non-existent file",
			careDataPath: "../testdata/nonexistent.json",
			wantErr:     true,
		},
		{
			name:        "Empty path",
			careDataPath: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewCareDataService(tt.careDataPath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewCareDataService() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewCareDataService() unexpected error: %v", err)
				return
			}

			if service == nil {
				t.Errorf("NewCareDataService() returned nil service")
			}

			if service.careData == nil {
				t.Errorf("NewCareDataService() careData is nil")
			}
		})
	}
}

func TestGetCareInstructions(t *testing.T) {
	// Setup service with test data
	service, err := NewCareDataService("../testdata/care_data_test.json")
	if err != nil {
		t.Fatalf("Failed to create test service: %v", err)
	}

	tests := []struct {
		name        string
		species     string
		genus       string
		wantErr     bool
		expectedSunlight string
	}{
		{
			name:        "Get species-level care",
			species:     "test_genus_species",
			genus:       "test_genus",
			wantErr:     false,
			expectedSunlight: "Test species sunlight",
		},
		{
			name:        "Fallback to genus-level care",
			species:     "test_genus_nonexistent",
			genus:       "test_genus",
			wantErr:     false,
			expectedSunlight: "Test genus sunlight",
		},
		{
			name:        "No care data found",
			species:     "nonexistent_species",
			genus:       "nonexistent_genus",
			wantErr:     true,
			expectedSunlight: "",
		},
		{
			name:        "Empty species, valid genus",
			species:     "",
			genus:       "test_genus",
			wantErr:     false,
			expectedSunlight: "Test genus sunlight",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			care, err := service.GetCareInstructions(tt.species, tt.genus)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetCareInstructions() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetCareInstructions() unexpected error: %v", err)
				return
			}

			if care.Sunlight != tt.expectedSunlight {
				t.Errorf("GetCareInstructions() sunlight = %v, expected %v",
					care.Sunlight, tt.expectedSunlight)
			}
		})
	}
}
