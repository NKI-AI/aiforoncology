import dotenv
import hydra
import lightning.pytorch as pl
import torch
import torch.nn as nn
from pathlib import Path
from omegaconf import DictConfig

dotenv.load_dotenv(override=True)

from ahcore.hydra_plugins import register_additional_config_search_path  # noqa: E402


class ModuleToJit(nn.Module):
    def __init__(self, module):
        super().__init__()
        self._model = module._model
        self._model.eval()

    def forward(self, x):
        if hasattr(self, "_augmentations"):
            x = self._augmentations(x)
        x = self._model(x / 255.0)
        return x


def compile_jit(module: pl.LightningModule, config: DictConfig) -> None:
    """Compile the model using Torch JIT and save it."""
    jitted_model = torch.jit.script(ModuleToJit(module))
    output_path = config.get("jitted_model_path", "jitted_model.pt")
    jitted_model.save(output_path)


@hydra.main(
    config_path=str(Path(__file__).parent / "config"),
    config_name="inference.yaml",
    version_base="1.3",
)
def main(config: DictConfig) -> None:
    # Imports can be nested inside @hydra.main to optimize tab completion
    # https://github.com/facebookresearch/hydra/issues/934
    from ahcore.entrypoints import general_setup
    from ahcore.utils.io import extras, print_config, validate_config

    # Validate config -- Fails if there are mandatory missing values
    validate_config(config)

    # Applies optional utilities
    extras(config)

    print_config(config, resolve=True)

    # Setup model and datamodule
    # TODO: Datamodule not needed for JIT compilation
    model, datamodule = general_setup(config)

    # Compile model
    compile_jit(model, config)


def jit_model_with_additional_config(additional_config_path: Path):
    register_additional_config_search_path(additional_config_path)
    main()


if __name__ == "__main__":
    main()  # pylint: disable=no-value-for-parameter
