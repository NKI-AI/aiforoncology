# Tissue background

_Authors_: Jonas Teuwen and Joren Brunekreef

## Table of Contents

- [Overview](#overview)

## Overview

In this project we have a segmentation model the tissue and background in an H&E slide

## How to run training

```shell
bazelisk run //projects/pathology/tissue_background:train -- experiment=segmentation datamodule.num_workers=<num_workers>
```

Before training you should copy the data to the local SSD. The output will be relative to your `$SCRATCH` folder.

```shell
bazelisk run //aifo/llm_serveahcore:cli -- data copy-data-from-manifest \
    sqlite:////data/groups/aiforoncology/derived/pathology/tissue_background/manifest.sqlite  \
    v1.0 \
    v1.0 \
    /data/groups/public/archive/ tissue_background_dataset
```

## How to setup

The manifest was created with

```shell
bazelisk run //projects/pathology/tissue_background:create_manifest -- \
    --image-folder-root /data/groups/public/archive/ \
    --tcga-mapping /data/groups/public/archive/TCGA/identifier_mapping.json \
    --camelyon16-mapping /data/groups/public/archive/Camelyon16/identifier_mapping.json \
    --annotation-folder /data/groups/aiforoncology/derived/pathology/tissue_background/ \
    --manifest-path /data/groups/aiforoncology/derived/pathology/tissue_background/manifest.sqlite
```

## Contributions

Contributions, bug reports, and feature requests are welcome. Please open an issue or submit a pull request.

## License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for details.

## Contact

For more information or support, please open an issue [here](https://github.com/NKI-AI/aiforoncology/issues)
