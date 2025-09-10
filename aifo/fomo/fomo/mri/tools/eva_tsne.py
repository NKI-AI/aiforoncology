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
from sklearn.manifold import TSNE
import os
import matplotlib.pyplot as plt
import torch
import argparse
from matplotlib import cm
import time
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


# Parse command-line arguments
parser = argparse.ArgumentParser(description="Perform t-SNE on embeddings.")
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
    "--perplexity",
    type=float,
    default=30.0,
    help="Perplexity parameter for t-SNE (default: 30.0)",
)
parser.add_argument(
    "--learning_rate",
    type=float,
    default=200.0,
    help="Learning rate for t-SNE (default: 200.0)",
)
parser.add_argument(
    "--iterations",
    type=int,
    default=1000,
    help="Maximum number of iterations for t-SNE (default: 1000)",
)
parser.add_argument(
    "--dimensions",
    type=int,
    default=2,
    choices=[2, 3],
    help="Number of dimensions for t-SNE (default: 2)",
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
if args.middle_slices_only:
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
    sample_idx = row["sample_index"]
    slice_idx = row["slice_index"]

    # Store patient ID if it exists
    if has_patient_id:
        patient_ids.append(row["patient_id"])

    # Store metadata for coloring if requested
    if args.color_by and args.color_by in manifest.columns:
        metadata_values[args.color_by].append(row[args.color_by])

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

        # Handle different tensor shapes appropriately
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
    except Exception as e:
        logger.error(f"Error loading {full_embedding_path}: {str(e)}")

if not embeddings_list:
    raise ValueError("No valid embeddings found. Check your embeddings directory and manifest file.")

# Stack all embeddings into a single array
all_embeddings = np.vstack(embeddings_list)
logger.info(f"Final embeddings matrix shape for t-SNE: {all_embeddings.shape}")

# Perform t-SNE with specified parameters
logger.info(
    f"Running t-SNE with perplexity={args.perplexity}, learning_rate={args.learning_rate}, "
    f"iterations={args.iterations}, dimensions={args.dimensions}..."
)
start_time = time.time()
tsne = TSNE(
    n_components=args.dimensions,
    perplexity=args.perplexity,
    learning_rate=args.learning_rate,
    n_iter=args.iterations,
    random_state=42,
    verbose=1,
)
tsne_result = tsne.fit_transform(all_embeddings)
end_time = time.time()
logger.info(f"t-SNE completed in {end_time - start_time:.2f} seconds")

# Create DataFrame with t-SNE results and sample/slice information
if args.dimensions == 2:
    tsne_df = pd.DataFrame(tsne_result, columns=["TSNE1", "TSNE2"])
else:  # 3D
    tsne_df = pd.DataFrame(tsne_result, columns=["TSNE1", "TSNE2", "TSNE3"])

tsne_df["sample_index"] = [idx[0] for idx in sample_slice_indices]
tsne_df["slice_index"] = [idx[1] for idx in sample_slice_indices]

# Add patient_id column if it exists
if has_patient_id:
    tsne_df["patient_id"] = patient_ids

# Add color_by column to t-SNE dataframe if it exists
if args.color_by and args.color_by in manifest.columns:
    tsne_df[args.color_by] = metadata_values[args.color_by]
    logger.info(
        f"Added '{args.color_by}' to t-SNE results with {len(set(metadata_values[args.color_by]))} unique values"
    )

# Add any target column if it exists in the manifest and was specified by the user
is_classification = False
output_label = f"tsne{args.dimensions}d"  # Default output label if no target is specified

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

    tsne_df[args.target] = target_values

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
    output_label = f"{args.target}_tsne{args.dimensions}d"
elif args.target is not None:
    logger.warning(
        f"Warning: Specified target '{args.target}' not found in manifest columns: {manifest.columns.tolist()}"
    )
else:
    logger.info("No target specified. Using default visualization without target coloring.")

# Override output label if explicitly provided
if args.output_label:
    output_label = f"{args.output_label}_tsne{args.dimensions}d"

# Ensure the output directory exists
output_dir = "/home/v.v.veldhuizen/aiforoncology-internal/output/TSNE/"
os.makedirs(output_dir, exist_ok=True)

# Add max_patients to the filename if specified
if args.max_patients is not None and has_patient_id:
    mode = f"{mode}_top{args.max_patients}patients"

# Save to a CSV file
output_path = os.path.join(output_dir, f"{file_identifier}_{output_label}_{mode}.csv")
tsne_df.to_csv(output_path, index=False)
logger.info(f"t-SNE results saved to CSV: {output_path}")

# Visualize the t-SNE results - set up figure based on dimensions
if args.dimensions == 2:
    plt.figure(figsize=(10, 8))
else:  # 3D
    fig = plt.figure(figsize=(12, 10))
    ax = fig.add_subplot(111, projection="3d")

# Determine what to color by
if args.color_by and args.color_by in tsne_df.columns:
    color_by = args.color_by
    color_label = args.color_by.replace("_", " ").title()
    logger.info(f"Coloring plot by {color_by} as requested ({len(tsne_df[color_by].unique())} unique values)")
else:
    if args.color_by:
        logger.warning(
            f"Warning: Requested color column '{args.color_by}' not found. "
            f"Available columns: {tsne_df.columns.tolist()}"
        )

    # Default coloring logic
    if has_patient_id and len(tsne_df["patient_id"].unique()) <= 20:
        color_by = "patient_id"
        color_label = "Patient"
        logger.info(f"Coloring plot by patient_id ({len(tsne_df['patient_id'].unique())} unique patients)")
    else:
        color_by = "sample_index"
        color_label = "Volume"
        logger.info(f"Coloring plot by sample_index ({len(tsne_df['sample_index'].unique())} unique volumes)")

# Determine if we should show the legend based on number of unique values
show_legend = len(tsne_df[color_by].unique()) <= 20

# When getting unique identifiers for coloring, filter out blank and NaN values
unique_values = tsne_df[color_by].unique()

# Remove NaN and empty string values
valid_mask = ~(pd.isna(unique_values) | (unique_values == ""))
valid_values = unique_values[valid_mask]

# Sort the values (all should be of the same type now)
try:
    unique_identifiers = sorted(valid_values)
except TypeError:
    # If we still have mixed types, use string representation for sorting
    unique_identifiers = sorted(valid_values, key=str)

num_identifiers = len(unique_identifiers)
logger.info(f"Found {num_identifiers} valid unique values for {color_by} (excluding blanks/NaN)")

# Generate distinct colors for each unique identifier
colors = get_distinct_colors(num_identifiers)

# Create a mapping from identifier to color
color_map = dict(zip(unique_identifiers, colors))

# Color points by the chosen identifier, skipping blank/NaN values
for i, identifier in enumerate(unique_identifiers):
    # Only plot points with this exact identifier
    subset = tsne_df[tsne_df[color_by] == identifier]
    label = f"{color_label} {identifier}"

    if len(subset) > 0:  # Only try to plot if there are points
        if args.dimensions == 2:
            plt.scatter(
                subset["TSNE1"],
                subset["TSNE2"],
                color=color_map[identifier],
                label=label,
                alpha=0.7,
            )
        else:  # 3D
            ax.scatter(
                subset["TSNE1"],
                subset["TSNE2"],
                subset["TSNE3"],
                color=color_map[identifier],
                label=label,
                alpha=0.7,
            )

# Add a note that blank values are excluded
if np.any(pd.isna(tsne_df[color_by]) | (tsne_df[color_by] == "")):
    excluded_count = np.sum(pd.isna(tsne_df[color_by]) | (tsne_df[color_by] == ""))
    logger.info(f"Excluded {excluded_count} points with blank or NaN {color_by} values from the visualization")

# Set labels and title
if args.dimensions == 2:
    plt.xlabel("t-SNE Component 1")
    plt.ylabel("t-SNE Component 2")
    plt.title(
        f"t-SNE ({args.dimensions}D) of {mode.replace('_', ' ').title()} Embeddings\n"
        f"{model_name} - {output_dir_label} - {dataset_name}"
    )
else:  # 3D
    ax.set_xlabel("t-SNE Component 1")
    ax.set_ylabel("t-SNE Component 2")
    ax.set_zlabel("t-SNE Component 3")
    ax.set_title(
        f"t-SNE ({args.dimensions}D) of {mode.replace('_', ' ').title()} Embeddings\n"
        f"{model_name} - {output_dir_label} - {dataset_name}"
    )

# If too many unique identifiers, don't show the legend as it would be too cluttered
if not show_legend:
    if args.dimensions == 2:
        plt.legend().set_visible(False)
    else:  # 3D
        ax.legend().set_visible(False)
    logger.info(f"Too many unique {color_by}s, hiding legend in plot")
else:
    # Add a legend with a reasonable number of columns based on the number of identifiers
    if args.dimensions == 2:
        if len(unique_identifiers) > 10:
            plt.legend(ncol=2, loc="best", fontsize="small")
        else:
            plt.legend(loc="best")
    else:  # 3D
        if len(unique_identifiers) > 10:
            ax.legend(ncol=2, loc="best", fontsize="small")
        else:
            ax.legend(loc="best")

# Add the chosen color method to the filename
figure_path = os.path.join(output_dir, f"{file_identifier}_tsne{args.dimensions}d_by_{color_by}_{mode}.png")

# Save the figure
plt.savefig(figure_path, dpi=150, bbox_inches="tight")
logger.info(f"{color_by}-colored t-SNE plot saved to: {figure_path}")

# For 3D plots, create additional views from different angles
if args.dimensions == 3:
    # Save a few different views for 3D plots
    for angle_elevation in [0, 30, 60]:
        for angle_azimuth in [0, 45, 90, 135, 180, 225, 270, 315]:
            ax.view_init(elev=angle_elevation, azim=angle_azimuth)
            angle_path = os.path.join(
                output_dir,
                f"{file_identifier}_tsne3d_by_{color_by}_{mode}_elev{angle_elevation}_azim{angle_azimuth}.png",
            )
            plt.savefig(angle_path, dpi=150, bbox_inches="tight")

    logger.info("Saved multiple angle views for 3D t-SNE")

plt.close()

# If target column exists and appears to be a classification, create a plot colored by target
if args.target is not None and args.target in tsne_df.columns and is_classification:
    # Create a new figure
    if args.dimensions == 2:
        plt.figure(figsize=(10, 8))
    else:  # 3D
        fig = plt.figure(figsize=(12, 10))
        ax = fig.add_subplot(111, projection="3d")

    # Get unique target values and sort them
    unique_targets = sorted([val for val in tsne_df[args.target].unique() if not pd.isna(val)])
    num_targets = len(unique_targets)

    # Generate colors for the target values
    target_colors = get_distinct_colors(num_targets)
    target_color_map = dict(zip(unique_targets, target_colors))

    for target_value in unique_targets:
        subset = tsne_df[tsne_df[args.target] == target_value]
        if len(subset) > 0:  # Only try to plot if there are points
            if args.dimensions == 2:
                plt.scatter(
                    subset["TSNE1"],
                    subset["TSNE2"],
                    color=target_color_map[target_value],
                    label=f"{args.target.capitalize()} {target_value}",
                    alpha=0.7,
                )
            else:  # 3D
                ax.scatter(
                    subset["TSNE1"],
                    subset["TSNE2"],
                    subset["TSNE3"],
                    color=target_color_map[target_value],
                    label=f"{args.target.capitalize()} {target_value}",
                    alpha=0.7,
                )

    # Set labels and title
    if args.dimensions == 2:
        plt.xlabel("t-SNE Component 1")
        plt.ylabel("t-SNE Component 2")
        plt.title(
            f"t-SNE ({args.dimensions}D) of {mode.replace('_', ' ').title()} Embeddings\n"
            f"{model_name} - {output_dir_label} - {dataset_name} "
            f"(by {args.target.capitalize()})"
        )
        plt.legend()
    else:  # 3D
        ax.set_xlabel("t-SNE Component 1")
        ax.set_ylabel("t-SNE Component 2")
        ax.set_zlabel("t-SNE Component 3")
        ax.set_title(
            f"t-SNE ({args.dimensions}D) of {mode.replace('_', ' ').title()} Embeddings\n"
            f"{model_name} - {output_dir_label} - {dataset_name} "
            f"(by {args.target.capitalize()})"
        )
        ax.legend()

    # Save the figure
    target_figure_path = os.path.join(output_dir, f"{file_identifier}_tsne{args.dimensions}d_{args.target}_{mode}.png")
    plt.savefig(target_figure_path, dpi=150, bbox_inches="tight")
    logger.info(f"Target-colored t-SNE plot saved to: {target_figure_path}")

    # For 3D plots, create additional views from different angles
    if args.dimensions == 3:
        # Save a few different views for 3D plots
        for angle_elevation in [0, 30, 60]:
            for angle_azimuth in [0, 45, 90, 135, 180, 225, 270, 315]:
                ax.view_init(elev=angle_elevation, azim=angle_azimuth)
                angle_path = os.path.join(
                    output_dir,
                    f"{file_identifier}_tsne3d_{args.target}_{mode}_elev{angle_elevation}_azim{angle_azimuth}.png",
                )
                plt.savefig(angle_path, dpi=150, bbox_inches="tight")

        logger.info("Saved multiple angle views for 3D t-SNE target visualization")

    plt.close()
else:
    if args.target is not None:
        logger.info(f"Skipping target visualization as '{args.target}' does not appear to be a classification target.")
    else:
        logger.info("Skipping target visualization (no target specified).")

logger.info(f"Analysis complete. CSV and t-SNE plots saved to {output_dir}")
