# CT Foundation Model Evaluation Framework

This repository contains tools for evaluating foundation models trained on CT images, including both the core evaluation framework and specific downstream task evaluation scripts building upon Kaiko's EVA framework.

## Foundation Model Evaluation Framework (FEF)

The Foundation Model Evaluation Framework is a comprehensive toolkit for evaluating CT foundation models on various downstream tasks. The framework provides:

- **Modular Architecture**: Standardized interfaces for models, datasets, and evaluation metrics
- **Configurable Pipeline**: YAML-based configuration system for flexible experiment setup with executable SLURM scripts.

### Core Components

- **Models**: Implementation of backbone architectures including ViT variants
- **Datasets**: CT-specific dataset loaders with medical imaging transformations
- **Transforms**: Pre-processing pipelines for CT images
- **Configs**: YAML configuration files for different downstream tasks
- **Utils**: Helper functions and utilities for the evaluation pipeline

## FEF Evaluation Scripts

The FEF project provides scripts and tools to run standardized downstream evaluation tasks on trained CT foundation models. These scripts enable consistent evaluation and benchmarking of different foundation model architectures.

### Available Evaluation Tasks

#### Classification

- **Lesion Classification**: Scripts to evaluate models on identifying and classifying lesions in CT scans
- **Kidney Tumor Segmentation (KiTS23)**: Scripts to evaluate models on KiTS23 dataset.

### Running Evaluations

To run an evaluation task:

1. Change the scripts to be specific to your SLURM cluster
2. Set the appropriate environment variables
3. Ensure your foundation model checkpoint is available
4. Run the appropriate evaluation script:

```bash
# Example: Running lesion classification with a 3D ViT-Base model
./ct_fomo/scripts/fef/classification/lesion/vit_base_patch16_224_3d.sh [EVAL_NAME] [PATH_TO_CHECKPOINT] [PROJECT_ROOT]
```

Parameters:

- `EVAL_NAME`: A name for this evaluation run
- `PATH_TO_CHECKPOINT`: Path to the foundation model checkpoint
- `PROJECT_ROOT`: (Optional) Root directory of the project

## Requirements

See `requirements.in` for core dependencies. The framework supports both Linux and macOS platforms with separate requirements files.

## Acknowledgements

FEF builds upon the Oncology Foundation Model Evaluation Framework (EVA) by Kaiko. This project is being maintained by Vivien van Veldhuizen and Tim Veenboer.
