"""
Image preprocessing utilities for inference
"""

from PIL import Image
import torch
from torchvision import transforms


class ImagePreprocessor:
    """Handles image preprocessing for model inference"""

    def __init__(self, img_size=224):
        """
        Initialize preprocessor with image transformations

        Args:
            img_size: Target image size for model input
        """
        self.img_size = img_size
        self.transform = transforms.Compose([
            transforms.Resize((img_size, img_size)),
            transforms.ToTensor(),
            transforms.Normalize(
                mean=[0.485, 0.456, 0.406],
                std=[0.229, 0.224, 0.225]
            )
        ])

    def load_image(self, image_path):
        """
        Load image from file path

        Args:
            image_path: Path to image file

        Returns:
            PIL Image object

        Raises:
            FileNotFoundError: If image file doesn't exist
            IOError: If image cannot be opened
        """
        try:
            image = Image.open(image_path)
            # Convert to RGB if necessary
            if image.mode != 'RGB':
                image = image.convert('RGB')
            return image
        except FileNotFoundError:
            raise FileNotFoundError(f"Image file not found: {image_path}")
        except Exception as e:
            raise IOError(f"Error loading image {image_path}: {str(e)}")

    def preprocess(self, image):
        """
        Preprocess image for model inference

        Args:
            image: PIL Image object

        Returns:
            Preprocessed tensor with batch dimension
        """
        # Apply transformations
        tensor = self.transform(image)
        # Add batch dimension
        tensor = tensor.unsqueeze(0)
        return tensor

    def preprocess_from_path(self, image_path):
        """
        Load and preprocess image from file path

        Args:
            image_path: Path to image file

        Returns:
            Preprocessed tensor with batch dimension
        """
        image = self.load_image(image_path)
        return self.preprocess(image)


def validate_image_file(file_path):
    """
    Validate if file is a valid image

    Args:
        file_path: Path to file

    Returns:
        tuple: (is_valid, error_message)
    """
    try:
        image = Image.open(file_path)
        image.verify()
        return True, None
    except Exception as e:
        return False, str(e)
