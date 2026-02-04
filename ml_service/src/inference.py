"""
FastAPI inference service for succulent plant classification
"""

import os
import json
import torch
import torch.nn as nn
from contextlib import asynccontextmanager
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import List, Dict
from torchvision import models
import logging

from preprocessing import ImagePreprocessor

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


# Configuration
class Config:
    MODEL_PATH = os.getenv("MODEL_PATH", "../models/succulent_classifier_best.pth")
    LABELS_PATH = os.getenv("LABELS_PATH", "../labels.json")
    TOP_K = int(os.getenv("TOP_K", "3"))
    IMG_SIZE = 224
    DEVICE = torch.device("cuda" if torch.cuda.is_available() else "cpu")


# Request/Response models
class InferenceRequest(BaseModel):
    image_path: str


class Prediction(BaseModel):
    label: str
    confidence: float


class InferenceResponse(BaseModel):
    predictions: List[Prediction]


# Global variables for model and labels
model = None
labels_map = None
preprocessor = None


def load_labels(labels_path: str) -> Dict[int, str]:
    """
    Load labels mapping from JSON file

    Args:
        labels_path: Path to labels.json file

    Returns:
        Dictionary mapping class indices to labels
    """
    try:
        with open(labels_path, 'r') as f:
            labels_data = json.load(f)

        # Convert string keys to integers
        labels_map = {int(k): v for k, v in labels_data.items()}
        logger.info(f"Loaded {len(labels_map)} labels from {labels_path}")
        return labels_map

    except FileNotFoundError:
        logger.error(f"Labels file not found: {labels_path}")
        raise
    except json.JSONDecodeError as e:
        logger.error(f"Error parsing labels JSON: {e}")
        raise
    except Exception as e:
        logger.error(f"Unexpected error loading labels: {e}")
        raise


def build_model(num_classes: int) -> nn.Module:
    """
    Build EfficientNet-B0 model architecture

    Args:
        num_classes: Number of output classes

    Returns:
        Model instance
    """
    model = models.efficientnet_b0(weights=None)

    # Replace classifier to match training configuration
    num_features = model.classifier[1].in_features
    model.classifier = nn.Sequential(
        nn.Dropout(p=0.2, inplace=True),
        nn.Linear(num_features, num_classes)
    )

    return model


def load_model(model_path: str, num_classes: int, device: torch.device) -> nn.Module:
    """
    Load trained model from checkpoint

    Args:
        model_path: Path to model checkpoint file
        num_classes: Number of output classes
        device: Device to load model on

    Returns:
        Loaded model in evaluation mode
    """
    try:
        # Build model architecture
        model = build_model(num_classes)

        # Load checkpoint
        checkpoint = torch.load(model_path, map_location=device)

        # Load model weights
        model.load_state_dict(checkpoint['model_state_dict'])

        # Move to device and set to eval mode
        model = model.to(device)
        model.eval()

        logger.info(f"Model loaded successfully from {model_path}")
        logger.info(f"Model validation accuracy: {checkpoint.get('val_acc', 'N/A')}")

        return model

    except FileNotFoundError:
        logger.error(f"Model file not found: {model_path}")
        raise
    except Exception as e:
        logger.error(f"Error loading model: {e}")
        raise


def predict(image_path: str, top_k: int = 3) -> List[Dict[str, float]]:
    """
    Run inference on an image

    Args:
        image_path: Path to input image
        top_k: Number of top predictions to return

    Returns:
        List of top-k predictions with labels and confidence scores
    """
    global model, labels_map, preprocessor

    if model is None or labels_map is None or preprocessor is None:
        raise RuntimeError("Model not loaded. Service not initialized properly.")

    try:
        # Preprocess image
        input_tensor = preprocessor.preprocess_from_path(image_path)
        input_tensor = input_tensor.to(Config.DEVICE)

        # Run inference
        with torch.no_grad():
            outputs = model(input_tensor)
            probabilities = torch.nn.functional.softmax(outputs, dim=1)

        # Get top-k predictions
        top_probs, top_indices = torch.topk(probabilities, min(top_k, len(labels_map)))

        # Format predictions
        predictions = []
        for prob, idx in zip(top_probs[0], top_indices[0]):
            predictions.append({
                "label": labels_map[idx.item()],
                "confidence": round(prob.item(), 4)
            })

        return predictions

    except FileNotFoundError as e:
        logger.error(f"Image file not found: {image_path}")
        raise HTTPException(status_code=404, detail=f"Image file not found: {image_path}")
    except Exception as e:
        logger.error(f"Error during inference: {e}")
        raise HTTPException(status_code=500, detail=f"Inference error: {str(e)}")


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Lifespan event handler for startup and shutdown"""
    global model, labels_map, preprocessor

    # Startup
    logger.info("Starting ML inference service...")
    logger.info(f"Using device: {Config.DEVICE}")

    try:
        # Load labels
        labels_map = load_labels(Config.LABELS_PATH)
        num_classes = len(labels_map)

        # Load model
        model = load_model(Config.MODEL_PATH, num_classes, Config.DEVICE)

        # Initialize preprocessor
        preprocessor = ImagePreprocessor(img_size=Config.IMG_SIZE)

        logger.info("Service initialized successfully!")

    except Exception as e:
        logger.error(f"Failed to initialize service: {e}")
        raise

    yield

    # Shutdown
    logger.info("Shutting down ML inference service...")


# Initialize FastAPI app
app = FastAPI(
    title="Succulent Identifier ML Service",
    description="Image classification service for succulent plant identification",
    version="1.0.0",
    lifespan=lifespan
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/")
async def root():
    """Health check endpoint"""
    return {
        "service": "Succulent Identifier ML Service",
        "status": "running",
        "model_loaded": model is not None,
        "device": str(Config.DEVICE)
    }


@app.get("/health")
async def health():
    """Detailed health check"""
    return {
        "status": "healthy",
        "model_loaded": model is not None,
        "labels_loaded": labels_map is not None,
        "num_classes": len(labels_map) if labels_map else 0,
        "device": str(Config.DEVICE)
    }


@app.post("/infer", response_model=InferenceResponse)
async def infer(request: InferenceRequest):
    """
    Run inference on an image

    Args:
        request: InferenceRequest with image_path

    Returns:
        InferenceResponse with top-k predictions
    """
    logger.info(f"Inference request for image: {request.image_path}")

    try:
        predictions = predict(request.image_path, top_k=Config.TOP_K)
        logger.info(f"Inference completed. Top prediction: {predictions[0]['label']} "
                   f"(confidence: {predictions[0]['confidence']:.2%})")

        return InferenceResponse(predictions=predictions)

    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Unexpected error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


if __name__ == "__main__":
    import uvicorn

    port = int(os.getenv("PORT", "8000"))
    uvicorn.run(app, host="0.0.0.0", port=port)
