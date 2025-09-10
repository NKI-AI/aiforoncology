from typing import Iterator, Optional, Union

from llm_serve.serving.config import ServerConfig
from openai import OpenAI


class LLMApi:
    """Lightweight API wrapper for LLM interactions"""

    def __init__(
        self,
        host: Optional[str] = None,
        port: Optional[int] = None,
        timeout: Optional[int] = None,
    ):
        self.config = ServerConfig()
        self.model_name = self.config.model_name
        if not host:
            print(f"Using default host: {self.config.host}")
        self.host = host or self.config.host
        if not port:
            print(f"Using default port: {self.config.port}")
        self.port = port or self.config.port
        self.timeout = timeout or self.config.request_timeout

        self.client = OpenAI(
            base_url=f"http://{self.host}:{self.port}/v1",
            api_key="no-key-required",
            timeout=self.timeout,
        )

    def prompt(
        self,
        text: str,
        max_tokens: Optional[int] = None,
        temperature: float = 0.7,
        stream: bool = False,
        print_response: bool = True,
    ) -> Union[str, Iterator[str]]:
        """Send a prompt to the LLM and get a response"""
        response = self.client.chat.completions.create(
            model=self.config.model_name,
            messages=[{"role": "user", "content": text}],
            temperature=temperature,
            max_tokens=max_tokens,
            stream=stream,
        )

        if stream:
            full_content = ""
            for chunk in response:
                if chunk.choices[0].delta.content is not None:
                    content_chunk = chunk.choices[0].delta.content
                    print(content_chunk, end="", flush=True)
                    full_content += content_chunk
            return full_content
        else:
            full_content = response.choices[0].message.content

        if print_response and not stream:
            print(full_content)
        return full_content


if __name__ == "__main__":
    api = LLMApi()
    result = api.prompt("What is the meaning of life?", stream=False, print_response=True)
