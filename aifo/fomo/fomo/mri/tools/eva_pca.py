# Copyright 2025 AI for Oncology Research Group. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import pandas as pd
import numpy as np
from sklearn.decomposition import PCA
import os
import matplotlib.pyplot as plt
import torch
import argparse
from matplotlib import cm
import SimpleITK as sitk
import logging

logger = logging.getLogger(__name__)


# Define a custom color palette with 20 distinct colors
def get_distinct_colors(n):
    """Generate a colormap with n distinct colors."""
    if n <= 10:
        # Use the Tab10 colormap for up to 10 colors
        cmap = cm.get_cmap("tab10", 10)
        return [cmap(i) for i in range(n)]
    elif n <= 20:
        # Combine Tab10 and Set3 for up to 20 colors
        tab10 = cm.get_cmap("tab10", 10)
        set3 = cm.get_cmap("Set3", 12)
        return [tab10(i) for i in range(10)] + [set3(i) for i in range(n - 10)]
    else:
        # For more than 20, use a continuous colormap but select distinct points
        cmap = cm.get_cmap("viridis")
        return [cmap(i / (n - 1)) for i in range(n)]


# Function to check if a mask contains any non-zero pixels
def check_mask_has_pixels(mask_path):
    """
    Check if a mask file contains any non-zero pixels.

    Args:
        mask_path: Path to the mask file (.pt format)

    Returns:
        bool: True if the mask contains any non-zero pixels, False otherwise
    """
    try:
        # Load the mask tensor
        mask = torch.load(mask_path, map_location="cpu")

        # Check if the mask contains any non-zero pixels
        if isinstance(mask, torch.Tensor):
            return torch.any(mask > 0).item()
        else:
            logger.warning(f"Mask at {mask_path} is not a tensor")
            return False
    except Exception as e:
        logger.error(f"Error loading mask at {mask_path}: {str(e)}")
        return False


# Parse command-line arguments
parser = argparse.ArgumentParser(description="Perform PCA on embeddings.")
parser.add_argument(
    "--embeddings_dir",
    type=str,
    required=True,
    help="Directory containing precomputed embeddings and manifest",
)
parser.add_argument(
    "--manifest",
    type=str,
    default="manifest.csv",
    help="Name of the manifest file (default: manifest.csv)",
)
parser.add_argument(
    "--target",
    type=str,
    help="Target column name (if not specified, assumes segmentation with no target visualization)",
)
parser.add_argument("--output_label", type=str, help="Custom label name for the output")
parser.add_argument(
    "--max_patients",
    type=int,
    help="Maximum number of unique patient_ids to include (default: all)",
)
parser.add_argument(
    "--middle_slices_only",
    action="store_true",
    help="Only include middle slice from each sample/volume",
)
parser.add_argument(
    "--color_by",
    type=str,
    help="Column name to use for coloring the plots (default: patient_id if available, otherwise sample_index)",
)
parser.add_argument(
    "--plot_pc34",
    action="store_true",
    help="Additionally plot PC3 vs PC4 in a separate plot",
)
parser.add_argument(
    "--mil_mode",
    action="store_true",
    help="Enable MIL mode for handling embeddings that contain lists of tensors",
)
parser.add_argument("--save_csv", action="store_true", help="Save PCA results to CSV file")
parser.add_argument(
    "--mask_base_path",
    type=str,
    default="/data/groups/aiforoncology/archive/radiology/mama-mia/segmentations/expert/",
    help="Base path for mask files if they are relative paths in the target column",
)
parser.add_argument(
    "--output_dir",
    type=str,
    default="./output/PCA/",
    help="Output directory for PCA results and plots (default: ./output/PCA/)",
)
args = parser.parse_args()

# Determine model name, output label, and dataset from the embeddings directory path
embeddings_dir = args.embeddings_dir
# Split the path into components
path_components = embeddings_dir.split("/")
# Clean empty components
path_components = [p for p in path_components if p]

# Look for 'embeddings' in the path to extract the relevant portions
model_name = None
output_dir_label = None
dataset_name = None

if "embeddings" in path_components:
    embeddings_idx = path_components.index("embeddings")

    # Extract model name (component after 'embeddings')
    if len(path_components) > embeddings_idx + 1:
        model_name = path_components[embeddings_idx + 1]

        # Extract output label (component after model name)
        if len(path_components) > embeddings_idx + 2:
            output_dir_label = path_components[embeddings_idx + 2]

            # Extract dataset name (component after output label)
            if len(path_components) > embeddings_idx + 3:
                dataset_name = path_components[embeddings_idx + 3]
else:
    # Fallback: use the last components if 'embeddings' isn't found
    if len(path_components) >= 3:
        model_name = path_components[-3]
        output_dir_label = path_components[-2]
        dataset_name = path_components[-1]
    elif len(path_components) >= 2:
        model_name = path_components[-2]
        output_dir_label = path_components[-1]
    else:
        model_name = path_components[-1]

# Construct a comprehensive identifier for filenames
file_identifier = model_name
if output_dir_label:
    file_identifier += f"_{output_dir_label}"
if dataset_name:
    file_identifier += f"_{dataset_name}"

logger.info(f"Using model: {model_name}")
if output_dir_label:
    logger.info(f"Output label: {output_dir_label}")
if dataset_name:
    logger.info(f"Dataset: {dataset_name}")
logger.info(f"File identifier: {file_identifier}")

# Set mode to a fixed string for filename purposes, adjusted if middle_slices_only is set
mode = "middle_slices" if args.middle_slices_only else "all_slices"

# Check if we should use MIL mode based on path or explicit flag
use_mil_mode = args.mil_mode or "/mil" in embeddings_dir
if use_mil_mode:
    logger.info("Using MIL mode: treating each .pt file as a list of CLS embeddings (one per volume)")
    mode = f"{mode}_mil"

# Load the manifest file from the embeddings directory
manifest_path = os.path.join(embeddings_dir, args.manifest)
logger.info(f"Loading manifest from: {manifest_path}")

# Determine the delimiter by checking the first line of the file
with open(manifest_path, "r") as file:
    first_line = file.readline()
delimiter = "," if "," in first_line else ";"

# Load the manifest file with the determined delimiter
manifest = pd.read_csv(manifest_path, delimiter=delimiter)

# Print manifest columns to verify structure
logger.info(f"Manifest columns: {manifest.columns.tolist()}")

# Check if required columns exist
required_columns = ["sample_index", "slice_index"]
if use_mil_mode:
    # In MIL mode, we need multi_id instead of sample_index
    required_columns = ["multi_id", "slice_index"]
    logger.info("MIL mode: using 'multi_id' instead of 'sample_index' for sample identification")
for col in required_columns:
    if col not in manifest.columns:
        raise ValueError(f"Required column '{col}' not found in manifest file")

# Check if patient_id column exists for patient-based filtering
has_patient_id = "patient_id" in manifest.columns
if has_patient_id:
    logger.info(f"Found patient_id column with {manifest['patient_id'].nunique()} unique patients")

    # If max_patients is specified, filter the manifest to include only that many patients
    if args.max_patients is not None and args.max_patients > 0:
        unique_patients = manifest["patient_id"].unique()

        if len(unique_patients) > args.max_patients:
            # Limit to the first max_patients unique patient IDs
            selected_patients = unique_patients[: args.max_patients]
            manifest = manifest[manifest["patient_id"].isin(selected_patients)]
            logger.info(f"Limited to {args.max_patients} patients: {selected_patients}")
        else:
            logger.info(
                f"Requested {args.max_patients} patients, but only {len(unique_patients)} available. Using all."
            )
else:
    logger.info("No patient_id column found in manifest. Will use sample_index for visualization.")

# If middle_slices_only is specified, filter the manifest to include only middle slices
if args.middle_slices_only and not use_mil_mode:
    logger.info("Middle slices only mode: filtering manifest to include only middle slice for each sample")

    # For each unique sample, find the middle slice
    middle_slice_rows = []
    for sample_idx in manifest["sample_index"].unique():
        sample_slices = manifest[manifest["sample_index"] == sample_idx]

        # Find the middle slice index
        slice_indices = sorted(sample_slices["slice_index"].unique())
        if len(slice_indices) > 0:
            middle_idx = len(slice_indices) // 2
            middle_slice_idx = slice_indices[middle_idx]

            # Get the row for the middle slice
            middle_slice_row = sample_slices[sample_slices["slice_index"] == middle_slice_idx]
            middle_slice_rows.append(middle_slice_row)

    if middle_slice_rows:
        # Combine all middle slice rows
        manifest = pd.concat(middle_slice_rows)
        logger.info(f"Filtered to {len(manifest)} middle slices (one per volume)")
    else:
        raise ValueError("No middle slices found after filtering")

# Check if embeddings column exists
if "embeddings" not in manifest.columns:
    logger.warning(
        "Warning: 'embeddings' column not found in manifest. Will construct filenames using sample_index and slice_index."
    )
    use_embeddings_column = False
else:
    use_embeddings_column = True
    logger.info("Found 'embeddings' column in manifest. Using these filenames to load embeddings.")

# Add mask column if color_by is 'mask' and target column exists
if args.color_by == "mask" and "target" in manifest.columns:
    logger.info("Adding mask column to indicate tumor presence...")
    mask_values = []

    # Cache for loaded masks
    mask_cache = {}

    for idx, row in manifest.iterrows():
        # Extract the base filename from the embeddings path
        embedding_filename = row["embeddings"]
        base_filename = embedding_filename.split("-")[0]  # e.g., "DUKE_001_0001.nii" from "DUKE_001_0001.nii-0.pt"
        base_filename = base_filename.replace(".nii", "")  # Remove .nii if it exists

        # Extract just the patient ID (e.g., "DUKE_001" from "DUKE_001_0001")
        patient_id = "_".join(base_filename.split("_")[:2])

        # Extract slice index from embedding filename
        slice_idx = int(embedding_filename.split("-")[1].split(".")[0])  # e.g., "0" from "DUKE_001_0001.nii-0.pt"

        # Check if we already have this mask in cache
        if patient_id not in mask_cache:
            # Construct the full mask path
            if args.mask_base_path:
                mask_path = os.path.join(args.mask_base_path, f"{patient_id}.nii.gz")
            else:
                mask_path = (
                    f"/data/groups/aiforoncology/archive/radiology/mama-mia/segmentations/expert/{patient_id}.nii.gz"
                )

            logger.info(f"Loading mask for patient {patient_id}...")

            # Load the 3D mask and cache it
            try:
                mask_3d = sitk.ReadImage(mask_path)
                mask_array = sitk.GetArrayFromImage(mask_3d)
                mask_cache[patient_id] = mask_array
            except Exception as e:
                logger.error(f"Error loading mask for {patient_id}: {str(e)}")
                mask_cache[patient_id] = None

        # Get the mask from cache
        mask_array = mask_cache[patient_id]
        if mask_array is None:
            mask_values.append(0)  # Default to no tumor if mask couldn't be loaded
            continue

        # Find which dimension is the depth (the one that's different from the other two)
        dims = mask_array.shape
        depth_dim = np.argmax([abs(dims[0] - dims[1]), abs(dims[1] - dims[2]), abs(dims[0] - dims[2])])

        # Extract the slice
        if depth_dim == 0:
            mask_slice = mask_array[slice_idx, :, :]
        elif depth_dim == 1:
            mask_slice = mask_array[:, slice_idx, :]
        else:  # depth_dim == 2
            mask_slice = mask_array[:, :, slice_idx]

        # Convert to tensor and check for non-zero pixels
        mask_tensor = torch.from_numpy(mask_slice)
        has_tumor = torch.any(mask_tensor > 0).item()
        mask_values.append(1 if has_tumor else 0)

    # Add the mask column to the manifest
    manifest["mask"] = mask_values

    # Print statistics about tumor presence
    tumor_count = sum(mask_values)
    logger.info(
        f"Found {tumor_count} tumor-positive slices out of {len(mask_values)} "
        f"total slices ({tumor_count / len(mask_values) * 100:.2f}%)"
    )

# Load the embeddings
embeddings_list = []
sample_slice_indices = []  # To keep track of which sample and slice each embedding comes from
patient_ids = []  # To keep track of patient IDs if they exist
metadata_values = {}  # Track values for the color_by column if it exists

# Check if color_by column exists in manifest
if args.color_by and args.color_by in manifest.columns:
    metadata_values[args.color_by] = []
    logger.info(f"Will use '{args.color_by}' column for coloring")

# Loop through all entries in the manifest
for _, row in manifest.iterrows():
    # Use multi_id in MIL mode, otherwise use sample_index
    sample_idx = row["multi_id"] if use_mil_mode else row["sample_index"]
    slice_idx = row["slice_index"]

    # Determine the embedding file path
    if use_embeddings_column:
        # Use the filename from the embeddings column
        embedding_filename = row["embeddings"]
        full_embedding_path = os.path.join(embeddings_dir, embedding_filename)
    else:
        # Construct the filename using sample_index and slice_index
        embedding_filename = f"embedding_{sample_idx}_{slice_idx}.pt"
        full_embedding_path = os.path.join(embeddings_dir, embedding_filename)

    # Check if file exists
    if not os.path.exists(full_embedding_path):
        logger.warning(f"Warning: Embedding file {full_embedding_path} not found, skipping")
        continue

    # Load the embedding
    try:
        tensor = torch.load(full_embedding_path, map_location="cpu")

        # Handle lists containing tensors
        if isinstance(tensor, list):
            if len(tensor) == 1:
                tensor = tensor[0]  # Extract the tensor from the list
            else:
                logger.warning(
                    f"Warning: List of length {len(tensor)} found at {full_embedding_path}, using first element"
                )
                tensor = tensor[0]

        # Handle MIL mode (list of CLS embeddings)
        if use_mil_mode:
            if isinstance(tensor, list):
                # Process each CLS embedding in the list
                for i, cls_tensor in enumerate(tensor):
                    # For CLS embeddings, we don't need spatial averaging
                    # Just convert to numpy and ensure it's a 1D vector
                    embeddings = cls_tensor.flatten().numpy().astype(np.float32)

                    if np.isnan(embeddings).any():
                        logger.warning(
                            f"Warning: NaN values found in embeddings at {full_embedding_path} (slice {i}), skipping"
                        )
                        continue

                    embeddings_list.append(embeddings)
                    sample_slice_indices.append((sample_idx, i))  # Use i as slice index for MIL mode

                    # For MIL mode, we need to duplicate the metadata for each slice
                    if has_patient_id:
                        patient_ids.append(row["patient_id"])
                    if args.color_by and args.color_by in manifest.columns:
                        metadata_values[args.color_by].append(row[args.color_by])
            else:
                logger.warning(
                    f"Warning: Expected list of tensors in MIL mode, but got {type(tensor)} at {full_embedding_path}, "
                    f"skipping"
                )
                continue
        else:
            # Handle different tensor shapes appropriately (non-MIL mode)
            if len(tensor.shape) == 3:  # Shape like [384, 25, 36]
                # Average over spatial dimensions to get a consistent vector
                embeddings = tensor.mean(dim=[1, 2]).numpy().astype(np.float32)
            elif len(tensor.shape) == 2:  # Shape like [384, N]
                # Average over sequence length
                embeddings = tensor.mean(dim=1).numpy().astype(np.float32)
            else:
                # If it's already a 1D vector or other shape, flatten it
                embeddings = tensor.flatten().numpy().astype(np.float32)

            if np.isnan(embeddings).any():
                logger.warning(f"Warning: NaN values found in embeddings at {full_embedding_path}, skipping")
                continue

            embeddings_list.append(embeddings)
            sample_slice_indices.append((sample_idx, slice_idx))

            # Store patient ID if it exists
            if has_patient_id:
                patient_ids.append(row["patient_id"])

            # Store metadata for coloring if requested
            if args.color_by and args.color_by in manifest.columns:
                metadata_values[args.color_by].append(row[args.color_by])
    except Exception as e:
        logger.error(f"Error loading {full_embedding_path}: {str(e)}")

if not embeddings_list:
    raise ValueError("No valid embeddings found. Check your embeddings directory and manifest file.")

# Verify that all lists have the same length
logger.info(f"Number of embeddings: {len(embeddings_list)}")
logger.info(f"Number of sample/slice indices: {len(sample_slice_indices)}")
if has_patient_id:
    logger.info(f"Number of patient IDs: {len(patient_ids)}")
    if len(patient_ids) != len(embeddings_list):
        logger.warning(
            f"Warning: Length mismatch between patient_ids ({len(patient_ids)}) and "
            f"embeddings_list ({len(embeddings_list)})"
        )
        # Truncate or extend patient_ids to match embeddings_list length
        if len(patient_ids) > len(embeddings_list):
            patient_ids = patient_ids[: len(embeddings_list)]
        else:
            # If we have fewer patient IDs than embeddings, duplicate the last ones
            while len(patient_ids) < len(embeddings_list):
                patient_ids.append(patient_ids[-1] if patient_ids else None)
        logger.info(f"Adjusted number of patient IDs: {len(patient_ids)}")

# Check and adjust metadata_values if needed
for key in metadata_values:
    if len(metadata_values[key]) != len(embeddings_list):
        logger.warning(
            f"Warning: Length mismatch between metadata_values['{key}'] ({len(metadata_values[key])}) and "
            f"embeddings_list ({len(embeddings_list)})"
        )
        # Truncate or extend metadata_values to match embeddings_list length
        if len(metadata_values[key]) > len(embeddings_list):
            metadata_values[key] = metadata_values[key][: len(embeddings_list)]
        else:
            # If we have fewer metadata values than embeddings, duplicate the last ones
            while len(metadata_values[key]) < len(embeddings_list):
                metadata_values[key].append(metadata_values[key][-1] if metadata_values[key] else None)
        logger.info(f"Adjusted number of metadata values for '{key}': {len(metadata_values[key])}")

# Stack all embeddings into a single array
all_embeddings = np.vstack(embeddings_list)
logger.info(f"Final embeddings matrix shape for PCA: {all_embeddings.shape}")

# Perform PCA
n_components = 4 if args.plot_pc34 else 2
pca = PCA(n_components=n_components)
pca_result = pca.fit_transform(all_embeddings)

# Create DataFrame with PCA results and sample/slice information
if args.plot_pc34:
    pca_df = pd.DataFrame(pca_result, columns=["PC1", "PC2", "PC3", "PC4"])
else:
    pca_df = pd.DataFrame(pca_result, columns=["PC1", "PC2"])

# Use multi_id in MIL mode, otherwise use sample_index
if use_mil_mode:
    pca_df["multi_id"] = [idx[0] for idx in sample_slice_indices]
else:
    pca_df["sample_index"] = [idx[0] for idx in sample_slice_indices]
pca_df["slice_index"] = [idx[1] for idx in sample_slice_indices]

# Add patient_id column if it exists
if has_patient_id:
    pca_df["patient_id"] = patient_ids

# Add color_by column to PCA dataframe if it exists
if args.color_by and args.color_by in manifest.columns:
    pca_df[args.color_by] = metadata_values[args.color_by]
    logger.info(f"Added '{args.color_by}' to PCA results with {len(set(metadata_values[args.color_by]))} unique values")

# Add any target column if it exists in the manifest and was specified by the user
is_classification = False
output_label = "visualization"  # Default output label if no target is specified

if args.target is not None and args.target in manifest.columns:
    # We need to match the target values to our loaded embeddings
    target_values = []
    for sample_idx, slice_idx in sample_slice_indices:
        # Find the corresponding row in the manifest
        match = manifest[(manifest["sample_index"] == sample_idx) & (manifest["slice_index"] == slice_idx)]
        if not match.empty:
            target_values.append(match[args.target].values[0])
        else:
            target_values.append(None)

    pca_df[args.target] = target_values

    # Determine if target is likely a classification (not a path)
    # Check unique values to determine if it's a classification target
    unique_targets = manifest[args.target].unique()

    # If we have a small number of unique values (typically < 10 for classification)
    # and they don't look like file paths, assume it's classification
    is_classification = len(unique_targets) < 10 and not any(
        [
            "/" in str(val) or "\\" in str(val) or ".nii" in str(val) or ".nrrd" in str(val)
            for val in unique_targets
            if isinstance(val, str)
        ]
    )

    logger.info(f"Target '{args.target}' has {len(unique_targets)} unique values.")
    logger.info(f"Detected as {'classification' if is_classification else 'segmentation mask paths'}")
    logger.info(f"Unique values: {unique_targets}")

    # Use the target name for output label if not otherwise specified
    output_label = args.target
elif args.target is not None:
    logger.warning(
        f"Warning: Specified target '{args.target}' not found in manifest columns: {manifest.columns.tolist()}"
    )
else:
    logger.info("No target specified. Assuming segmentation data (no target visualization).")

# Override output label if explicitly provided
if args.output_label:
    output_label = args.output_label

# Ensure the output directory exists
os.makedirs(args.output_dir, exist_ok=True)

# Add max_patients to the filename if specified
if args.max_patients is not None and has_patient_id:
    mode = f"{mode}_top{args.max_patients}patients"

# Save to a CSV file only if requested
if args.save_csv:
    output_path = os.path.join(args.output_dir, f"{file_identifier}_{output_label}_{mode}.csv")
    pca_df.to_csv(output_path, index=False)
    logger.info(f"PCA results saved to CSV: {output_path}")

# Visualize the PCA results
plt.figure(figsize=(10, 8))

# Determine what to color by
if args.color_by and args.color_by in pca_df.columns:
    color_by = args.color_by
    color_label = args.color_by.replace("_", " ").title()
    logger.info(f"Coloring plot by {color_by} as requested ({len(pca_df[color_by].unique())} unique values)")

    # Special handling for mask column
    if color_by == "mask" and has_patient_id:
        if len(pca_df["patient_id"].unique()) <= 20:
            color_by = "patient_id"  # We'll use patient_id for base colors
            color_label = "Patient"
            logger.info(f"Using patient_id with tumor markers ({len(pca_df['patient_id'].unique())} unique patients)")

            # Generate base colors for each patient
            base_colors = get_distinct_colors(len(pca_df["patient_id"].unique()))
            color_map = {}

            # Create color map for patients
            for i, patient_id in enumerate(sorted(pca_df["patient_id"].unique())):
                color = "#{:02x}{:02x}{:02x}".format(
                    int(base_colors[i][0] * 255),
                    int(base_colors[i][1] * 255),
                    int(base_colors[i][2] * 255),
                )
                color_map[patient_id] = color

            # Create combined identifiers for plotting
            unique_identifiers = []
            for patient_id in sorted(pca_df["patient_id"].unique()):
                unique_identifiers.append((patient_id, 0))  # No tumor
                unique_identifiers.append((patient_id, 1))  # Tumor
        else:
            # For more than 20 patients, use simple tumor/non-tumor coloring
            color_label = "Tumor Status"
            logger.info("More than 20 patients, using simple tumor/non-tumor coloring")

            # Use distinct colors for tumor and non-tumor
            tumor_colors = ["#3498db", "#e74c3c"]  # Blue for no tumor, Red for tumor
            color_map = {0: tumor_colors[0], 1: tumor_colors[1]}
            unique_identifiers = [0, 1]  # No tumor first, then tumor
    else:
        # For other columns, use standard coloring
        unique_values = pca_df[color_by].unique()
        valid_mask = ~(pd.isna(unique_values) | (unique_values == ""))
        valid_values = unique_values[valid_mask]
        unique_identifiers = sorted(valid_values)
        colors = get_distinct_colors(len(unique_identifiers))
        color_map = dict(zip(unique_identifiers, colors))
else:
    if args.color_by:
        logger.warning(
            f"Warning: Requested color column '{args.color_by}' not found. Available columns: {pca_df.columns.tolist()}"
        )

    # Default coloring logic
    if has_patient_id:
        color_by = "patient_id"
        color_label = "Patient"
        logger.info(f"Coloring plot by patient_id ({len(pca_df['patient_id'].unique())} unique patients)")

        # Generate base colors for each patient
        base_colors = get_distinct_colors(len(pca_df["patient_id"].unique()))
        color_map = {}

        # Create color map for patients
        for i, patient_id in enumerate(sorted(pca_df["patient_id"].unique())):
            color = "#{:02x}{:02x}{:02x}".format(
                int(base_colors[i][0] * 255),
                int(base_colors[i][1] * 255),
                int(base_colors[i][2] * 255),
            )
            color_map[patient_id] = color

        # Create combined identifiers for plotting
        unique_identifiers = []
        for patient_id in sorted(pca_df["patient_id"].unique()):
            unique_identifiers.append((patient_id, 0))  # No tumor
            unique_identifiers.append((patient_id, 1))  # Tumor
    else:
        # Use multi_id in MIL mode, otherwise use sample_index
        if use_mil_mode:
            color_by = "multi_id"
            color_label = "Volume"
            logger.info(f"Coloring plot by multi_id ({len(pca_df['multi_id'].unique())} unique volumes)")
        else:
            color_by = "sample_index"
            color_label = "Volume"
            logger.info(f"Coloring plot by sample_index ({len(pca_df['sample_index'].unique())} unique volumes)")

        # For standard coloring
        unique_values = pca_df[color_by].unique()
        valid_mask = ~(pd.isna(unique_values) | (unique_values == ""))
        valid_values = unique_values[valid_mask]
        unique_identifiers = sorted(valid_values)
        colors = get_distinct_colors(len(unique_identifiers))
        color_map = dict(zip(unique_identifiers, colors))

# Color points by the chosen identifier, skipping blank/NaN values
for i, identifier in enumerate(unique_identifiers):
    # Special handling for patient_id + tumor combination
    if isinstance(identifier, tuple):
        patient_id, tumor_status = identifier
        subset = pca_df[(pca_df["patient_id"] == patient_id) & (pca_df["mask"] == tumor_status)]
        label = f"Patient {patient_id} {'(Tumor)' if tumor_status else '(No Tumor)'}"
    else:
        subset = pca_df[pca_df[color_by] == identifier]
        if color_by == "mask":
            label = "No Tumor" if identifier == 0 else "Tumor"
        else:
            label = f"{color_label} {identifier}"

    plt.scatter(
        subset["PC1"],
        subset["PC2"],
        color=color_map[identifier[0] if isinstance(identifier, tuple) else identifier],
        label=label,
        alpha=0.7,
        marker="x" if isinstance(identifier, tuple) and identifier[1] == 1 else "o",
    )  # 'x' for tumor, 'o' for non-tumor

# Add a note that blank values are excluded
if np.any(pd.isna(pca_df[color_by]) | (pca_df[color_by] == "")):
    excluded_count = np.sum(pd.isna(pca_df[color_by]) | (pca_df[color_by] == ""))
    logger.info(f"Excluded {excluded_count} points with blank or NaN {color_by} values from the visualization")

plt.xlabel("Principal Component 1")
plt.ylabel("Principal Component 2")
plt.title(f"PCA of {mode.replace('_', ' ').title()} Embeddings\n{model_name} - {output_dir_label} - {dataset_name}")

# If too many unique identifiers, don't show the legend as it would be too cluttered
if len(unique_identifiers) > 7:
    plt.legend().set_visible(False)
    logger.info(f"Too many unique {color_by}s, hiding legend in volume plot")
else:
    plt.legend(loc="best")

# Add the chosen color method to the filename
figure_path = os.path.join(args.output_dir, f"{file_identifier}_by_{color_by}_{mode}.png")

# Save the figure
plt.savefig(figure_path, dpi=150, bbox_inches="tight")
logger.info(f"{color_by}-colored plot saved to: {figure_path}")
plt.close()

# Create PC3 vs PC4 plot if requested
if args.plot_pc34:
    plt.figure(figsize=(10, 8))

    # Color points by the chosen identifier, skipping blank/NaN values
    for i, identifier in enumerate(unique_identifiers):
        subset = pca_df[pca_df[color_by] == identifier]

        # Special handling for mask column
        if color_by == "mask":
            label = "No Tumor" if identifier == 0 else "Tumor"
        else:
            label = f"{color_label} {identifier}"

        plt.scatter(
            subset["PC3"],
            subset["PC4"],
            color=color_map[identifier],
            label=label,
            alpha=0.7,
        )

    plt.xlabel("Principal Component 3")
    plt.ylabel("Principal Component 4")
    plt.title(
        f"PCA (PC3 vs PC4) of {mode.replace('_', ' ').title()} "
        f"Embeddings\n{model_name} - {output_dir_label} - {dataset_name}"
    )

    # Handle legend visibility same as PC1 vs PC2 plot
    if len(unique_identifiers) > 7:
        plt.legend().set_visible(False)
    else:
        plt.legend(loc="best")

    # Save PC3 vs PC4 plot
    figure_path = os.path.join(args.output_dir, f"{file_identifier}_by_{color_by}_{mode}_pc34.png")
    plt.savefig(figure_path, dpi=150, bbox_inches="tight")
    logger.info(f"PC3 vs PC4 plot saved to: {figure_path}")
    plt.close()

    # If target column exists and appears to be classification, create PC3 vs PC4 plot colored by target
    if args.target is not None and args.target in pca_df.columns and is_classification:
        plt.figure(figsize=(10, 8))

        # Create a color map for target values
        target_color_map = {}
        for i, target_value in enumerate(unique_targets):
            target_color_map[target_value] = colors[i % len(colors)]

        for target_value in unique_targets:
            subset = pca_df[pca_df[args.target] == target_value]
            plt.scatter(
                subset["PC3"],
                subset["PC4"],
                color=target_color_map[target_value],
                label=f"{output_label.capitalize()} {target_value}",
                alpha=0.7,
            )

        plt.xlabel("Principal Component 3")
        plt.ylabel("Principal Component 4")
        plt.title(
            f"PCA (PC3 vs PC4) of {mode.replace('_', ' ').title()} "
            f"Embeddings\n{model_name} - {output_dir_label} - {dataset_name} "
            f"(by {output_label.capitalize()})"
        )
        plt.legend()

        # Save the figure
        figure_path = os.path.join(args.output_dir, f"{file_identifier}_{output_label}_{mode}_pc34.png")
        plt.savefig(figure_path, dpi=150, bbox_inches="tight")
        logger.info(f"Target-colored PC3 vs PC4 plot saved to: {figure_path}")
        plt.close()

logger.info(f"Analysis complete. CSV and plots saved to {args.output_dir}")
