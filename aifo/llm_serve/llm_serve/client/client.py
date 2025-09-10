import logging
from typing import Any, Dict, Generator, List, Optional

from llm_serve.serving.config import ServerConfig
from openai import OpenAI


class LLMChatClient:
    def __init__(self, config: Optional[ServerConfig] = None):
        self.config = config or ServerConfig()
        self.client = OpenAI(
            base_url=f"http://{self.config.host}:{self.config.port}/v1",
            api_key="no-key-required",
            timeout=self.config.request_timeout,
        )
        self.conversation: List[Dict[str, str]] = []
        self.logger = logging.getLogger(__name__)

    def estimate_tokens(self, text: str) -> int:
        """Rough estimation of tokens (4 chars ≈ 1 token)."""
        return len(text) // 4

    def trim_conversation_history(self, max_tokens: Optional[int] = None) -> None:
        """Trim conversation history to fit within token limit."""
        if not max_tokens:
            # Default to 75% of max model length to leave room for new messages
            max_tokens = int(self.config.max_model_length * 0.75)

        while self.conversation:
            total_tokens = sum(self.estimate_tokens(msg["content"]) for msg in self.conversation)
            if total_tokens <= max_tokens:
                break
            # Remove oldest message pair (user + assistant)
            self.conversation = self.conversation[2:] if len(self.conversation) >= 2 else []

    def add_to_history(self, role: str, content: str) -> None:
        self.conversation.append({"role": role, "content": content})
        self.trim_conversation_history()

    def chat(
        self,
        prompt: str,
        include_history: bool = True,
        stream: bool = True,
        print_response: bool = False,
    ) -> Any:
        """
        Chat with the model. If `stream=True`, returns a generator of chunks;
        otherwise returns a single string with the full response.
        """
        messages = self.conversation if include_history else []

        # Add the user's message to history first
        if include_history:
            self.add_to_history("user", prompt)
            # Use the updated conversation for the request
            messages = self.conversation

        # Rough estimation of input tokens (4 chars ≈ 1 token)
        input_tokens = sum(len(msg["content"]) // 4 for msg in messages)

        # Leave some buffer for the response (20% of max length)
        buffer_tokens = self.config.max_model_length // 5
        max_response_tokens = max(
            100,  # Minimum response length
            min(
                self.config.max_model_length - input_tokens - buffer_tokens,
                self.config.max_model_length // 2,  # Cap at 50% of max length
            ),
        )

        if not stream:
            # Non-streaming request: return the entire response at once
            try:
                response = self.client.chat.completions.create(
                    model=self.config.model_name,
                    messages=messages,
                    temperature=0.6,
                    max_tokens=max_response_tokens,
                    stream=False,
                )
                content = response.choices[0].message.content

                # Add only the assistant's response to history
                if include_history:
                    self.add_to_history("assistant", content)

                if print_response:
                    print(content)

                return content

            except Exception as e:
                self.logger.error(f"Error in chat: {str(e)}")

                def error_generator(error_msg: str) -> Generator[str, None, None]:
                    yield f"Error: {error_msg}"

                return error_generator(str(e))

        # Streaming request: return a generator
        try:
            response = self.client.chat.completions.create(
                model=self.config.model_name,
                messages=messages,
                temperature=0.6,
                max_tokens=max_response_tokens,
                stream=True,
            )

            def stream_generator() -> Generator[str, None, None]:
                full_response = ""
                try:
                    for chunk in response:
                        if chunk.choices and chunk.choices[0].delta.content:
                            content = chunk.choices[0].delta.content
                            full_response += content
                            yield content

                    # Add only the assistant's response to history after successful streaming
                    if include_history and full_response.strip():
                        self.add_to_history("assistant", full_response)

                except Exception as e:
                    self.logger.error(f"Error in stream: {str(e)}")
                    yield f"Error: {str(e)}"

            return stream_generator()

        except Exception as e:
            self.logger.error(f"Error in chat: {str(e)}")
            return (f"Error: {str(e)}" for _ in range(1))  # Single-item generator for consistency

    def interactive_chat(self) -> None:
        """Start interactive chat session (for terminal usage)."""
        print("Starting interactive chat (type 'exit' to quit)")
        while True:
            user_input = input("\nUser: ")
            if user_input.lower() in ["exit", "quit"]:
                break
            print("\nAssistant: ", end="", flush=True)  # Start response line

            # Use the generator version
            response_chunks = self.chat(user_input, stream=True, print_response=True)
            # Here we just exhaust the generator so that it prints to console
            if not isinstance(response_chunks, str):
                list(response_chunks)  # Force it to read all chunks
