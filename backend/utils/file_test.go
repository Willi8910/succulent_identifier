package utils

import (
	"bytes"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"
)

// mockFile implements multipart.File interface for testing
type mockFile struct {
	*bytes.Reader
}

func (m *mockFile) Close() error {
	return nil
}

func newMockFile(content []byte) multipart.File {
	return &mockFile{Reader: bytes.NewReader(content)}
}

func TestNewFileUploader(t *testing.T) {
	tests := []struct {
		name              string
		uploadDir         string
		maxFileSize       int64
		allowedExtensions []string
		wantErr           bool
	}{
		{
			name:              "Valid configuration",
			uploadDir:         "../testdata/uploads",
			maxFileSize:       1024 * 1024,
			allowedExtensions: []string{".jpg", ".png"},
			wantErr:           false,
		},
		{
			name:              "Valid nested directory",
			uploadDir:         "../testdata/nested/uploads",
			maxFileSize:       1024,
			allowedExtensions: []string{".jpg"},
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uploader, err := NewFileUploader(tt.uploadDir, tt.maxFileSize, tt.allowedExtensions)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewFileUploader() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewFileUploader() unexpected error: %v", err)
				return
			}

			if uploader == nil {
				t.Errorf("NewFileUploader() returned nil uploader")
			}

			// Cleanup
			if tt.uploadDir != "" {
				os.RemoveAll(tt.uploadDir)
			}
		})
	}
}

func TestValidateFile(t *testing.T) {
	uploader, _ := NewFileUploader("../testdata/uploads", 1024*1024, []string{".jpg", ".jpeg", ".png"})
	defer os.RemoveAll("../testdata/uploads")

	tests := []struct {
		name       string
		filename   string
		fileSize   int64
		wantErr    bool
		errContains string
	}{
		{
			name:     "Valid JPG file",
			filename: "test.jpg",
			fileSize: 1024,
			wantErr:  false,
		},
		{
			name:     "Valid PNG file",
			filename: "test.png",
			fileSize: 1024,
			wantErr:  false,
		},
		{
			name:     "Valid JPEG file",
			filename: "test.jpeg",
			fileSize: 1024,
			wantErr:  false,
		},
		{
			name:        "File too large",
			filename:    "test.jpg",
			fileSize:    2 * 1024 * 1024,
			wantErr:     true,
			errContains: "exceeds maximum",
		},
		{
			name:        "Invalid extension",
			filename:    "test.pdf",
			fileSize:    1024,
			wantErr:     true,
			errContains: "not allowed",
		},
		{
			name:        "Empty file",
			filename:    "test.jpg",
			fileSize:    0,
			wantErr:     true,
			errContains: "empty",
		},
		{
			name:     "Case insensitive extension",
			filename: "test.JPG",
			fileSize: 1024,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock file header
			fileHeader := &multipart.FileHeader{
				Filename: tt.filename,
				Size:     tt.fileSize,
			}

			err := uploader.ValidateFile(fileHeader)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateFile() expected error, got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateFile() error = %v, should contain %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateFile() unexpected error: %v", err)
			}
		})
	}
}

func TestSaveFile(t *testing.T) {
	uploadDir := "../testdata/uploads_test"
	uploader, _ := NewFileUploader(uploadDir, 1024*1024, []string{".jpg", ".png"})
	defer os.RemoveAll(uploadDir)

	tests := []struct {
		name     string
		filename string
		content  []byte
		wantErr  bool
	}{
		{
			name:     "Save valid file",
			filename: "test.jpg",
			content:  []byte("fake image content"),
			wantErr:  false,
		},
		{
			name:     "Save PNG file",
			filename: "test.png",
			content:  []byte("fake png content"),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock multipart file
			mockFile := newMockFile(tt.content)
			fileHeader := &multipart.FileHeader{
				Filename: tt.filename,
				Size:     int64(len(tt.content)),
			}

			// Save the file
			savedPath, err := uploader.SaveFile(mockFile, fileHeader)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SaveFile() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("SaveFile() unexpected error: %v", err)
				return
			}

			// Verify file exists
			if _, err := os.Stat(savedPath); os.IsNotExist(err) {
				t.Errorf("SaveFile() file not created at %v", savedPath)
			}

			// Verify file has correct extension
			ext := filepath.Ext(savedPath)
			expectedExt := filepath.Ext(tt.filename)
			if ext != expectedExt {
				t.Errorf("SaveFile() extension = %v, expected %v", ext, expectedExt)
			}

			// Verify file content
			content, _ := os.ReadFile(savedPath)
			if !bytes.Equal(content, tt.content) {
				t.Errorf("SaveFile() content mismatch")
			}

			// Cleanup
			os.Remove(savedPath)
		})
	}
}

func TestDeleteFile(t *testing.T) {
	uploadDir := "../testdata/uploads_delete"
	uploader, _ := NewFileUploader(uploadDir, 1024*1024, []string{".jpg"})
	defer os.RemoveAll(uploadDir)

	// Create a test file
	testFile := filepath.Join(uploadDir, "test.jpg")
	os.WriteFile(testFile, []byte("test"), 0644)

	tests := []struct {
		name     string
		filePath string
		wantErr  bool
	}{
		{
			name:     "Delete existing file",
			filePath: testFile,
			wantErr:  false,
		},
		{
			name:     "Delete non-existent file",
			filePath: filepath.Join(uploadDir, "nonexistent.jpg"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := uploader.DeleteFile(tt.filePath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DeleteFile() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("DeleteFile() unexpected error: %v", err)
			}

			// Verify file is deleted
			if _, err := os.Stat(tt.filePath); !os.IsNotExist(err) {
				t.Errorf("DeleteFile() file still exists at %v", tt.filePath)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
