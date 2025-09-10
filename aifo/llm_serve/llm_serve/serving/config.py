import os
from dataclasses import dataclass


@dataclass
class ServerConfig:
    model_path: str = os.environ.get("AIFO_MODEL_PATH", "/projects/public_llms/qwen32bdeepseek")
    model_name: str = os.environ.get("AIFO_MODEL_NAME", "qwen32bdeepseek")

    host: str = os.environ.get("AIFO_HOST", "0.0.0.0")
    port: int = int(os.environ.get("AIFO_PORT", "8030"))

    max_model_length: int = int(os.environ.get("AIFO_MAX_MODEL_LENGTH", "32768"))

    request_timeout: int = int(os.environ.get("AIFO_REQUEST_TIMEOUT", "60"))
