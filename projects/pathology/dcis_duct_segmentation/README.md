# DCIS duct segmentation project

_Author_: Jonas Teuwen

## Table of Contents

- [Overview](#overview)

## Overview

In this project we have a segmentation model for DCIS ducts.

## How to run training

```shell
bazelisk run //projects/pathology/dcis_duct_segmentation:train -- experiment=segmentation datamodule.num_workers=<num workers>
```

Before training you should copy the data to the local SSD. The output will be relative to your `$SCRATCH` folder.

```shell
bazelisk run //aifo/ahcore:cli -- data copy-data-from-manifest \
    sqlite:////data/groups/aiforoncology/derived/pathology/PRECISION/DCIS_duct_segmentations/manifest.sqlite \
    v1.0 \
    v1.0 \
    /data/groups/aiforoncology/archive/pathology/PRECISION dcis_ducts_dataset
```

## How to setup

The original images are stored as Halo annotations, and should be converted using

```shell
bazelisk run //projects/pathology/dcis_duct_segmentation:convert_halo_annotations_to_dlup_xml
```

Subsequently, the manifest was created

```shell
bazelisk run //projects/pathology/dcis_duct_segmentation:create_manifest -- \
    --image-root /data/groups/aiforoncology/archive/pathology/PRECISION \
    --annotations-root /data/groups/aiforoncology/derived/pathology/PRECISION/DCIS_duct_segmentations \
    --manifest-path /data/groups/aiforoncology/derived/pathology/PRECISION/DCIS_duct_segmentations/
    manifest.sqlite \
    --validation-codes /data/groups/aiforoncology/derived/pathology/PRECISION/DCIS_duct_segmentations/
    validation_codes.txt
```

This script will also output all the mappings, which you should inspect for anomalies.

## Contributions

Contributions, bug reports, and feature requests are welcome. Please open an issue or submit a pull request.

## License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for details.

## Contact

For more information or support, please open an issue [here](https://github.com/NKI-AI/aiforoncology/issues)
