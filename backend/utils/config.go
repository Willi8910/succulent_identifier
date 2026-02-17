package utils

import (
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	// Server configuration
	ServerPort string

	// ML Service configuration
	MLServiceURL string

	// File upload configuration
	UploadDir         string
	MaxFileSize       int64 // in bytes
	AllowedExtensions []string

	// Confidence threshold
	SpeciesThreshold float64

	// Care data path
	CareDataPath string

	// OpenAI configuration
	OpenAIAPIKey string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	maxFileSize, _ := strconv.ParseInt(getEnv("MAX_FILE_SIZE", "5242880"), 10, 64) // Default 5MB
	speciesThreshold, _ := strconv.ParseFloat(getEnv("SPECIES_THRESHOLD", "0.4"), 64)

	return &Config{
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		MLServiceURL:      getEnv("ML_SERVICE_URL", "http://localhost:8000"),
		UploadDir:         getEnv("UPLOAD_DIR", "./uploads"),
		MaxFileSize:       maxFileSize,
		AllowedExtensions: []string{".jpg", ".jpeg", ".png"},
		SpeciesThreshold:  speciesThreshold,
		CareDataPath:      getEnv("CARE_DATA_PATH", "../care_data.json"),
		OpenAIAPIKey:      getEnv("OPENAI_API_KEY", ""),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
