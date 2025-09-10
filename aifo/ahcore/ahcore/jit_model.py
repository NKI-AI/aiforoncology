import dotenv
import hydra
import lightning.pytorch as pl
import torch
import torch.nn as nn
from pathlib import Path
from omegaconf import DictConfig

dotenv.load_dotenv(override=True)


class ModuleToJit(nn.Module):
    def __init__(self, module):
        super().__init__()
        self._model = module._model
        self._model.eval()

    def forward(self, x):
        return self._model(x)


def compile_jit(module: pl.LightningModule, config: DictConfig) -> None:
    """Compile the model using Torch JIT and save it."""
    jitted_model = torch.jit.script(ModuleToJit(module))
    output_path = config.get("jitted_model_path", "jitted_model.pt")
    jitted_model.save(output_path)


@hydra.main(
    config_path=str(Path(__file__).parent / "config"),
    config_name="jit_model.yaml",
    version_base="1.3",
)
def main(config: DictConfig) -> None:
    # data description is needed for num_classes
    data_description = hydra.utils.instantiate(config.data_description)
    model = hydra.utils.instantiate(config.lit_module, data_description=data_description)

    ckpt_path = config.get("ckpt_path", None)
    assert ckpt_path is not None, "Checkpoint path not provided in config."
    lit_ckpt = torch.load(ckpt_path, weights_only=False)
    state_dict = lit_ckpt["state_dict"]
    # remove augmentations from state_dict
    state_dict = {k: v for k, v in state_dict.items() if "_augmentations" not in k}
    model.load_state_dict(state_dict)

    # Compile model
    compile_jit(model, config)


if __name__ == "__main__":
    main()  # pylint: disable=no-value-for-parameter
