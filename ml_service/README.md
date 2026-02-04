# Succulent Identifier - ML Service

Machine learning service for succulent plant identification using PyTorch and FastAPI.

## Overview

This service provides image classification capabilities for identifying succulent plants. It uses transfer learning with EfficientNet-B0 and exposes a REST API for inference.

## Directory Structure

```
ml_service/
├── data/
│   └── raw/              # Training dataset (organized by species)
├── models/               # Trained model files
├── src/
│   ├── train.py         # Model training script
│   ├── inference.py     # FastAPI inference service
│   └── preprocessing.py # Image preprocessing utilities
├── labels.json          # Class label mappings
├── requirements.txt     # Python dependencies
├── Dockerfile           # Docker configuration
└── README.md           # This file
```

## Requirements

- Python 3.10+
- PyTorch 2.1.0
- FastAPI 0.104.1
- Other dependencies listed in `requirements.txt`

## Installation

### Local Setup

1. Create a virtual environment:
```bash
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
```

2. Install dependencies:
```bash
pip install -r requirements.txt
```

### Docker Setup

Build the Docker image:
```bash
docker build -t succulent-ml-service .
```

## Dataset Preparation

Organize your training data in the following structure:

```
data/raw/
├── cryptanthus-cryptanthus_bivittatus/
│   ├── image1.jpg
│   ├── image2.jpg
│   └── ...
├── haworthia-haworthia_zebrina/
│   ├── image1.jpg
│   └── ...
└── opuntia-opuntia_microdasys/
    ├── image1.jpg
    └── ...
```

Each subdirectory should contain images for one species. The folder naming convention is:
`<genus>-<genus>_<species>`

## Training

### Run Training Locally

```bash
cd src
python train.py
```

The training script will:
- Load images from `data/raw/`
- Split data into train/validation sets (80/20)
- Train EfficientNet-B0 with transfer learning
- Save the best model to `models/succulent_classifier_best.pth`
- Save training history plot to `models/training_history.png`
- Update `labels.json` with class mappings

### Training Configuration

Edit the `Config` class in `src/train.py` to adjust:
- `BATCH_SIZE`: Training batch size (default: 16)
- `NUM_EPOCHS`: Number of training epochs (default: 25)
- `LEARNING_RATE`: Initial learning rate (default: 0.001)
- `TRAIN_SPLIT`: Train/validation split ratio (default: 0.8)
- `IMG_SIZE`: Input image size (default: 224)

## Inference Service

### Run Service Locally

```bash
cd src
python inference.py
```

The service will start on `http://localhost:8000`

### Run Service with Docker

```bash
docker run -p 8000:8000 \
  -v $(pwd)/models:/app/models \
  succulent-ml-service
```

### Environment Variables

- `MODEL_PATH`: Path to model file (default: `../models/succulent_classifier_best.pth`)
- `LABELS_PATH`: Path to labels file (default: `../labels.json`)
- `TOP_K`: Number of top predictions to return (default: 3)
- `PORT`: Service port (default: 8000)

## API Endpoints

### Health Check

```bash
GET /
GET /health
```

Response:
```json
{
  "status": "healthy",
  "model_loaded": true,
  "labels_loaded": true,
  "num_classes": 3,
  "device": "cpu"
}
```

### Inference

```bash
POST /infer
Content-Type: application/json

{
  "image_path": "/path/to/image.jpg"
}
```

Response:
```json
{
  "predictions": [
    {
      "label": "echeveria_perle_von_nurnberg",
      "confidence": 0.52
    },
    {
      "label": "echeveria_elegans",
      "confidence": 0.21
    },
    {
      "label": "haworthia_zebrina",
      "confidence": 0.15
    }
  ]
}
```

### Example Usage

```bash
# Test inference
curl -X POST "http://localhost:8000/infer" \
  -H "Content-Type: application/json" \
  -d '{"image_path": "/path/to/succulent.jpg"}'
```

## Model Details

### Architecture
- Base model: EfficientNet-B0 (pretrained on ImageNet)
- Transfer learning: Early layers frozen, classifier fine-tuned
- Input size: 224x224 RGB images
- Output: Softmax probabilities over species classes

### Data Augmentation (Training)
- Random horizontal flip
- Random rotation (±15°)
- Color jitter (brightness, contrast, saturation)
- Normalization (ImageNet statistics)

### Preprocessing (Inference)
- Resize to 224x224
- Normalize with ImageNet statistics
- Convert to tensor

## Performance Considerations

### CPU vs GPU
- The service automatically detects and uses GPU if available
- CPU inference is sufficient for single-image requests
- Expected inference time: <1 second on CPU

### Model Size
- EfficientNet-B0: ~17MB
- Lightweight and suitable for production deployment

## Troubleshooting

### Model Not Loading
- Ensure `models/succulent_classifier_best.pth` exists
- Check that the model was trained with the correct number of classes
- Verify `labels.json` matches the model's class count

### Image Loading Errors
- Verify image path is absolute and accessible
- Supported formats: JPG, PNG
- Ensure images are valid and not corrupted

### Memory Issues
- Reduce `BATCH_SIZE` during training
- Use CPU instead of GPU if VRAM is limited
- Process images one at a time during inference

## Future Improvements

- Add batch inference endpoint
- Implement model versioning
- Add data augmentation strategies
- Support for additional plant families
- Model quantization for faster inference
- Confidence calibration

## License

Internal use only.

## Contact

For questions or issues, contact the development team.
