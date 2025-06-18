# Theoretical Bounds for AI-based Adaptive Breast Cancer Screening with MRI

## Overview

This project calculates and visualizes theoretical bounds for an AI-based adaptive breast cancer screening protocol using MRI. It analyzes how different parameters affect:

1. Recall rates: The proportion of patients who need additional full protocol scans
2. Protocol duration: The average time required per patient

The tool helps evaluate the potential benefits and trade-offs of implementing an AI-based adaptive screening approach compared to standard abbreviated or full protocols.

## Quick Start

### Using Bazel (Recommended)

From the monorepo root:

```bash
bazelisk run //projects/radiology/adaptive_mri_bounds:main -- --save_dir=<output_directory>
```

### Using Python directly

From the monorepo root:

1. Install dependencies:
   ```bash
   pip install -r projects/radiology/adaptive_mri_bounds/requirements.txt
   ```
2. Run the script:
   ```bash
   python projects/radiology/adaptive_mri_bounds/main.py --save_dir=<output_directory>
   ```

## Configuration

The analysis parameters can be customized by editing the TOML files in the `configs` directory:

- `performance_parameters.toml`: Base performance metrics
  - Sensitivity and specificity for radiologists and AI
  - Protocol durations
  - Cancer prevalence
- `changing_parameters.toml`: Parameter ranges for analysis
- `plot_parameters.toml`: Visualization settings

## Output

The script generates figures showing:

- Recall rates for different protocols (Adaptive, Abbreviated, Full)
- Average protocol times (Adaptive, Abbreviated, Full)
- Best and worst-case scenarios for the adaptive protocol

All figures are saved to the specified output directory.

## License

This project is licensed under the Apache 2.0 License. See the [LICENSE](/LICENSE) file for details.

## Support

For questions or issues, please open an issue [here](https://github.com/NKI-AI/aiforoncology/issues).
