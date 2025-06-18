import click
from llm_serve.client.client import LLMChatClient
from llm_serve.serving.config import ServerConfig


@click.command()
@click.option("--interactive", is_flag=True, help="Start interactive chat")
@click.option("--prompt", help="Single prompt to process")
@click.option("--model-path", help="Path to model", default=None)
@click.option("--port", type=int, help="Server port", default=None)
@click.option("--host", help="Server host", default=None)
def main(interactive: bool, prompt: str, model_path: str, port: int, host: str) -> None:
    """Chat with the LLM"""
    # Create config with CLI overrides
    config = ServerConfig()
    if model_path:
        config.model_path = model_path
    if port:
        config.port = port
    if host:
        config.host = host

    client = LLMChatClient(config=config)

    if interactive:
        client.interactive_chat()
    elif prompt:
        client.chat(prompt, print_response=True)
    else:
        print("Please specify --interactive or --prompt")


if __name__ == "__main__":
    main()
