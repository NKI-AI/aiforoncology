import hydra
import torch
import os
from omegaconf import DictConfig
from ahcore.utils import debug_utils
from hydra.core.config_search_path import ConfigSearchPath
from hydra.core.plugins import Plugins
from hydra.plugins.search_path_plugin import SearchPathPlugin
from pathlib import Path

debug_utils.TIME_IT_ENABLE = False

from ahcore.hydra_plugins import register_additional_config_search_path  # noqa: E402


# We need to set multiprocessing start method to spawn to avoid issues with the dataloader
# torch.multiprocessing.set_start_method("spawn", force=True)


@hydra.main(
    config_path=str(Path(__file__).parent / "config"),
    config_name="train.yaml",
    version_base="1.3",
)
def main(config: DictConfig) -> torch.Tensor | None:
    # Imports can be nested inside @hydra.main to optimize tab completion
    # https://github.com/facebookresearch/hydra/issues/934
    from ahcore.entrypoints import train
    from ahcore.utils.io import extras, print_config, validate_config

    # Validate config -- Fails if there are mandatory missing values
    validate_config(config)

    # Applies optional utilities
    extras(config)

    if config.get("print_config"):
        print_config(config, resolve=True)

    # Train model
    return train(config)


def train_with_additional_config(additional_config_path: Path):
    register_additional_config_search_path(additional_config_path)
    main()


if __name__ == "__main__":
    main()  # pylint: disable=no-value-for-parameter
