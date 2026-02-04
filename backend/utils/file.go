package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// FileUploader handles file upload operations
type FileUploader struct {
	uploadDir         string
	maxFileSize       int64
	allowedExtensions []string
}

// NewFileUploader creates a new file uploader
func NewFileUploader(uploadDir string, maxFileSize int64, allowedExtensions []string) (*FileUploader, error) {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	return &FileUploader{
		uploadDir:         uploadDir,
		maxFileSize:       maxFileSize,
		allowedExtensions: allowedExtensions,
	}, nil
}

// ValidateFile validates the uploaded file
func (fu *FileUploader) ValidateFile(fileHeader *multipart.FileHeader) error {
	// Check file size
	if fileHeader.Size > fu.maxFileSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", fu.maxFileSize)
	}

	if fileHeader.Size == 0 {
		return fmt.Errorf("file is empty")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !fu.isAllowedExtension(ext) {
		return fmt.Errorf("file type '%s' not allowed. Allowed types: %v", ext, fu.allowedExtensions)
	}

	return nil
}

// SaveFile saves an uploaded file and returns the file path
func (fu *FileUploader) SaveFile(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	// Validate file first
	if err := fu.ValidateFile(fileHeader); err != nil {
		return "", err
	}

	// Generate unique filename
	ext := filepath.Ext(fileHeader.Filename)
	filename := uuid.New().String() + ext
	filepath := filepath.Join(fu.uploadDir, filename)

	// Create destination file
	dst, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy uploaded file to destination
	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filepath) // Clean up on error
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return filepath, nil
}

// DeleteFile deletes a file from the upload directory
func (fu *FileUploader) DeleteFile(filepath string) error {
	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// isAllowedExtension checks if the file extension is allowed
func (fu *FileUploader) isAllowedExtension(ext string) bool {
	for _, allowed := range fu.allowedExtensions {
		if ext == allowed {
			return true
		}
	}
	return false
}
