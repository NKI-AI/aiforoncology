# LLM Serve

A lightweight serving infrastructure for Large Language Models using vLLM backend with OpenAI-compatible API.

## Features

- OpenAI-compatible API interface
- Interactive web chat interface with:
  - Streaming responses
  - Session management
  - Markdown support
  - Code syntax highlighting
- Command-line chat client
- Conversation history management
- CLI tools for easy interaction
- SLURM integration for cluster deployment

## Installation

```bash
bazelisk build //...
```

## Quick Start

### Starting the Server

You can start the model server either directly or using SLURM (for 1 gpu):

```bash
# Direct start
bazelisk run //aifo/llm_serve:api_server -- \
    --model /path/to/your/model \
    --served-model-name model_name \
    --tensor-parallel-size 1 \
    --gpu-memory-utilization 0.99 \
    --dtype half \
    --max-model-len 8192 \
    --host 0.0.0.0 \
    --port 8030

# Or using SLURM, note this script also starts the web server
sbatch serving/slurm_serve_a100.sh
```

### Using the Web Interface

The web interface provides an interactive chat experience:

```bash
# Start the web server (default port 8050)
bazelisk run //aifo/llm_serve:server

# Or with custom port
bazelisk run //aifo/llm_serve:server -- --port 8060
```

Features:

- Real-time streaming responses
- Markdown formatting
- Code syntax highlighting
- Session management (30-minute timeout)

### Using the API Client

The API client provides a simple interface for single prompts:

```python
from llm_serve.client.api import LLMApi

# Create API client
api = LLMApi()

# Basic prompt
response = api.prompt("What is the meaning of life?")

# With streaming
response = api.prompt(
    "Tell me a story",
    stream=True,
    temperature=0.8,
    max_tokens=500
)

# Custom server
api = LLMApi(host="localhost", port=8030)
```

### Using the Chat Client

The chat client supports conversation history and interactive sessions:

```python
from llm_serve.client.client import LLMChatClient

# Create chat client
client = LLMChatClient()

# Interactive chat session
client.interactive_chat()

# Single chat with history
response = client.chat("What is the meaning of life?")
response = client.chat("Why do you think that?")  # Follows up on previous

# Chat without history
response = client.chat("Start fresh", include_history=False)

# Streaming vs non-streaming
response = client.chat("Tell me a story", stream=True)
response = client.chat("Tell me another", stream=False)
```

### CLI Usage

```bash
# Interactive chat
llm-chat --interactive

# Single prompt
llm-chat --prompt "What is the meaning of life?"

# Custom server
llm-chat --host localhost --port 8030 --interactive
```

## Model Setup

For example, to use DeepSeek-R1-Distill-Qwen-32B:

```bash
huggingface-cli download deepseek-ai/DeepSeek-R1-Distill-Qwen-32B --local-dir /path/to/model
```

(this model is already stored in `/projects/public_llms/` on the cluster.)

## Configuration

Server settings can be customized through `ServerConfig` in `serving/config.py`.

## Updating Requirements

```bash
bazelisk run //aifo/llm_serve:generate_requirements_{darwin,linux}_txt
```

this will generate `requirements_linux.txt` and `requirements_darwin.txt` depending on the system in the `src/llm_serve` directory. You need to commit these files to the repo.
Do not modify the `requirements_{darwin,linux}.txt` files directly, only modify `requirements.in`.
