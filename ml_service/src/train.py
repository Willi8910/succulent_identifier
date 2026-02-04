"""
Training script for succulent plant classifier using EfficientNet-B0
"""

import os
import json
import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import DataLoader, random_split
from torchvision import datasets, transforms, models
from tqdm import tqdm
import matplotlib.pyplot as plt
from pathlib import Path


# Configuration
class Config:
    DATA_DIR = "../data/raw"
    MODEL_DIR = "../models"
    LABELS_PATH = "../labels.json"

    # Training hyperparameters
    BATCH_SIZE = 16
    NUM_EPOCHS = 25
    LEARNING_RATE = 0.001
    TRAIN_SPLIT = 0.8

    # Image parameters
    IMG_SIZE = 224

    # Device
    DEVICE = torch.device("cuda" if torch.cuda.is_available() else "cpu")


def get_data_transforms():
    """Define data augmentation and normalization transforms"""
    train_transforms = transforms.Compose([
        transforms.Resize((Config.IMG_SIZE, Config.IMG_SIZE)),
        transforms.RandomHorizontalFlip(),
        transforms.RandomRotation(15),
        transforms.ColorJitter(brightness=0.2, contrast=0.2, saturation=0.2),
        transforms.ToTensor(),
        transforms.Normalize([0.485, 0.456, 0.406], [0.229, 0.224, 0.225])
    ])

    val_transforms = transforms.Compose([
        transforms.Resize((Config.IMG_SIZE, Config.IMG_SIZE)),
        transforms.ToTensor(),
        transforms.Normalize([0.485, 0.456, 0.406], [0.229, 0.224, 0.225])
    ])

    return train_transforms, val_transforms


def prepare_data(train_transforms, val_transforms):
    """Load and split dataset"""
    # Load full dataset with training transforms first
    full_dataset = datasets.ImageFolder(root=Config.DATA_DIR)

    # Calculate split sizes
    train_size = int(Config.TRAIN_SPLIT * len(full_dataset))
    val_size = len(full_dataset) - train_size

    # Split dataset
    train_dataset, val_dataset = random_split(
        full_dataset,
        [train_size, val_size],
        generator=torch.Generator().manual_seed(42)
    )

    # Apply transforms
    train_dataset.dataset.transform = train_transforms
    val_dataset.dataset.transform = val_transforms

    # Create data loaders
    train_loader = DataLoader(
        train_dataset,
        batch_size=Config.BATCH_SIZE,
        shuffle=True,
        num_workers=4
    )

    val_loader = DataLoader(
        val_dataset,
        batch_size=Config.BATCH_SIZE,
        shuffle=False,
        num_workers=4
    )

    return train_loader, val_loader, full_dataset.classes


def build_model(num_classes):
    """Build EfficientNet-B0 model with transfer learning"""
    # Load pretrained EfficientNet-B0
    model = models.efficientnet_b0(weights=models.EfficientNet_B0_Weights.DEFAULT)

    # Freeze early layers
    for param in model.parameters():
        param.requires_grad = False

    # Replace classifier
    num_features = model.classifier[1].in_features
    model.classifier = nn.Sequential(
        nn.Dropout(p=0.2, inplace=True),
        nn.Linear(num_features, num_classes)
    )

    return model.to(Config.DEVICE)


def train_epoch(model, train_loader, criterion, optimizer):
    """Train for one epoch"""
    model.train()
    running_loss = 0.0
    correct = 0
    total = 0

    pbar = tqdm(train_loader, desc="Training")
    for inputs, labels in pbar:
        inputs, labels = inputs.to(Config.DEVICE), labels.to(Config.DEVICE)

        # Zero gradients
        optimizer.zero_grad()

        # Forward pass
        outputs = model(inputs)
        loss = criterion(outputs, labels)

        # Backward pass
        loss.backward()
        optimizer.step()

        # Statistics
        running_loss += loss.item()
        _, predicted = torch.max(outputs.data, 1)
        total += labels.size(0)
        correct += (predicted == labels).sum().item()

        pbar.set_postfix({
            'loss': f'{running_loss / (pbar.n + 1):.4f}',
            'acc': f'{100 * correct / total:.2f}%'
        })

    epoch_loss = running_loss / len(train_loader)
    epoch_acc = 100 * correct / total

    return epoch_loss, epoch_acc


def validate_epoch(model, val_loader, criterion):
    """Validate for one epoch"""
    model.eval()
    running_loss = 0.0
    correct = 0
    total = 0

    with torch.no_grad():
        pbar = tqdm(val_loader, desc="Validation")
        for inputs, labels in pbar:
            inputs, labels = inputs.to(Config.DEVICE), labels.to(Config.DEVICE)

            # Forward pass
            outputs = model(inputs)
            loss = criterion(outputs, labels)

            # Statistics
            running_loss += loss.item()
            _, predicted = torch.max(outputs.data, 1)
            total += labels.size(0)
            correct += (predicted == labels).sum().item()

            pbar.set_postfix({
                'loss': f'{running_loss / (pbar.n + 1):.4f}',
                'acc': f'{100 * correct / total:.2f}%'
            })

    epoch_loss = running_loss / len(val_loader)
    epoch_acc = 100 * correct / total

    return epoch_loss, epoch_acc


def plot_training_history(history, save_path):
    """Plot and save training history"""
    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(12, 4))

    # Loss plot
    ax1.plot(history['train_loss'], label='Train Loss')
    ax1.plot(history['val_loss'], label='Val Loss')
    ax1.set_xlabel('Epoch')
    ax1.set_ylabel('Loss')
    ax1.set_title('Training and Validation Loss')
    ax1.legend()
    ax1.grid(True)

    # Accuracy plot
    ax2.plot(history['train_acc'], label='Train Acc')
    ax2.plot(history['val_acc'], label='Val Acc')
    ax2.set_xlabel('Epoch')
    ax2.set_ylabel('Accuracy (%)')
    ax2.set_title('Training and Validation Accuracy')
    ax2.legend()
    ax2.grid(True)

    plt.tight_layout()
    plt.savefig(save_path)
    print(f"Training history plot saved to {save_path}")


def save_labels_mapping(classes, save_path):
    """Save class index to label mapping"""
    # Convert folder names to proper format: genus_species
    labels_map = {}
    for idx, class_name in enumerate(classes):
        # Remove genus prefix if present in folder name
        # e.g., "cryptanthus-cryptanthus_bivittatus" -> "cryptanthus_bivittatus"
        if '-' in class_name:
            parts = class_name.split('-')
            label = parts[1]  # Take the second part
        else:
            label = class_name

        labels_map[str(idx)] = label

    with open(save_path, 'w') as f:
        json.dump(labels_map, f, indent=2)

    print(f"Labels mapping saved to {save_path}")


def main():
    """Main training function"""
    print(f"Using device: {Config.DEVICE}")
    print(f"Training on data from: {Config.DATA_DIR}")

    # Create model directory if it doesn't exist
    os.makedirs(Config.MODEL_DIR, exist_ok=True)

    # Prepare data
    print("\nPreparing data...")
    train_transforms, val_transforms = get_data_transforms()
    train_loader, val_loader, classes = prepare_data(train_transforms, val_transforms)

    print(f"Number of classes: {len(classes)}")
    print(f"Classes: {classes}")
    print(f"Training samples: {len(train_loader.dataset)}")
    print(f"Validation samples: {len(val_loader.dataset)}")

    # Build model
    print("\nBuilding model...")
    num_classes = len(classes)
    model = build_model(num_classes)

    # Loss and optimizer
    criterion = nn.CrossEntropyLoss()
    optimizer = optim.Adam(model.classifier.parameters(), lr=Config.LEARNING_RATE)
    scheduler = optim.lr_scheduler.ReduceLROnPlateau(
        optimizer, mode='min', patience=3, factor=0.5
    )

    # Training loop
    print(f"\nStarting training for {Config.NUM_EPOCHS} epochs...")
    history = {
        'train_loss': [],
        'train_acc': [],
        'val_loss': [],
        'val_acc': []
    }

    best_val_acc = 0.0

    for epoch in range(Config.NUM_EPOCHS):
        print(f"\nEpoch {epoch + 1}/{Config.NUM_EPOCHS}")
        print("-" * 50)

        # Train
        train_loss, train_acc = train_epoch(model, train_loader, criterion, optimizer)

        # Validate
        val_loss, val_acc = validate_epoch(model, val_loader, criterion)

        # Update learning rate
        scheduler.step(val_loss)

        # Save history
        history['train_loss'].append(train_loss)
        history['train_acc'].append(train_acc)
        history['val_loss'].append(val_loss)
        history['val_acc'].append(val_acc)

        print(f"\nEpoch {epoch + 1} Summary:")
        print(f"Train Loss: {train_loss:.4f}, Train Acc: {train_acc:.2f}%")
        print(f"Val Loss: {val_loss:.4f}, Val Acc: {val_acc:.2f}%")

        # Save best model
        if val_acc > best_val_acc:
            best_val_acc = val_acc
            model_path = os.path.join(Config.MODEL_DIR, "succulent_classifier_best.pth")
            torch.save({
                'epoch': epoch,
                'model_state_dict': model.state_dict(),
                'optimizer_state_dict': optimizer.state_dict(),
                'val_acc': val_acc,
                'classes': classes
            }, model_path)
            print(f"Best model saved! Val Acc: {val_acc:.2f}%")

    # Save final model
    final_model_path = os.path.join(Config.MODEL_DIR, "succulent_classifier_final.pth")
    torch.save({
        'epoch': Config.NUM_EPOCHS,
        'model_state_dict': model.state_dict(),
        'optimizer_state_dict': optimizer.state_dict(),
        'val_acc': val_acc,
        'classes': classes
    }, final_model_path)
    print(f"\nFinal model saved to {final_model_path}")

    # Plot training history
    plot_path = os.path.join(Config.MODEL_DIR, "training_history.png")
    plot_training_history(history, plot_path)

    # Save labels mapping
    save_labels_mapping(classes, Config.LABELS_PATH)

    print(f"\nTraining completed!")
    print(f"Best validation accuracy: {best_val_acc:.2f}%")


if __name__ == "__main__":
    main()
