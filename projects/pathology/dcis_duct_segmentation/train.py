import dotenv
from ahcore.train import train_with_additional_config
from pathlib import Path
import os
import sys

dotenv.load_dotenv(override=True)

train_with_additional_config(Path(__file__).parent / "additional_config")
