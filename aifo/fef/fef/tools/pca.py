import logging
import re
from pathlib import Path
from typing import List, Optional, Tuple

import cv2
import matplotlib.pyplot as plt
import numpy as np
import torch
import torchvision.transforms as tt
from dinov2.configs import load_and_merge_config
from dinov2.eval.setup import build_model_for_eval
from sklearn.decomposition import PCA
from sklearn.preprocessing import minmax_scale
from torch import nn

log = logging.getLogger(__name__)


def load_dino_model(
    model_path: str | Path,
    config_name: str = "eval/vits14_pretrain",
    image_size: int = 224,
) -> nn.Module:
    """
    Load a DINO model from the given model name and configuration path.

    Parameters:
    - model_path: Path to the pretrained model.
    - config_name: Name of the configuration file.
    - image_size: input size of the images.

    Returns:
    - number of iterations, The pretrained model.
    """
    if re.search(r"training_(\d+)", model_path) is not None:
        iterations = int(re.search(r"training_(\d+)", model_path).group(1))
    else:
        log.warning("No iteration number found in the model path. Using -1 as the iteration number.")
        iterations = -1
    conf = load_and_merge_config(config_name)
    conf.crops.global_crops_size = image_size
    model = build_model_for_eval(conf, model_path)
    return iterations, model


def load_and_preprocess_images(
    image_paths: List[str], size: Tuple[int, int] = (224, 224), channels: int = 3
) -> Tuple[List[np.ndarray], torch.Tensor]:
    """
    Load and preprocess images for model evaluation.

    Parameters:
    - image_paths: List of paths to the images.
    - size: Desired size of the images as a tuple (width, height).

    Returns:
    - A torch.Tensor of preprocessed images.
    """
    images = []
    for path in image_paths:
        image = cv2.imread(path)
        image = cv2.resize(image, size)
        if channels == 1:
            image = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY).astype("float32") / 255.0
        elif channels == 3:
            image = cv2.cvtColor(image, cv2.COLOR_BGR2RGB).astype("float32") / 255.0
        else:
            raise ValueError("Invalid number of channels. Must be 1 or 3.")
        images.append(np.expand_dims(image, -1) if channels == 1 else image)

    images_array = np.stack(images)
    input_tensor = torch.Tensor(np.transpose(images_array, [0, 3, 2, 1]))

    # Normalize the images
    if channels == 1:
        transform = tt.Compose([tt.Normalize(mean=(0.1008,), std=(0.193391,))])
    elif channels == 3:
        transform = tt.Compose([tt.Normalize(mean=(0.1008, 0.1008, 0.1008), std=(0.193391, 0.193391, 0.193391))])
    else:
        raise ValueError("Invalid number of channels. Must be 1 or 3.")
    return images, transform(input_tensor)


def evaluate_model(model: torch.nn.Module, input_tensor: torch.Tensor) -> Tuple[np.ndarray, np.ndarray]:
    """
    Evaluate the model on the given input tensor and return features necessary for PCA analysis.

    Parameters:
    - model: The pretrained model to evaluate.
    - input_tensor: The preprocessed input tensor.

    Returns:
    - Tuple of numpy arrays: (output_cls, output_patch)
    """
    output = model.forward_features(input_tensor)
    output_cls = output["x_norm_clstoken"].cpu().detach().numpy()
    output_patch = output["x_norm_patchtokens"].cpu().detach().numpy()
    return output_cls, output_patch


def perform_pca_and_plot_old(
    images: List[np.ndarray],
    output_patch: np.ndarray,
    save_path: str,
    plot_title: str | None = None,
) -> None:
    """
    Perform PCA analysis on the model outputs and generate plots, saving them to a specified path.

    Parameters:
    - images: List of original images.
    - output_patch: Model output patches for PCA analysis.
    - save_path: Path to save the plot.
    - threshold: Threshold for masking in PCA analysis.
    """
    num_images = len(images)
    embedding_size = output_patch.shape[-1]
    pca = PCA(n_components=3)  # Adjust components for RGB visualization
    plt.figure(figsize=(6, 2 * num_images))

    # Perform PCA on reshaped patches
    all_patches = output_patch.reshape([-1, embedding_size])
    reduced_patches = pca.fit_transform(all_patches)
    reduced_patches = minmax_scale(reduced_patches)  # Scale features to (0,1)

    for i in range(num_images):
        plt.subplot(num_images, 2, 2 * i + 1)
        plt.imshow(images[i])

        pca_image = reduced_patches[i * 256 : (i + 1) * 256].reshape(16, 16, 3)
        plt.subplot(num_images, 2, 2 * i + 2)
        plt.imshow(pca_image.transpose(1, 0, 2))
        if plot_title is not None:
            plt.title(plot_title, fontsize=10)  # Add the title if provided

    plt.tight_layout()
    plt.savefig(save_path)
    plt.close()


def compute_foreground_masks(output_patch: np.ndarray, threshold: float = 0.0) -> List[np.ndarray]:
    """
    Computes foreground masks for each image using PCA.

    Parameters:
    - output_patch: Model output patches for PCA analysis.
    - threshold: Threshold for masking in PCA analysis.

    Returns:
    - masks: List of boolean arrays indicating foreground patches for each image.
    """
    num_images = output_patch.shape[0]
    embedding_size = output_patch.shape[-1]
    fg_bg_pca = PCA(n_components=1)

    # Perform PCA on reshaped patches
    all_patches = output_patch.reshape(-1, embedding_size)
    reduced_patches = fg_bg_pca.fit_transform(all_patches)
    norm_patches = minmax_scale(reduced_patches)  # Scale features to (0,1)

    # Reshape the feature values to per-image patches
    image_norm_patches = norm_patches.reshape(num_images, -1)

    masks = []
    for i in range(num_images):
        image_patches = image_norm_patches[i, :]
        mask = image_patches > threshold
        masks.append(mask)

    return masks


def perform_pca_on_foreground_patches(
    output_patch: np.ndarray, masks: List[np.ndarray]
) -> Tuple[np.ndarray, List[int]]:
    """
    Performs PCA on the foreground patches and returns the reduced patches and mask indices.

    Parameters:
    - output_patch: Model output patches.
    - masks: List of masks indicating foreground patches for each image.

    Returns:
    - reduced_patches: PCA-reduced patches of foreground regions.
    - mask_indices: List of starting indices for the patches of each image in reduced_patches.
    """
    object_pca = PCA(n_components=3)

    # Extract foreground patches for all images
    fg_patches_list = [output_patch[i][masks[i]] for i in range(len(masks))]
    fg_patches = np.vstack(fg_patches_list)

    # Compute mask indices to map back to images
    lengths = [len(p) for p in fg_patches_list]
    mask_indices = np.cumsum([0] + lengths)

    # Perform PCA on foreground patches
    reduced_patches = object_pca.fit_transform(fg_patches)
    reduced_patches = minmax_scale(reduced_patches)

    return reduced_patches, mask_indices


def perform_pca_and_plot(
    images: List[np.ndarray],
    output_patch: np.ndarray,
    save_path: str,
    plot_title: Optional[str] = None,
    threshold: float = 0.0,
) -> None:
    """
    Perform PCA analysis on the model outputs focusing on foreground patches and generate plots,
    saving them to a specified path.

    Parameters:
    - images: List of original images.
    - output_patch: Model output patches for PCA analysis.
    - save_path: Path to save the plot.
    - plot_title: Title for the plot.
    - threshold: Threshold for masking in PCA analysis.
    """
    num_images = len(images)

    # Compute foreground masks
    masks = compute_foreground_masks(output_patch, threshold)

    # Perform PCA on foreground patches
    reduced_patches, mask_indices = perform_pca_on_foreground_patches(output_patch, masks)

    # Plotting
    ncols = 3  # Number of columns in the subplot grid
    plt.figure(figsize=(5 * ncols, 5 * num_images))

    for i in range(num_images):
        # Create a placeholder for the patches
        num_patches = output_patch.shape[1]
        patch_image = np.zeros((num_patches, 3), dtype="float32")

        # Fill in the foreground patches
        start_idx = mask_indices[i]
        end_idx = mask_indices[i + 1]
        patch_image[masks[i], :] = reduced_patches[start_idx:end_idx, :]

        # Reshape to the original image size
        patch_size = int(np.sqrt(num_patches))
        color_patches = patch_image.reshape(patch_size, patch_size, 3).transpose(1, 0, 2)

        # Plot the PCA visualization
        plt.subplot(num_images, ncols, ncols * i + 1)
        plt.imshow(color_patches)
        if plot_title:
            plt.title(f"{plot_title} - PCA Visualization", fontsize=10)
        plt.axis("off")

        # Plot the original image
        plt.subplot(num_images, ncols, ncols * i + 2)
        plt.imshow(images[i], cmap="gray")
        if plot_title:
            plt.title(f"{plot_title} - Original Image", fontsize=10)
        plt.axis("off")

        # Plot the foreground mask overlayed on the original image
        mask_image = masks[i].reshape(patch_size, patch_size).transpose(1, 0)
        plt.subplot(num_images, ncols, ncols * i + 3)
        plt.imshow(images[i], alpha=0.9)
        plt.imshow(mask_image, cmap="gray", alpha=0.5)
        if plot_title:
            plt.title(f"{plot_title} - Foreground Mask", fontsize=10)
        plt.axis("off")

    plt.tight_layout()
    plt.savefig(save_path)
    plt.close()


# Example usage
if __name__ == "__main__":
    iterations, dinov2_model = load_dino_model(
        "/processing/something/teacher_checkpoint.pth",
        "something/dinov2/dinov2/configs/train/mri_vits14_no_pretrain",
        224,
    )
    image_paths = [f"/image{i}.jpg" for i in range(1, 5)]

    device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
    images, preprocessed_images = load_and_preprocess_images(image_paths, size=(224, 224), channels=1)
    dinov2_model = dinov2_model.to(device)
    preprocessed_images = preprocessed_images.to(device)
    output_cls, output_patch = evaluate_model(dinov2_model, preprocessed_images)
    perform_pca_and_plot(images, output_patch, "pca.png", threshold=0.45, plot_title="fomo")
