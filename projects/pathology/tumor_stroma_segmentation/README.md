# Tumor Stroma Segmentation Project

_Autohor_: Ajey Pai Karkala

## Table of Contents

- [Overview](#overview)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Setup and Execution](#setup-and-execution)
  - [Running the Conversion Script](#running-the-conversion-script)
- [Contributions](#contributions)
- [License](#license)
- [Contact](#contact)

## Overview

In the tumor stroma project we have a pipeline that segments the tumor and stroma in a WSI.

## Project Structure

- **darwin_to_dlup_xml.py**: Script to convert darwin json files to dlup xml files
- **BUILD.bazel**: Bazel build configuration for running the conversion script
- **README.md**: Project documentation

## Prerequisites

- **Bazelisk**: A Bazel wrapper to simplify builds. Install from [Bazelisk GitHub](https://github.com/bazelbuild/bazelisk) or via your package manager.
- Darwin V7 annotated JSON files for input or dlup xml formatted

## Setup and Execution

### Running the Conversion Script

If you are using the darwin json files, follow these steps to process your annotations:

1. **Prepare Your Files**:

   - Place your Darwin V7 JSON annotation files in a directory (e.g., `path/to/darwin_annotations`).
   - Create an output directory for the DLUP XML files (e.g., `path/to/output`).

2. **Execute Command with Bazelisk**:

   Run the following command in your terminal:

   ```
   bazelisk run //projects/pathology/tumor_stroma_segmentation:darwin_to_dlup_xml -- path/to/darwin_annotations path/to/output
   ```

   This command will:

   - Build the conversion tool using Bazelisk
   - Process the Darwin V7 annotations
   - Output the resulting DLUP XML files to your specified output directory

## Contributions

Contributions, bug reports, and feature requests are welcome. Please open an issue or submit a pull request.

## License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for details.

## Contact

For more information or support, please open an issue [here](https://github.com/NKI-AI/aiforoncology/issues)
