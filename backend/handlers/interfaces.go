package handlers

import (
	"context"
	"mime/multipart"
	"succulent-identifier-backend/db"
	"succulent-identifier-backend/models"
	"succulent-identifier-backend/services"
)

// MLClientInterface defines the interface for ML service client
type MLClientInterface interface {
	Infer(imagePath string) (*models.MLInferenceResponse, error)
	HealthCheck() error
}

// FileUploaderInterface defines the interface for file uploader
type FileUploaderInterface interface {
	ValidateFile(fileHeader *multipart.FileHeader) error
	SaveFile(file multipart.File, fileHeader *multipart.FileHeader) (string, error)
	DeleteFile(filepath string) error
}

// IdentificationRepositoryInterface defines the interface for identification repository
type IdentificationRepositoryInterface interface {
	Create(identification *db.Identification) error
	GetByID(id string) (*db.Identification, error)
	GetAll(limit, offset int) ([]db.Identification, error)
	Count() (int, error)
	Delete(id string) error
}

// ChatRepositoryInterface defines the interface for chat repository
type ChatRepositoryInterface interface {
	Create(message *db.ChatMessage) error
	GetByIdentificationID(identificationID string) ([]db.ChatMessage, error)
	GetLatestMessages(identificationID string, limit int) ([]db.ChatMessage, error)
	CountByIdentificationID(identificationID string) (int, error)
}

// ChatServiceInterface defines the interface for chat service
type ChatServiceInterface interface {
	Chat(ctx context.Context, req services.ChatRequest) (*services.ChatResponse, error)
	GenerateCareInstructions(ctx context.Context, genus, species string) (*db.CareGuide, error)
}

// CareInstructionsRepositoryInterface defines the interface for care instructions repository
type CareInstructionsRepositoryInterface interface {
	GetBySpecies(genus, species string) (*db.CareInstructionsCache, error)
	Create(cache *db.CareInstructionsCache) error
	Update(cache *db.CareInstructionsCache) error
}
