import torch
from timm.models._factory import create_model
from timm.layers.mlp import SwiGLUPacked
from transformers import AutoModel, AutoConfig
from huggingface_hub import login
import yaml
from pathlib import Path
import argparse
import logging

logging.basicConfig(level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s", datefmt="%H:%M:%S")


class TimmWrapper(torch.nn.Module):
    def __init__(self, model, shift_index):
        super(TimmWrapper, self).__init__()
        self.model = model
        self.shift_index = shift_index

    def forward(self, x):
        out = self.model.forward_features(x)  # shape [B, N, D] for many ViT-like TIMM models
        cls_token = out[:, 0, :]  # shape [B, 1, D]
        patch_token = out[:, self.shift_index :, :]  # skip the CLS and register tokens

        return {"patch_token": patch_token, "cls_token": cls_token}


class TransformerWrapper(torch.nn.Module):
    """
    Wraps a Hugging Face vision transformer model output into { "patch_token": ..., "cls_token": ... }.
    We expect the forward pass to return a tuple (a, b).
    """

    def __init__(self, model, shift_index):
        super().__init__()
        self.model = model
        self.model.eval()
        self.shift_index = shift_index

    def forward(self, x):
        a, b = self.model(x)
        return {"patch_token": a[:, self.shift_index :, :], "cls_token": b}  # skip the CLS and register tokens


def load_config():
    config_path = Path(__file__).parent / "config.yaml"
    with open(config_path) as f:
        return yaml.safe_load(f)["models"]


def generate_jit_models(output_dir: Path):
    config = load_config()
    output_dir.mkdir(parents=True, exist_ok=True)

    total_models = len(config)
    logging.info(f"Starting conversion of {total_models} models to JIT format")

    for idx, (model_key, model_config) in enumerate(config.items(), 1):
        output_path = output_dir / f"{model_key}.jit"

        if output_path.exists():
            logging.info(f"[{idx}/{total_models}] Skipping {model_key} - file already exists at {output_path}")
            continue

        logging.info(f"[{idx}/{total_models}] Processing {model_key}")
        input_shape = model_config["input_shape"]
        dummy_input = torch.rand((1, 3, *input_shape)).float()

        if model_config["type"] == "timm":
            # Handle special cases for SwiGLUPacked and SiLU
            if model_config.get("timm_kwargs", {}).get("mlp_layer") == "SwiGLUPacked":
                model_config["timm_kwargs"]["mlp_layer"] = SwiGLUPacked
            if model_config.get("timm_kwargs", {}).get("act_layer") == "SiLU":
                model_config["timm_kwargs"]["act_layer"] = torch.nn.SiLU

            logging.info(f"Downloading TIMM model: {model_config['model_name']}")
            model = create_model(model_config["model_name"], pretrained=True, **model_config.get("timm_kwargs", {}))
            wrapped_model = TimmWrapper(model, shift_index=model_config["shift_index"])
            logging.info("Converting model to JIT format")
            scripted_model = torch.jit.script(wrapped_model)
            torch.jit.save(scripted_model, output_path)
        elif model_config["type"] == "transformer":
            logging.info(f"Downloading Transformer model: {model_config['model_name']}")
            config = AutoConfig.from_pretrained(model_config["model_name"])
            # Disable any training-related settings
            config.use_cache = False
            config.output_attentions = False
            config.output_hidden_states = False
            config.torchscript = True

            model = AutoModel.from_pretrained(
                model_config["model_name"],
                config=config,
            )
            if hasattr(model, "hf_quantizer") and model.hf_quantizer is not None:
                logging.info("Model has quantization enabled")
                is_trainable = model.hf_quantizer.is_trainable
                logging.info(f"Quantizer is trainable: {is_trainable}")

            model.eval()
            wrapped_model = TransformerWrapper(model, shift_index=model_config["shift_index"])

            # Run a forward pass to ensure the model is traced correctly
            logging.info("Converting model to JIT format")
            with torch.no_grad():
                # Warm up pass to ensure all shapes are correct
                out = wrapped_model(dummy_input)
                logging.info(
                    f"Model output shapes - cls_token: {out['cls_token'].shape}, patch_token: {out['patch_token'].shape}"
                )

                # Now trace the model
                traced_model = torch.jit.trace(wrapped_model, dummy_input, strict=False)
                torch.jit.save(traced_model, output_path)

        logging.info(f"Successfully saved model to {output_path}")


def print_shapes(output_dir: Path):
    config = load_config()
    batch_size = 4

    logging.info("Testing saved models with random inputs")
    for model_key, model_config in config.items():
        input_shape = model_config["input_shape"]
        x = torch.rand((batch_size, 3, *input_shape)).float()

        model_path = output_dir / f"{model_key}.jit"
        logging.info(f"Loading and testing {model_path}")
        model = torch.jit.load(model_path)
        out = model(x)
        print(f"{model_path} output shapes (input shape: {x.shape}):")
        for key, value in out.items():
            print(f"  {key}: {value.shape}")


def run_fomo_jits(args):
    """Run the FOMO JITs download command.

    Args:
        args: Command line arguments
    """
    from pathlib import Path

    output_dir = Path(args.output_dir)
    if args.use_hf_login:
        login()
    generate_jit_models(output_dir)
    if args.test:
        print_shapes(output_dir)


def register_parser(
    parser: argparse._SubParsersAction,  # pylint: disable=unsubscriptable-object
) -> None:  # pylint: disable=E1136
    """Register fomo-jits command to a parser."""
    fomo_parser = parser.add_parser("fomo-jits", help="Download FOMO JIT models from HuggingFace and TIMM")
    fomo_parser.add_argument(
        "--output-dir",
        type=str,
        default="models",
        help="Directory where the JIT models will be saved (default: ./models)",
    )
    fomo_parser.add_argument(
        "--test",
        action="store_true",
        help="Test the models after downloading by running inference with random inputs",
    )
    fomo_parser.add_argument(
        "--use-hf-login",
        action="store_true",
        help="Use HuggingFace login for downloading private models",
    )

    fomo_parser.set_defaults(subcommand=run_fomo_jits)


def parse_args():
    parser = argparse.ArgumentParser(description="Download and convert models to JIT format")
    parser.add_argument(
        "--output-dir",
        type=Path,
        default=Path("models"),
        help="Directory where the JIT models will be saved (default: ./models)",
    )
    return parser.parse_args()


if __name__ == "__main__":
    args = parse_args()
    login()
    generate_jit_models(args.output_dir)
    print_shapes(args.output_dir)
