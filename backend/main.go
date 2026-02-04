package main

import (
	"fmt"
	"log"
	"net/http"
	"succulent-identifier-backend/handlers"
	"succulent-identifier-backend/services"
	"succulent-identifier-backend/utils"
)

func main() {
	// Load configuration
	config := utils.LoadConfig()

	log.Println("Starting Succulent Identifier Backend API...")
	log.Printf("Server Port: %s", config.ServerPort)
	log.Printf("ML Service URL: %s", config.MLServiceURL)
	log.Printf("Upload Directory: %s", config.UploadDir)
	log.Printf("Species Threshold: %.2f", config.SpeciesThreshold)

	// Initialize services
	mlClient := services.NewMLClient(config.MLServiceURL)
	log.Println("ML Client initialized")

	// Check ML service health
	if err := mlClient.HealthCheck(); err != nil {
		log.Printf("Warning: ML service health check failed: %v", err)
		log.Println("Service will start anyway, but requests may fail until ML service is available")
	} else {
		log.Println("ML service is healthy")
	}

	// Load care data
	careDataService, err := services.NewCareDataService(config.CareDataPath)
	if err != nil {
		log.Fatalf("Failed to initialize care data service: %v", err)
	}
	log.Println("Care data loaded successfully")

	// Initialize file uploader
	fileUploader, err := utils.NewFileUploader(
		config.UploadDir,
		config.MaxFileSize,
		config.AllowedExtensions,
	)
	if err != nil {
		log.Fatalf("Failed to initialize file uploader: %v", err)
	}
	log.Printf("File uploader initialized (Max size: %d bytes)", config.MaxFileSize)

	// Initialize handlers
	identifyHandler := handlers.NewIdentifyHandler(
		mlClient,
		careDataService,
		fileUploader,
		config.SpeciesThreshold,
	)

	// Setup routes
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"succulent-identifier-backend"}`)
	})

	// Root endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"service":"Succulent Identifier Backend","version":"1.0.0","endpoints":["/identify","/health"]}`)
	})

	// Identify endpoint
	mux.HandleFunc("/identify", identifyHandler.Handle)

	// Apply middleware
	handler := utils.CORSMiddleware(mux)

	// Start server
	addr := fmt.Sprintf(":%s", config.ServerPort)
	log.Printf("Server listening on %s", addr)
	log.Println("Ready to accept requests!")

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
