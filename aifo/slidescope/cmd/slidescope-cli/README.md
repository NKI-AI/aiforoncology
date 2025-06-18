# SlideScope CLI

A unified command-line tool for managing SlideScope slides, masks, and users.

## Installation

```bash
bazelisk build //aifo/slidescope/cmd/slidescope-cli
```

## Commands

The SlideScope CLI offers several commands:

### Import Slides and Masks

Import slides and their associated masks from a YAML configuration file:

```bash
./bazel-bin/aifo/slidescope/cmd/slidescope-cli/slidescope-cli import -c example.yaml -s http://localhost:3000 -u admin
```

Options:

- `-c, --config`: Path to the YAML configuration file (required)
- `-s, --server`: URL of the SlideScope server (default: http://localhost:3000)
- `-u, --username`: Username for authentication (optional, will prompt if not provided)

This requires a running SlideScope instance as the data is added through the API.

### User Management

#### Add a new user:

```bash
./bazel-bin/aifo/slidescope/cmd/slidescope-cli/slidescope-cli user add username -d path/to/database.sqlite
```

Options:

- `-d, --database`: Path to the database (or set DATABASE_URL environment variable)
- `-c, --cost`: Bcrypt cost level for password hashing (default: 10)

#### Change a user's password:

```bash
./bazel-bin/aifo/slidescope/cmd/slidescope-cli/slidescope-cli user passwd username -d path/to/database.sqlite
```

Options:

- `-d, --database`: Path to the database (or set DATABASE_URL environment variable)
- `-c, --cost`: Bcrypt cost level for password hashing (default: 10)

## YAML Configuration Format

The YAML file for imports defines the slides and masks to import:

```yaml
# Example with a single mask
- image_name: "Slide 1"
  image_path: "image.svs"
  mask_path: "mask.tiff"

# Example with multiple masks
- image_name: "Slide 2"
  image_path: "image2.svs"
  mask_path:
    - "mask2.tiff"
    - "another_mask.tiff"
```

### Fields

- `image_name`: Display name for the slide
- `image_path`: Path to the slide file (.svs)
- `mask_path`: Path to the mask file (.tiff) or a list of paths

All relative paths are automatically resolved to absolute paths during import.
