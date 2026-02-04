package handlers

import (
	"mime/multipart"
	"succulent-identifier-backend/models"
)

// MLClientInterface defines the interface for ML service client
type MLClientInterface interface {
	Infer(imagePath string) (*models.MLInferenceResponse, error)
	HealthCheck() error
}

// CareDataServiceInterface defines the interface for care data service
type CareDataServiceInterface interface {
	GetCareInstructions(species, genus string) (models.CareInstructions, error)
}

// FileUploaderInterface defines the interface for file uploader
type FileUploaderInterface interface {
	ValidateFile(fileHeader *multipart.FileHeader) error
	SaveFile(file multipart.File, fileHeader *multipart.FileHeader) (string, error)
	DeleteFile(filepath string) error
}
