import dotenv
from ahcore.jit_model import jit_model_with_additional_config
from pathlib import Path
import os
import sys

dotenv.load_dotenv(override=True)

jit_model_with_additional_config(Path(__file__).parent / "additional_config")
