import argparse
import asyncio
import html
import json
import socket
import time
from contextlib import asynccontextmanager
from dataclasses import dataclass
from pathlib import Path
from typing import AsyncGenerator, Dict, Optional

from llm_serve.client.api import LLMApi
from llm_serve.client.client import LLMChatClient
from fastapi import FastAPI, Request
from fastapi.responses import HTMLResponse, StreamingResponse
from fastapi.templating import Jinja2Templates


@dataclass
class SessionState:
    last_active: float
    is_expired: bool = False
    client: Optional[LLMChatClient] = None


# Configuration
SESSION_TIMEOUT = 1800  # 30 minutes
CLEANUP_INTERVAL = 300  # 5 minutes
FINAL_CLEANUP_TIMEOUT = 86400  # 24 hours - time after expiry when session is completely removed

# Store session states and active streams
sessions: Dict[str, SessionState] = {}
active_streams: Dict[str, asyncio.Event] = {}


def get_or_create_client(session_id: str) -> LLMChatClient:
    now = time.time()
    if session_id not in sessions:
        # Create new session
        client = LLMChatClient()
        sessions[session_id] = SessionState(last_active=now, client=client)
    else:
        # Reset expiry and update activity time
        sessions[session_id].is_expired = False
        sessions[session_id].last_active = now
        if not sessions[session_id].client:
            sessions[session_id].client = LLMChatClient()

    return sessions[session_id].client  # type: ignore


async def session_reaper_task() -> None:
    """Background task to clean up expired sessions."""
    while True:
        await asyncio.sleep(CLEANUP_INTERVAL)
        print("Checking for expired sessions")
        now = time.time()

        # Phase 1: Mark expired sessions
        for session_id, state in sessions.items():
            if not state.is_expired and (now - state.last_active > SESSION_TIMEOUT):
                print(f"Marking session as expired: {session_id}")
                state.is_expired = True
                state.client = None  # Free the client
                # Stop any active streams
                stop_event = active_streams.pop(session_id, None)
                if stop_event is not None:
                    stop_event.set()

        # Phase 2: Remove very old expired sessions
        to_remove = []
        for session_id, state in sessions.items():
            if state.is_expired and (now - state.last_active > FINAL_CLEANUP_TIMEOUT):
                print(f"Removing old expired session: {session_id}")
                to_remove.append(session_id)

        # Remove the sessions outside the iteration
        for session_id in to_remove:
            del sessions[session_id]


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    # Startup: create background task
    reaper_task = asyncio.create_task(session_reaper_task())
    yield
    # Shutdown: cancel background task
    reaper_task.cancel()
    try:
        await reaper_task
    except asyncio.CancelledError:
        pass


# Initialize FastAPI app with lifespan
app = FastAPI(lifespan=lifespan)
web_dir = Path(__file__).parent
templates = Jinja2Templates(directory=str(web_dir))

# Get model info
api = LLMApi()
try:
    MODEL_NAME = api.model_name or "Unknown Model"
except Exception:
    MODEL_NAME = "Unknown Model"


@app.get("/", response_class=HTMLResponse)  # type: ignore
async def root(request: Request) -> HTMLResponse:
    return templates.TemplateResponse("index.html", {"request": request, "model_name": MODEL_NAME})


@app.post("/stop/{session_id}")  # type: ignore
async def stop_stream(session_id: str) -> dict[str, str]:
    """Stop the active stream for a session"""
    # Check if session exists and is expired
    if session_id in sessions and sessions[session_id].is_expired:
        return {
            "status": "session_expired",
            "message": "Session expired. Please refresh the page.",
        }

    if session_id in active_streams:
        active_streams[session_id].set()
        return {"status": "stopped"}
    return {"status": "no_active_stream"}


@app.post("/chat/{session_id}")  # type: ignore
async def chat(session_id: str, request: Request) -> StreamingResponse:
    """
    Endpoint that streams chunks of text via Server-Sent Events (SSE).
    """
    # Check if session is expired
    if session_id in sessions and sessions[session_id].is_expired:
        msg = {
            "chunk": "[Session expired. Please refresh the page to start a new session.]",
            "full": "[Session expired. Please refresh the page to start a new session.]",
            "session_timeout": True,
        }

        async def expired_response() -> AsyncGenerator[str, None]:
            yield f"data: {json.dumps(msg)}\n\n"

        return StreamingResponse(expired_response(), media_type="text/event-stream")

    # Get or create client (this also updates activity time)
    client = get_or_create_client(session_id)
    data = await request.json()
    prompt = data["message"]

    # Create new stop event
    stop_event = asyncio.Event()
    active_streams[session_id] = stop_event

    async def cleanup() -> None:
        """Remove the stop event when the stream ends"""
        if session_id in active_streams:
            del active_streams[session_id]

    async def generate_response() -> AsyncGenerator[str, None]:
        try:
            full_response = ""
            response_chunks = client.chat(prompt, stream=True, print_response=False)
            assert not isinstance(response_chunks, str)  # Should be a generator

            for chunk in response_chunks:
                # Check if we should stop
                if stop_event.is_set():
                    interrupt_msg = "\n\n[Interrupted]"
                    data = {
                        "chunk": interrupt_msg,
                        "full": full_response + interrupt_msg,
                    }
                    yield f"data: {json.dumps(data)}\n\n"
                    break

                if chunk:
                    escaped_chunk = html.escape(chunk)
                    full_response += chunk
                    sse_data = json.dumps({"chunk": escaped_chunk, "full": html.escape(full_response)})
                    yield f"data: {sse_data}\n\n"
                    await asyncio.sleep(0)

        except Exception as e:
            error_msg = f"Error: {str(e)}"
            yield f"data: {json.dumps({'chunk': error_msg, 'full': error_msg})}\n\n"
        finally:
            await cleanup()

    return StreamingResponse(generate_response(), media_type="text/event-stream")


if __name__ == "__main__":
    import uvicorn

    # Add command line argument parsing
    parser = argparse.ArgumentParser(description="Start the LLM web server")
    parser.add_argument(
        "--port",
        type=int,
        default=8050,
        help="Port to run the server on (default: 8050)",
    )
    parser.add_argument(
        "--host",
        type=str,
        default="0.0.0.0",
        help="Host to run the server on (default: 0.0.0.0)",
    )

    args = parser.parse_args()
    hostname = socket.gethostname()

    print("\nServer starting on:")
    print(f"- Local URL: http://localhost:{args.port}")
    print(f"- Network URL: http://{hostname}:{args.port}")
    print("\nUse the Network URL to access from other machines in the network.\n")

    uvicorn.run(app, host=args.host, port=args.port)
