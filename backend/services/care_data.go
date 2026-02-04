package services

import (
	"encoding/json"
	"fmt"
	"os"
	"succulent-identifier-backend/models"
)

// CareDataService manages plant care information
type CareDataService struct {
	careData map[string]models.CareInstructions
}

// NewCareDataService creates a new care data service
func NewCareDataService(careDataPath string) (*CareDataService, error) {
	service := &CareDataService{
		careData: make(map[string]models.CareInstructions),
	}

	if err := service.loadCareData(careDataPath); err != nil {
		return nil, fmt.Errorf("failed to load care data: %w", err)
	}

	return service, nil
}

// loadCareData loads care instructions from JSON file
func (s *CareDataService) loadCareData(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read care data file: %w", err)
	}

	if err := json.Unmarshal(data, &s.careData); err != nil {
		return fmt.Errorf("failed to parse care data JSON: %w", err)
	}

	return nil
}

// GetCareInstructions retrieves care instructions for a species or genus
// Tries species-level first, falls back to genus-level
func (s *CareDataService) GetCareInstructions(species, genus string) (models.CareInstructions, error) {
	// Try species-level care first
	if species != "" {
		if care, exists := s.careData[species]; exists {
			return care, nil
		}
	}

	// Fall back to genus-level care
	if genus != "" {
		if care, exists := s.careData[genus]; exists {
			return care, nil
		}
	}

	return models.CareInstructions{}, fmt.Errorf("no care instructions found for species '%s' or genus '%s'", species, genus)
}
