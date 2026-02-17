# 🌵 Succulent Identifier

A full-stack web application that uses deep learning to identify succulent plants from photos, provides personalized care instructions, and offers an AI-powered chat assistant to answer your plant care questions. Simply upload a photo of your succulent, get instant identification with confidence scores, view your identification history, and chat with an AI expert about your plant's care needs.

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
- [Project Structure](#project-structure)
- [API Documentation](#api-documentation)
- [Model Performance](#model-performance)
- [Contributing](#contributing)
- [License](#license)

## 🌟 Overview

Succulent Identifier is a machine learning-powered application designed to help plant enthusiasts identify their succulent plants and learn how to care for them properly. The system uses a fine-tuned EfficientNet-B0 model trained on thousands of succulent images to provide accurate species identification with confidence scores. Built with PostgreSQL for data persistence and integrated with OpenAI's GPT-4o-mini for intelligent chat assistance, the application offers a complete plant identification and care management experience.

### What It Can Do

- **Identify Succulent Species**: Upload a photo to get instant identification
- **Confidence Scoring**: Visual confidence bar showing prediction certainty
- **Smart Fallback**: Shows genus-level information when species confidence is low
- **Care Instructions**: Get detailed care guidance including:
  - Sunlight requirements
  - Watering schedule
  - Soil recommendations
  - Additional care notes
- **AI Chat Assistant**: Ask questions about your identified plant using GPT-4o-mini
  - Context-aware responses based on identification
  - Persistent chat history
  - Natural language conversation
- **Identification History**: View and access all your past identifications
  - Left sidebar with chronological list
  - Click to reload historical identifications
  - Auto-refresh on new uploads
  - Display historical images
- **Responsive Design**: Works seamlessly on desktop, tablet, and mobile devices
- **Real-time Processing**: Fast inference with results in under 2 seconds
- **Persistent Storage**: PostgreSQL database for identification and chat history

### Current Species Support

The model is currently trained on 3 succulent species:
- **Cryptanthus bivittatus** (Earth Star)
- **Haworthia zebrina** (Zebra Haworthia)
- **Opuntia microdasys** (Bunny Ear Cactus)

The architecture supports easy expansion to additional species by:
1. Scraping more images with the included scraper tool
2. Retraining the model
3. Adding care data to the database

## 🛠 Tech Stack

### Frontend
- **React 18** - Modern UI library with hooks
- **Axios** - HTTP client for API communication
- **CSS3** - Custom styling with animations and gradients
- **Create React App** - Development tooling

### Backend API
- **Go 1.21+** - High-performance REST API
- **PostgreSQL** - Relational database for persistence
- **GORM** - ORM for database operations
- **OpenAI Go SDK** - GPT-4o-mini chat integration
- **godotenv** - Environment variable management
- **UUID** - Secure file handling
- **CORS Middleware** - Cross-origin support

### ML Service
- **Python 3.11** - ML runtime environment
- **PyTorch 2.5+** - Deep learning framework
- **EfficientNet-B0** - Transfer learning model
- **FastAPI** - Modern async API framework
- **Pillow** - Image processing
- **Uvicorn** - ASGI server

### Data Collection
- **Selenium** - Web scraping automation
- **Chrome WebDriver** - Browser automation
- **Bing Images** - Image data source

### Testing
- **Go testing** - Backend unit tests (70-90% coverage)
- **httptest** - HTTP handler testing
- **Mocks** - Dependency injection for testability

## 🏗 Architecture

### System Architecture Diagram

```mermaid
graph TB
    subgraph "Frontend Layer"
        UI[React Web App<br/>Port 3000]
        ChatDrawer[Chat Drawer Component]
        HistorySidebar[History Sidebar]
    end

    subgraph "Backend API Layer"
        API[Go REST API<br/>Port 8080]
        FileHandler[File Upload Handler]
        CareService[Care Data Service]
        ChatService[OpenAI Chat Service]
        HistoryHandler[History Handler]
    end

    subgraph "ML Service Layer"
        ML[FastAPI ML Service<br/>Port 8000]
        Model[EfficientNet-B0 Model]
        Preprocessor[Image Preprocessor]
    end

    subgraph "External Services"
        OpenAI[OpenAI API<br/>GPT-4o-mini]
    end

    subgraph "Data Layer"
        DB[(PostgreSQL Database)]
        Images[(Uploaded Images)]
        CareData[(care_data.json)]
        ModelFile[(trained_model.pth)]
    end

    UI -->|POST /identify| API
    UI -->|POST /chat| API
    UI -->|GET /history| API
    ChatDrawer -->|POST /chat| API
    HistorySidebar -->|GET /history| API

    API -->|Save image| FileHandler
    FileHandler -->|Store| Images
    API -->|Get care instructions| CareService
    CareService -->|Read| CareData
    API -->|POST /infer| ML
    API -->|Save identification| DB
    API -->|Save chat messages| DB
    API -->|Query history| HistoryHandler
    HistoryHandler -->|Read| DB

    ChatService -->|Call GPT-4o-mini| OpenAI
    API -->|Use| ChatService

    ML -->|Load| ModelFile
    ML -->|Process| Preprocessor
    Preprocessor -->|Read| Images
    Model -->|Predict| ML
    ML -->|Return predictions| API
    API -->|Return results| UI

    style UI fill:#48bb78,stroke:#38a169,stroke-width:3px,color:#fff
    style API fill:#4299e1,stroke:#3182ce,stroke-width:3px,color:#fff
    style ML fill:#ed8936,stroke:#dd6b20,stroke-width:3px,color:#fff
    style DB fill:#9f7aea,stroke:#805ad5,stroke-width:3px,color:#fff
    style OpenAI fill:#f56565,stroke:#e53e3e,stroke-width:3px,color:#fff
```

### Request Flow

```mermaid
sequenceDiagram
    participant User
    participant Frontend
    participant Backend
    participant MLService
    participant Model

    User->>Frontend: Upload succulent image
    Frontend->>Backend: POST /identify (multipart/form-data)
    Backend->>Backend: Validate file (type, size)
    Backend->>Backend: Save with UUID filename
    Backend->>MLService: POST /infer (absolute path)
    MLService->>MLService: Preprocess image (resize, normalize)
    MLService->>Model: Forward pass
    Model->>MLService: Predictions with confidence
    MLService->>Backend: Top-K predictions
    Backend->>Backend: Apply 0.4 threshold logic
    alt Confidence >= 0.4
        Backend->>Backend: Get species-level care data
    else Confidence < 0.4
        Backend->>Backend: Get genus-level care data
    end
    Backend->>Frontend: Return plant + care info
    Frontend->>User: Display results + care instructions
```

### Data Flow

```mermaid
graph LR
    subgraph "Data Collection"
        Scraper[Bing Image Scraper]
        RawImages[Raw Images<br/>~300 per species]
    end

    subgraph "Training Pipeline"
        DataAug[Data Augmentation]
        Train[Model Training]
        Validation[Validation]
        ModelCheckpoint[Best Model Checkpoint]
    end

    subgraph "Production"
        Inference[Inference Service]
        UserImage[User Uploaded Image]
        Prediction[Species Prediction]
    end

    Scraper -->|Download| RawImages
    RawImages -->|80/20 split| DataAug
    DataAug -->|Augmented batches| Train
    Train -->|Evaluate| Validation
    Validation -->|Save best| ModelCheckpoint
    ModelCheckpoint -->|Load| Inference
    UserImage -->|Process| Inference
    Inference -->|Return| Prediction

    style Scraper fill:#fbd38d,stroke:#ed8936,stroke-width:2px
    style Train fill:#f6ad55,stroke:#ed8936,stroke-width:2px
    style Inference fill:#ed8936,stroke:#dd6b20,stroke-width:3px,color:#fff
```

## 🚀 Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:

- **Node.js** >= 14.0.0 and npm >= 6.0.0
- **Python** >= 3.11
- **Go** >= 1.21
- **PostgreSQL** >= 14
- **Chrome** browser (for scraper)
- **Git**
- **OpenAI API Key** (for chat feature)

### Installation

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd succulent_identifier
   ```

2. **Install Python dependencies** (ML Service):
   ```bash
   cd ml_service
   pip install -r requirements.txt
   cd ..
   ```

3. **Install Go dependencies** (Backend):
   ```bash
   cd backend
   go mod download
   cd ..
   ```

4. **Install Node.js dependencies** (Frontend):
   ```bash
   cd frontend
   npm install
   cd ..
   ```

5. **Set up PostgreSQL database**:
   ```bash
   # Create database
   createdb succulent_identifier

   # The backend will auto-migrate tables on startup
   ```

6. **Configure environment variables** (Backend):
   ```bash
   cd backend
   cp .env.example .env
   # Edit .env and add:
   # - DATABASE_URL (PostgreSQL connection string)
   # - OPENAI_API_KEY (your OpenAI API key)
   cd ..
   ```

### Starting the Services

The application consists of three services that need to be running simultaneously. Start them in separate terminal windows:

#### 1. Start ML Service (Port 8000)

```bash
# Terminal 1
cd ml_service/src
python inference.py
```

Expected output:
```
INFO:     Started server process
INFO:     Waiting for application startup.
Loading model from: ../models/succulent_classifier_best.pth
Model loaded successfully.
INFO:     Application startup complete.
INFO:     Uvicorn running on http://0.0.0.0:8000
```

**Health check**: Open http://localhost:8000 in your browser. You should see:
```json
{"status": "healthy", "model_loaded": true}
```

#### 2. Start Backend API (Port 8080)

```bash
# Terminal 2
cd backend
go run main.go
```

Expected output:
```
2026/02/17 22:20:00 Starting Succulent Identifier Backend API...
2026/02/17 22:20:00 Successfully connected to PostgreSQL database
2026/02/17 22:20:00 Database migrations completed successfully
2026/02/17 22:20:00 ML service is healthy
2026/02/17 22:20:00 Care data loaded successfully
2026/02/17 22:20:00 Chat service initialized with OpenAI
2026/02/17 22:20:00 History endpoints registered
2026/02/17 22:20:00 Server listening on :8080
```

**Health check**: Test the API:
```bash
curl http://localhost:8080/health
```

Expected response:
```json
{"status": "healthy"}
```

#### 3. Start Frontend (Port 3000)

```bash
# Terminal 3
cd frontend
npm start
```

Expected output:
```
Compiled successfully!

You can now view frontend in the browser.

  Local:            http://localhost:3000
  On Your Network:  http://192.168.1.x:3000
```

The browser will automatically open to http://localhost:3000.

### Verify All Services

Run this command to check all services are running:

```bash
# Check if all ports are active
lsof -ti:3000,8000,8080 | wc -l
```

Expected output: `3` (or more)

### Testing the Application

1. **Open the app**: Navigate to http://localhost:3000
2. **Upload an image**:
   - Drag and drop a succulent image, or
   - Click the upload area to browse
3. **View results**:
   - Species identification
   - Confidence score
   - Care instructions

**Test images** are available in:
```
ml_service/data/raw/cryptanthus-cryptanthus_bivittatus/
ml_service/data/raw/haworthia-haworthia_zebrina/
ml_service/data/raw/opuntia-opuntia_microdasys/
```

### Stopping the Services

To stop all services:

1. Press `Ctrl+C` in each terminal window
2. Or run:
   ```bash
   # Kill all services at once
   lsof -ti:3000,8000,8080 | xargs kill -9
   ```

## 📁 Project Structure

```
succulent_identifier/
├── README.md                    # This file
├── swagger.yaml                 # OpenAPI/Swagger API specification
├── PRD.txt                      # Product Requirements Document
├── TDD.txt                      # Technical Design Document
├── TODO.md                      # Project progress tracker
├── care_data.json              # Plant care instructions database
│
├── ml_service/                 # ML Service (Python + PyTorch + FastAPI)
│   ├── data/
│   │   └── raw/               # Training images (870 images)
│   ├── models/                # Trained model files
│   │   ├── succulent_classifier_best.pth
│   │   ├── succulent_classifier_final.pth
│   │   └── training_history.png
│   ├── src/
│   │   ├── train.py           # Model training script
│   │   ├── inference.py       # FastAPI inference service
│   │   └── preprocessing.py   # Image preprocessing utilities
│   ├── scraper.py             # Bing image scraper
│   ├── labels.json            # Species label mappings
│   ├── requirements.txt       # Python dependencies
│   ├── Dockerfile             # ML service container
│   └── README.md              # ML service documentation
│
├── backend/                    # Backend API (Golang)
│   ├── db/                    # Database layer
│   │   ├── db.go              # Database connection
│   │   ├── models.go          # Database models
│   │   ├── identification_repository.go
│   │   ├── identification_repository_test.go
│   │   ├── chat_repository.go
│   │   └── chat_repository_test.go
│   ├── handlers/              # HTTP request handlers
│   │   ├── identify.go
│   │   ├── identify_test.go
│   │   ├── chat.go            # NEW: Chat endpoint
│   │   ├── chat_test.go       # NEW: Chat tests
│   │   ├── history.go         # NEW: History endpoints
│   │   ├── history_test.go    # NEW: History tests
│   │   └── interfaces.go
│   ├── models/                # Data structures
│   │   └── types.go
│   ├── services/              # Business logic
│   │   ├── care_data.go
│   │   ├── care_data_test.go
│   │   ├── ml_client.go
│   │   ├── ml_client_test.go
│   │   └── chat_service.go    # NEW: OpenAI integration
│   ├── utils/                 # Utilities
│   │   ├── config.go
│   │   ├── file.go
│   │   ├── middleware.go
│   │   ├── plant.go
│   │   └── *_test.go
│   ├── uploads/               # Uploaded images (served via /uploads/)
│   ├── testdata/              # Test fixtures
│   ├── main.go                # Entry point
│   ├── go.mod                 # Go dependencies
│   ├── .env                   # Environment variables
│   ├── Dockerfile             # Backend container
│   ├── README.md              # Backend documentation
│   └── TESTING.md             # Testing guide
│
└── frontend/                   # Frontend (React)
    ├── public/                # Static assets
    ├── src/
    │   ├── components/        # React components
    │   │   ├── ImageUpload.js
    │   │   ├── ResultsDisplay.js
    │   │   ├── CareInstructions.js
    │   │   ├── ErrorMessage.js
    │   │   ├── Loading.js
    │   │   ├── ChatDrawer.js      # NEW: Chat drawer
    │   │   ├── ChatDrawer.css
    │   │   ├── HistorySidebar.js  # NEW: History sidebar
    │   │   ├── HistorySidebar.css
    │   │   └── *.css
    │   ├── services/          # API layer
    │   │   └── api.js         # NEW: Axios HTTP client
    │   ├── App.js             # Main app component
    │   ├── App.css
    │   ├── index.js
    │   └── index.css
    ├── package.json           # Node dependencies
    └── README.md              # Frontend documentation
```

## 📡 API Documentation

Full API documentation is available in the OpenAPI/Swagger specification:
- **File**: [`swagger.yaml`](./swagger.yaml)
- **View**: Open the file in [Swagger Editor](https://editor.swagger.io/) or use Swagger UI

### Quick Reference

**Backend API** (`http://localhost:8080`):
- `POST /identify` - Identify plant from image
- `POST /chat` - Chat with AI about identified plant
- `GET /history` - List past identifications
- `GET /history/:id` - Get identification details
- `GET /chat/:identification_id` - Get chat history
- `GET /uploads/:filename` - Serve uploaded images
- `GET /health` - Health check

**ML Service** (`http://localhost:8000`):
- `POST /infer` - Get plant predictions
- `GET /` - Health check and model status

## 📊 Model Performance

### Training Results

- **Model**: EfficientNet-B0 (transfer learning)
- **Training Dataset**: 870 images (3 species, 290 each)
- **Train/Val Split**: 80/20
- **Validation Accuracy**: 99-100%
- **Training Accuracy**: ~98%
- **Training Time**: ~25 epochs
- **Final Loss**: 0.05-0.1

### Real-World Testing Results

Tested on actual plant images:

| Species | Confidence | Status |
|---------|-----------|--------|
| Opuntia microdasys | 97.3% | ✅ Excellent |
| Haworthia zebrina | 94.68% | ✅ Excellent |
| Cryptanthus bivittatus | 83.37% | ✅ Good |

### Inference Performance

- **Average inference time**: < 1 second
- **API response time**: < 2 seconds end-to-end
- **Throughput**: Suitable for real-time web applications

## 🔧 Configuration

### Environment Variables

#### Backend (backend/.env)
```bash
# Server Configuration
PORT=8080
ML_SERVICE_URL=http://localhost:8000
CARE_DATA_PATH=../care_data.json
UPLOAD_DIR=./uploads
SPECIES_THRESHOLD=0.4

# Database Configuration
DATABASE_URL=postgresql://username:password@localhost:5432/succulent_identifier?sslmode=disable

# OpenAI Configuration
OPENAI_API_KEY=sk-your-openai-api-key-here
```

#### Frontend (frontend/.env)
```bash
REACT_APP_API_URL=http://localhost:8080
```

## 🧪 Testing

### Backend Tests

Run all backend tests:
```bash
cd backend
go test ./... -v
```

Run with coverage:
```bash
go test ./... -cover
```

Expected coverage: 70-90%

**Test Suite Includes**:
- Identification handler tests
- Chat handler tests (8 tests)
- History handler tests (17 tests)
- Database repository tests
- ML client tests with mock server
- Care data service tests
- File operations tests
- Configuration tests

**Total: 94+ tests**

### Frontend Tests

```bash
cd frontend
npm test
```

### Manual End-to-End Testing

1. Start all services
2. Upload test images from `ml_service/data/raw/`
3. Verify predictions and care instructions

## 🎯 Future Enhancements

### Implemented ✅
- [x] **AI Chat Assistant** - Ask questions about identified plants
- [x] **Identification History** - View and access past identifications
- [x] **PostgreSQL Database** - Persistent storage for data
- [x] **Image Display** - View historical images

### Planned 📋
- [ ] Docker Compose for easy deployment
- [ ] Support for more succulent species (50+ planned)
- [ ] User authentication and multi-user support
- [ ] Mobile app (React Native)
- [ ] Image cropping before upload
- [ ] Batch upload support
- [ ] Community contributions (user-uploaded images)
- [ ] Plant disease detection
- [ ] Watering reminders
- [ ] Progressive Web App (PWA) support
- [ ] Chat history export
- [ ] Streaming responses for chat

## 🤝 Contributing

### Adding New Species

1. **Collect images** using the scraper:
   ```bash
   cd ml_service
   python scraper.py
   # Edit species_to_scrape array in the file
   ```

2. **Add to training data**:
   ```bash
   mv new_species/ ml_service/data/raw/
   ```

3. **Update labels**:
   - Edit `ml_service/labels.json`

4. **Add care data**:
   - Edit `care_data.json`
   - Add both genus and species entries

5. **Retrain the model**:
   ```bash
   cd ml_service/src
   python train.py
   ```

6. **Test the new model**:
   ```bash
   python inference.py
   # Test via API or frontend
   ```

## 📄 License

This project is part of a personal portfolio. All rights reserved.

## 👨‍💻 Author

Created as a demonstration of full-stack ML application development, showcasing:
- Modern web development (React, Go)
- Machine learning (PyTorch, transfer learning)
- Database design and integration (PostgreSQL)
- AI integration (OpenAI GPT-4o-mini)
- System architecture and API design
- RESTful API development
- Test-driven development (94+ tests)
- Real-time features (chat, history)
- Documentation best practices

## 📞 Support

For questions or issues:
1. Check the individual service README files
2. Review the [TODO.md](./TODO.md) for known issues
3. Refer to the [TDD.txt](./TDD.txt) for technical details

---

**Built with ❤️ using React, Go, PostgreSQL, PyTorch, FastAPI, and OpenAI GPT-4o-mini**
