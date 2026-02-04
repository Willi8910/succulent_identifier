# Succulent Identifier - Backend API

Golang REST API that orchestrates plant identification by coordinating between frontend requests, ML service inference, and care data delivery.

## Overview

The backend serves as the middleware layer between the frontend and ML service, handling:
- Image upload and validation
- ML service communication
- Confidence-based prediction logic
- Care data retrieval and fallback

## Directory Structure

```
backend/
├── handlers/          # HTTP request handlers
│   └── identify.go   # Main identification endpoint
├── models/           # Data structures and types
│   └── types.go      # Request/response models
├── services/         # Business logic services
│   ├── care_data.go # Care instructions loader
│   └── ml_client.go # ML service client
├── utils/            # Utilities and helpers
│   ├── config.go    # Configuration management
│   ├── file.go      # File upload handling
│   ├── middleware.go # HTTP middleware (CORS, logging)
│   └── plant.go     # Label parsing utilities
├── uploads/          # Temporary file storage
├── main.go          # Application entry point
├── go.mod           # Go module definition
├── Dockerfile       # Docker configuration
└── README.md        # This file
```

## Requirements

- Go 1.21 or higher
- ML Service running (for inference)
- Care data JSON file

## Installation

### Local Setup

1. **Install Go dependencies:**
```bash
go mod download
```

2. **Verify care data exists:**
```bash
# Care data should be at ../care_data.json
ls -la ../care_data.json
```

3. **Set environment variables (optional):**
```bash
export SERVER_PORT=8080
export ML_SERVICE_URL=http://localhost:8000
export UPLOAD_DIR=./uploads
export MAX_FILE_SIZE=5242880
export SPECIES_THRESHOLD=0.4
export CARE_DATA_PATH=../care_data.json
```

4. **Run the server:**
```bash
go run main.go
```

### Docker Setup

Build and run with Docker:

```bash
# Build image
docker build -t succulent-backend .

# Run container
docker run -p 8080:8080 \
  -v $(pwd)/../care_data.json:/root/care_data.json \
  -v $(pwd)/uploads:/root/uploads \
  -e ML_SERVICE_URL=http://host.docker.internal:8000 \
  succulent-backend
```

## Configuration

Configuration is managed through environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | Port for the API server | `8080` |
| `ML_SERVICE_URL` | URL of ML inference service | `http://localhost:8000` |
| `UPLOAD_DIR` | Directory for uploaded images | `./uploads` |
| `MAX_FILE_SIZE` | Maximum file size in bytes | `5242880` (5MB) |
| `SPECIES_THRESHOLD` | Confidence threshold for species display | `0.4` |
| `CARE_DATA_PATH` | Path to care data JSON file | `../care_data.json` |

## API Endpoints

### Health Check

```
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "succulent-identifier-backend"
}
```

### Root

```
GET /
```

**Response:**
```json
{
  "service": "Succulent Identifier Backend",
  "version": "1.0.0",
  "endpoints": ["/identify", "/health"]
}
```

### Identify Plant

```
POST /identify
Content-Type: multipart/form-data
```

**Request:**
- `image`: Image file (JPG/PNG, max 5MB)

**Response (High Confidence ≥ 0.4):**
```json
{
  "plant": {
    "genus": "Haworthia",
    "species": "Haworthia zebrina",
    "confidence": 0.85
  },
  "care": {
    "sunlight": "Bright indirect light...",
    "watering": "Water thoroughly when soil is completely dry...",
    "soil": "Very well-draining cactus mix...",
    "notes": "Also called Zebra Plant..."
  }
}
```

**Response (Low Confidence < 0.4):**
```json
{
  "plant": {
    "genus": "Haworthia",
    "confidence": 0.32
  },
  "care": {
    "sunlight": "Bright indirect light to partial sun...",
    "watering": "Water deeply but infrequently...",
    "soil": "Well-draining succulent or cactus mix...",
    "notes": "Very tolerant and low-maintenance..."
  }
}
```

**Error Response:**
```json
{
  "error": "Bad Request",
  "message": "file type '.pdf' not allowed. Allowed types: [.jpg .jpeg .png]"
}
```

## Business Logic

### Confidence Threshold Logic

The API implements confidence-based prediction display:

1. **High Confidence (≥ threshold)**:
   - Display genus + species
   - Use species-level care if available
   - Fall back to genus care if species care not found

2. **Low Confidence (< threshold)**:
   - Display genus only
   - Use genus-level care

### Label Parsing

Labels from ML service are formatted as `genus_species` (e.g., `haworthia_zebrina`):
- Genus: First part before underscore
- Species: Full label (used as care data key)

### Care Data Fallback

Care instruction retrieval follows this priority:
1. Species-level care (`genus_species`)
2. Genus-level care (`genus`)
3. Generic "information not available" message

## File Upload

### Validation

Uploaded files are validated for:
- **File size**: Must not exceed MAX_FILE_SIZE (default 5MB)
- **File type**: Must be JPG, JPEG, or PNG
- **Non-empty**: File must contain data

### Storage

- Files are saved with UUID-generated names
- Stored in UPLOAD_DIR directory
- Optional cleanup after processing (configurable)

## Error Handling

The API handles various error scenarios:

- **400 Bad Request**: Invalid file type, size, or missing image
- **404 Not Found**: Invalid endpoint
- **405 Method Not Allowed**: Wrong HTTP method
- **500 Internal Server Error**: ML service failure, care data issues

## Development

### Running Locally

```bash
# Install dependencies
go mod tidy

# Run with live reload (requires air)
go install github.com/cosmtrek/air@latest
air

# Or run normally
go run main.go
```

### Testing

```bash
# Test identify endpoint
curl -X POST http://localhost:8080/identify \
  -F "image=@/path/to/succulent.jpg"

# Test with invalid file
curl -X POST http://localhost:8080/identify \
  -F "image=@/path/to/document.pdf"
```

### Building

```bash
# Build binary
go build -o succulent-backend

# Run binary
./succulent-backend
```

## Integration

### With ML Service

The backend communicates with the ML service via HTTP:

```go
POST {ML_SERVICE_URL}/infer
Content-Type: application/json

{
  "image_path": "/absolute/path/to/image.jpg"
}
```

**Important**: The ML service must be running before starting the backend, or requests will fail.

### With Frontend

The frontend sends multipart form data to `/identify`:

```javascript
const formData = new FormData();
formData.append('image', fileBlob);

fetch('http://localhost:8080/identify', {
  method: 'POST',
  body: formData
});
```

## Dependencies

Main dependencies:
- `github.com/google/uuid` - UUID generation for filenames

All dependencies are managed via `go.mod`.

## Troubleshooting

### ML Service Unreachable

**Error:** `Failed to call ML service: dial tcp: connection refused`

**Solution:**
- Ensure ML service is running
- Check ML_SERVICE_URL is correct
- If using Docker, use `host.docker.internal` instead of `localhost`

### Care Data Not Found

**Error:** `Failed to initialize care data service: no such file`

**Solution:**
- Verify CARE_DATA_PATH points to valid JSON file
- Check file permissions
- Ensure care_data.json is in correct format

### File Upload Fails

**Error:** `file type '.xyz' not allowed`

**Solution:**
- Only JPG, JPEG, PNG files are accepted
- Check file size doesn't exceed 5MB
- Ensure form field name is `image`

## Performance

- **Upload handling**: Buffered multipart parsing (10MB limit)
- **ML inference**: Synchronous (blocks until ML service responds)
- **File storage**: Direct disk write with UUID naming
- **Memory**: Minimal footprint, files streamed to disk

## Security Considerations

- **File validation**: Type and size checks prevent abuse
- **UUID filenames**: Prevents path traversal attacks
- **CORS enabled**: Allows cross-origin requests (configure for production)
- **No authentication**: V1 has no auth (add in future versions)

## Future Improvements

- Add request rate limiting
- Implement file cleanup scheduler
- Add authentication/authorization
- Support batch image processing
- Add caching for ML predictions
- Implement request logging
- Add metrics and monitoring

## License

Internal use only.

## Contact

For questions or issues, contact the development team.
