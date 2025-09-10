# AH Core Library

A comprehensive library for AI-powered image analysis, providing powerful tools for processing and analyzing pathology images.

## Overview

AH Core combines C++ and Python components to provide a high-performance framework for:

- Processing large whole-slide images (WSIs)
- Tiling and patching strategies for pathology images
- Data loading and augmentation pipelines
- Model training utilities with PyTorch integration
- Inference pipelines for trained models
- Visualization tools for segmentation results

## Components

- **Data Processing**: Tools for reading, writing, and transforming pathology images
- **Tiling**: Advanced tiling strategies for whole-slide images
- **Models**: Deep learning architectures for pathology tasks
- **Training**: PyTorch Lightning integration for efficient training
- **Visualization**: Tools to visualize results on slide images

## CLI Usage

The library provides a command-line interface for common operations:

### Using Bazel (Recommended)

```bash
# Get help for CLI commands
bazelisk run //aifo/llm_serveahcore:cli -- -h
```

### Dependencies

The project uses Bazel for building and dependency management:

```bash
# Build the entire library
bazelisk build //aifo/ahcore:all

# Build specific components
bazelisk build //aifo/ahcore:data
```

### Testing

```bash
# Run tests
bazelisk test //aifo/ahcore/tests/...
```

### Important environment variables

- `AHCORE_KEEP_TEMPORARY_FILES=1`: will keep the temporary `.v` files in the `//aifo/ahcore:inference` target

## License

Internal use only - AI for Oncology project
