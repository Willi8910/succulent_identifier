# Dockerfile for Succulent Identifier ML Service

FROM python:3.10-slim

# Set working directory
WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements first for better caching
COPY requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY src/ ./src/
COPY labels.json .

# Create directory for models (will be mounted or copied)
RUN mkdir -p models

# Expose port
EXPOSE 8000

# Environment variables
ENV MODEL_PATH=/app/models/succulent_classifier_best.pth
ENV LABELS_PATH=/app/labels.json
ENV TOP_K=3
ENV PORT=8000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD python -c "import requests; requests.get('http://localhost:8000/health')"

# Run the inference service
CMD ["python", "-m", "uvicorn", "src.inference:app", "--host", "0.0.0.0", "--port", "8000"]
