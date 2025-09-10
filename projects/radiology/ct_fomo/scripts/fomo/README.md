# CT Foundation Models for Oncology (CT-FOMO)

A framework for training foundation models on CT medical imaging data using self-supervised learning.

## Overview

CT-FOMO applies self-supervised learning techniques to CT (Computed Tomography) medical images, with the goal of creating robust foundation models that can be fine-tuned for various downstream tasks in medical imaging analysis and oncology.

The project currently implements and alters the [DinoV2](https://github.com/facebookresearch/dinov2) framework for both 2D slice-based and 3D volume-based approaches.

## Features

- Self-supervised training on CT medical images
- Support for both 2D and 3D model variants

## Usage

To train a model, run one of the training scripts with a database URL containing CT image metadata:

```bash
sbatch scripts/2d/ct_vitb_16_no_pretrain.sh <DATABASE_URL> [PROJECT_ROOT]
```

Output will be saved to the `$SCRATCH` directory with a timestamp. Modification of the slurm scripts is required if using a
different cluster. Current scripts are specific to the NKI-AI cluster.

## Acknowledgements

This project builds upon the [DinoV2](https://github.com/facebookresearch/dinov2) framework for self-supervised learning. This project is currently being maintained by Tim Veenboer.
