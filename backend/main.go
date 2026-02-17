package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/joho/godotenv"
	"succulent-identifier-backend/db"
	"succulent-identifier-backend/handlers"
	"succulent-identifier-backend/services"
	"succulent-identifier-backend/utils"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it")
	}

	// Load configuration
	config := utils.LoadConfig()

	log.Println("Starting Succulent Identifier Backend API...")
	log.Printf("Server Port: %s", config.ServerPort)
	log.Printf("ML Service URL: %s", config.MLServiceURL)
	log.Printf("Upload Directory: %s", config.UploadDir)
	log.Printf("Species Threshold: %.2f", config.SpeciesThreshold)

	// Initialize database connection
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("Database connected successfully")

	// Run database migrations
	if err := db.RunMigrations(db.DB); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations completed")

	// Initialize repositories
	identificationRepo := db.NewIdentificationRepository(db.DB)
	chatRepo := db.NewChatRepository(db.DB)
	log.Println("Repositories initialized")

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

	// Initialize chat service
	var chatService *services.ChatService
	if config.OpenAIAPIKey != "" {
		chatService = services.NewChatService(config.OpenAIAPIKey)
		log.Println("Chat service initialized with OpenAI")
	} else {
		log.Println("Warning: OPENAI_API_KEY not set, chat feature will be disabled")
	}

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
		identificationRepo,
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

	// Chat endpoint
	if chatService != nil {
		chatHandler := handlers.NewChatHandler(chatService, identificationRepo, chatRepo)
		mux.HandleFunc("/chat", chatHandler.Handle)
		log.Println("Chat endpoint registered")
	}

	// History endpoints
	historyHandler := handlers.NewHistoryHandler(identificationRepo, chatRepo)
	historyRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		// Route based on path
		path := r.URL.Path
		if path == "/history" || path == "/history/" {
			historyHandler.HandleList(w, r)
		} else if strings.HasSuffix(path, "/with-chat") {
			historyHandler.HandleGetWithChat(w, r)
		} else {
			historyHandler.HandleGetByID(w, r)
		}
	}
	// Register both /history and /history/ patterns to handle all history routes
	mux.HandleFunc("/history", historyRouteHandler)
	mux.HandleFunc("/history/", historyRouteHandler)
	mux.HandleFunc("/chat/", historyHandler.HandleGetChatHistory)
	log.Println("History endpoints registered")

	// Serve uploaded images as static files
	fileServer := http.FileServer(http.Dir(config.UploadDir))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", fileServer))
	log.Println("Static file server registered for uploads")

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
