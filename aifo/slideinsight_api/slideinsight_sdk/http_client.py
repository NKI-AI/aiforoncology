"""
Low-level HTTP client for the SlideInsight SDK.

Handles authentication, token refresh, and HTTP operations with proper error handling.
"""

from __future__ import annotations

import asyncio
import logging
from datetime import datetime, timedelta
from typing import Any, Optional, Union
from urllib.parse import urljoin

import aiohttp
from aiohttp import ClientSession, ClientTimeout, ClientResponse, ClientError

from .exceptions import (
    AuthenticationError,
    NetworkError,
    create_exception_from_response,
)
from .models import AuthTokens


logger = logging.getLogger(__name__)


class HTTPClient:
    """Async HTTP client with authentication and token management.

    Parameters
    ----------
    base_url : str
        Base URL for the SlideInsight API.
    timeout : int, default=30
        Request timeout in seconds.
    max_retries : int, default=3
        Maximum number of retries for failed requests.
    retry_backoff : float, default=1.0
        Backoff factor for retries.
    auth_cookie_name : str, default="_auth"
        Name of the authentication cookie (should match server config).
    refresh_cookie_name : str, optional
        Name of the refresh token cookie. If not provided, defaults to
        auth_cookie_name + "_refresh".
    debug : bool, default=False
        Enable debug logging of HTTP requests and responses.

    Attributes
    ----------
    base_url : str
        Base URL for the SlideInsight API.
    """

    def __init__(
        self,
        base_url: str,
        timeout: int = 30,
        max_retries: int = 3,
        retry_backoff: float = 1.0,
        auth_cookie_name: str = "_auth",
        refresh_cookie_name: Optional[str] = None,
        debug: bool = False,
    ):
        """Initialize HTTP client.

        Parameters
        ----------
        base_url : str
            Base URL for the SlideInsight API.
        timeout : int, default=30
            Request timeout in seconds.
        max_retries : int, default=3
            Maximum number of retries for failed requests.
        retry_backoff : float, default=1.0
            Backoff factor for retries.
        auth_cookie_name : str, default="_auth"
            Name of the authentication cookie (should match server config).
        refresh_cookie_name : str, optional
            Name of the refresh token cookie. If not provided, defaults to
            auth_cookie_name + "_refresh".
        debug : bool, default=False
            Enable debug logging of HTTP requests and responses.
        """
        self.base_url = base_url.rstrip("/")
        self._timeout = ClientTimeout(total=timeout)
        self._max_retries = max_retries
        self._retry_backoff = retry_backoff
        self._debug = debug

        # Cookie configuration - matches server defaults
        self._auth_cookie_name = auth_cookie_name
        self._refresh_cookie_name = refresh_cookie_name or f"{auth_cookie_name}_refresh"

        self._session: Optional[ClientSession] = None
        self._tokens: Optional[AuthTokens] = None
        self._token_expires_at: Optional[datetime] = None
        self._refresh_expires_at: Optional[datetime] = None
        self._refresh_lock = asyncio.Lock()

    async def __aenter__(self) -> "HTTPClient":
        """Async context manager entry.

        Returns
        -------
        HTTPClient
            The HTTP client instance.
        """
        await self._ensure_session()
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb) -> None:
        """Async context manager exit.

        Parameters
        ----------
        exc_type : type, optional
            Exception type if an exception was raised.
        exc_val : Exception, optional
            Exception value if an exception was raised.
        exc_tb : traceback, optional
            Exception traceback if an exception was raised.
        """
        await self.close()

    async def _ensure_session(self) -> None:
        """Ensure HTTP session is created."""
        if not self._session or self._session.closed:
            self._session = aiohttp.ClientSession(
                timeout=self._timeout,
                connector=aiohttp.TCPConnector(limit=100, limit_per_host=30),
                headers={
                    "User-Agent": "slideinsight-python-sdk/1.0.0",
                    "Accept": "application/json",
                    "Content-Type": "application/json",
                },
            )

    async def close(self) -> None:
        """Close the HTTP session."""
        if self._session and not self._session.closed:
            await self._session.close()
        self._session = None

    def _build_url(self, endpoint: str) -> str:
        """Build full URL from endpoint.

        Parameters
        ----------
        endpoint : str
            API endpoint path.

        Returns
        -------
        str
            Complete URL for the API endpoint.
        """
        if endpoint.startswith("http"):
            return endpoint
        return urljoin(f"{self.base_url}/", endpoint.lstrip("/"))

    async def login(self, email: str, password: str) -> AuthTokens:
        """Authenticate with email and password.

        Parameters
        ----------
        email : str
            Username for authentication.
        password : str
            Password for authentication.

        Returns
        -------
        AuthTokens
            Authentication tokens.

        Raises
        ------
        AuthenticationError
            If authentication fails.
        """
        url = self._build_url("/api/v1/auth/login")
        data = {"email": email, "password": password}

        try:
            response_data = await self._make_request("POST", url, json=data, skip_auth=True)
            self._tokens = AuthTokens.from_dict(response_data)

            # Calculate token expiration times
            now = datetime.now()
            self._token_expires_at = now + timedelta(seconds=self._tokens.expires_in)
            self._refresh_expires_at = now + timedelta(seconds=self._tokens.refresh_expires_in)

            logger.info("Successfully authenticated with SlideInsight API")
            return self._tokens

        except Exception as e:
            if isinstance(e, AuthenticationError):
                raise
            raise AuthenticationError(f"Login failed: {e}") from e

    async def refresh_tokens(self) -> AuthTokens:
        """Refresh access token using refresh token.

        Returns
        -------
        AuthTokens
            New authentication tokens.

        Raises
        ------
        AuthenticationError
            If token refresh fails.
        """
        if not self._tokens:
            raise AuthenticationError("No tokens available for refresh")

        async with self._refresh_lock:
            # Check if tokens were already refreshed by another request
            if self._is_access_token_valid():
                return self._tokens

            url = self._build_url("/api/v1/auth/refresh")

            # Set refresh token in cookies as expected by the API
            # Use the configured refresh cookie name that matches server config
            refresh_cookies = {
                self._refresh_cookie_name: self._tokens.refresh_token,
            }

            try:
                # Make sure to use the refresh cookies directly without merging with auth cookies
                await self._ensure_session()

                async with self._session.request(
                    method="POST",
                    url=url,
                    cookies=refresh_cookies,
                    headers={
                        "User-Agent": "slideinsight-python-sdk/1.0.0",
                        "Accept": "application/json",
                        "Content-Type": "application/json",
                    },
                ) as response:
                    response_data = await self._handle_response(response)

                self._tokens = AuthTokens.from_dict(response_data)

                # Update expiration times
                now = datetime.now()
                self._token_expires_at = now + timedelta(seconds=self._tokens.expires_in)
                self._refresh_expires_at = now + timedelta(seconds=self._tokens.refresh_expires_in)

                logger.info("Successfully refreshed access token")
                return self._tokens

            except Exception as e:
                self._tokens = None
                self._token_expires_at = None
                self._refresh_expires_at = None
                raise AuthenticationError(f"Token refresh failed: {e}") from e

    def _is_access_token_valid(self) -> bool:
        """Check if access token is still valid (with 30 second buffer).

        Returns
        -------
        bool
            True if access token is valid, False otherwise.
        """
        if not self._token_expires_at:
            return False
        return datetime.now() < (self._token_expires_at - timedelta(seconds=30))

    def _is_refresh_token_valid(self) -> bool:
        """Check if refresh token is still valid.

        Returns
        -------
        bool
            True if refresh token is valid, False otherwise.
        """
        if not self._refresh_expires_at:
            return False
        return datetime.now() < self._refresh_expires_at

    async def _ensure_valid_token(self) -> None:
        """Ensure we have a valid access token, refreshing if necessary.

        Raises
        ------
        AuthenticationError
            If not authenticated or token refresh fails.
        """
        if not self._tokens:
            raise AuthenticationError("Not authenticated. Please login first.")

        if not self._is_access_token_valid():
            if not self._is_refresh_token_valid():
                raise AuthenticationError("Refresh token expired. Please login again.")
            await self.refresh_tokens()

    async def request(
        self,
        method: str,
        endpoint: str,
        params: Optional[dict[str, Any]] = None,
        json: Optional[dict[str, Any]] = None,
        data: Optional[Union[str, bytes]] = None,
        headers: Optional[dict[str, str]] = None,
        cookies: Optional[dict[str, str]] = None,
        skip_auth: bool = False,
    ) -> Any:
        """Make authenticated HTTP request.

        Parameters
        ----------
        method : str
            HTTP method (GET, POST, PUT, DELETE, etc.).
        endpoint : str
            API endpoint path.
        params : dict[str, Any], optional
            URL parameters.
        json : dict[str, Any], optional
            JSON data to send in request body.
        data : str or bytes, optional
            Raw data to send in request body.
        headers : dict[str, str], optional
            Additional HTTP headers.
        cookies : dict[str, str], optional
            Additional cookies.
        skip_auth : bool, default=False
            Skip authentication for this request.

        Returns
        -------
        Any
            Response data (JSON parsed if applicable).
        """
        url = self._build_url(endpoint)
        return await self._make_request(
            method=method,
            url=url,
            params=params,
            json=json,
            data=data,
            headers=headers,
            cookies=cookies,
            skip_auth=skip_auth,
        )

    async def request_binary(
        self,
        method: str,
        endpoint: str,
        params: Optional[dict[str, Any]] = None,
        headers: Optional[dict[str, str]] = None,
        skip_auth: bool = False,
    ) -> tuple[str, bytes]:
        """Make authenticated HTTP request expecting binary response.

        Parameters
        ----------
        method : str
            HTTP method (GET, POST, PUT, DELETE, etc.).
        endpoint : str
            API endpoint path.
        params : dict[str, Any], optional
            URL parameters.
        headers : dict[str, str], optional
            Additional HTTP headers.
        skip_auth : bool, default=False
            Skip authentication for this request.

        Returns
        -------
        tuple[str, bytes]
            Tuple of (content_type, binary_data).
        """
        url = self._build_url(endpoint)
        return await self._make_binary_request(
            method=method,
            url=url,
            params=params,
            headers=headers,
            skip_auth=skip_auth,
        )

    def _redact_sensitive_data(self, data: dict[str, Any]) -> dict[str, Any]:
        """Redact sensitive data from logging.

        Parameters
        ----------
        data : dict[str, Any]
            Data dictionary to redact.

        Returns
        -------
        dict[str, Any]
            Data with sensitive fields redacted.
        """
        if not data:
            return data

        redacted = data.copy()
        sensitive_keys = {
            "password",
            "token",
            "access_token",
            "refresh_token",
            "authorization",
            "cookie",
            "set-cookie",
            "auth",
            "secret",
            "key",
            "credential",
        }

        for key, value in redacted.items():
            if any(sensitive in key.lower() for sensitive in sensitive_keys):
                if isinstance(value, str) and len(value) > 8:
                    redacted[key] = f"{value[:4]}...{value[-4:]}"
                else:
                    redacted[key] = "***REDACTED***"
            elif isinstance(value, dict):
                redacted[key] = self._redact_sensitive_data(value)

        return redacted

    def _log_request(
        self,
        method: str,
        url: str,
        params: Optional[dict[str, Any]] = None,
        json_data: Optional[dict[str, Any]] = None,
        headers: Optional[dict[str, str]] = None,
        cookies: Optional[dict[str, str]] = None,
    ) -> None:
        """Log HTTP request details for debugging.

        Parameters
        ----------
        method : str
            HTTP method.
        url : str
            Request URL.
        params : dict[str, Any], optional
            URL parameters.
        json_data : dict[str, Any], optional
            JSON request body.
        headers : dict[str, str], optional
            Request headers.
        cookies : dict[str, str], optional
            Request cookies.
        """
        if not self._debug:
            return

        logger.debug("=" * 60)
        logger.debug(f"🔄 HTTP REQUEST: {method} {url}")
        logger.debug("=" * 60)

        if params:
            redacted_params = self._redact_sensitive_data(params)
            logger.debug(f"📝 Query Parameters: {redacted_params}")

        if headers:
            redacted_headers = self._redact_sensitive_data(headers)
            logger.debug(f"📋 Headers: {redacted_headers}")

        if cookies:
            redacted_cookies = self._redact_sensitive_data(cookies)
            logger.debug(f"🍪 Cookies: {redacted_cookies}")

        if json_data:
            redacted_json = self._redact_sensitive_data(json_data)
            logger.debug(f"📦 JSON Body: {redacted_json}")

        logger.debug("=" * 60)

    async def _make_request(
        self,
        method: str,
        url: str,
        params: Optional[dict[str, Any]] = None,
        json: Optional[dict[str, Any]] = None,
        data: Optional[Union[str, bytes]] = None,
        headers: Optional[dict[str, str]] = None,
        cookies: Optional[dict[str, str]] = None,
        skip_auth: bool = False,
    ) -> Any:
        """Make HTTP request with retries and error handling.

        Parameters
        ----------
        method : str
            HTTP method.
        url : str
            Complete URL for the request.
        params : dict[str, Any], optional
            URL parameters.
        json : dict[str, Any], optional
            JSON data to send.
        data : str or bytes, optional
            Raw data to send.
        headers : dict[str, str], optional
            Additional headers.
        cookies : dict[str, str], optional
            Additional cookies.
        skip_auth : bool, default=False
            Skip authentication for this request.

        Returns
        -------
        Any
            Response data.

        Raises
        ------
        NetworkError
            If the request fails after all retries.
        """
        await self._ensure_session()

        # Prepare authentication
        auth_headers = {}
        auth_cookies = {}

        if not skip_auth:
            await self._ensure_valid_token()
            auth_headers["Authorization"] = f"Bearer {self._tokens.access_token}"
            auth_cookies.update(
                {
                    self._auth_cookie_name: self._tokens.access_token,
                    self._refresh_cookie_name: self._tokens.refresh_token,
                }
            )

        # Merge headers and cookies
        final_headers = {**auth_headers, **(headers or {})}
        final_cookies = {**auth_cookies, **(cookies or {})}

        # Log request details in debug mode
        self._log_request(
            method=method,
            url=url,
            params=params,
            json_data=json,
            headers=final_headers,
            cookies=final_cookies,
        )

        last_exception = None

        for attempt in range(self._max_retries + 1):
            try:
                async with self._session.request(
                    method=method,
                    url=url,
                    params=params,
                    json=json,
                    data=data,
                    headers=final_headers,
                    cookies=final_cookies,
                ) as response:
                    return await self._handle_response(response)

            except (ClientError, asyncio.TimeoutError) as e:
                last_exception = e
                if attempt < self._max_retries:
                    wait_time = self._retry_backoff * (2**attempt)
                    logger.warning(
                        f"Request failed (attempt {attempt + 1}/{self._max_retries + 1}): {e}. "
                        f"Retrying in {wait_time:.1f}s..."
                    )
                    await asyncio.sleep(wait_time)
                    continue
                else:
                    raise NetworkError(f"Request failed after {self._max_retries + 1} attempts") from e

        # This should never be reached, but just in case
        raise NetworkError("Request failed") from last_exception

    async def _make_binary_request(
        self,
        method: str,
        url: str,
        params: Optional[dict[str, Any]] = None,
        headers: Optional[dict[str, str]] = None,
        skip_auth: bool = False,
    ) -> tuple[str, bytes]:
        """Make HTTP request expecting binary response.

        Parameters
        ----------
        method : str
            HTTP method.
        url : str
            Complete URL for the request.
        params : dict[str, Any], optional
            URL parameters.
        headers : dict[str, str], optional
            Additional headers.
        skip_auth : bool, default=False
            Skip authentication for this request.

        Returns
        -------
        tuple[str, bytes]
            Tuple of (content_type, binary_data).

        Raises
        ------
        NetworkError
            If the request fails.
        """
        await self._ensure_session()

        # Prepare authentication
        auth_headers = {}
        auth_cookies = {}

        if not skip_auth:
            await self._ensure_valid_token()
            auth_headers["Authorization"] = f"Bearer {self._tokens.access_token}"
            # Use configured cookie names
            auth_cookies.update(
                {
                    self._auth_cookie_name: self._tokens.access_token,
                    self._refresh_cookie_name: self._tokens.refresh_token,
                }
            )

        # Merge headers
        final_headers = {**auth_headers, **(headers or {})}
        # Remove Content-Type for binary requests
        final_headers.pop("Content-Type", None)

        try:
            async with self._session.request(
                method=method,
                url=url,
                params=params,
                headers=final_headers,
                cookies=auth_cookies,
            ) as response:
                if response.status >= 400:
                    # Try to get error message
                    try:
                        error_data = await response.json()
                        message = error_data.get("message", f"HTTP {response.status}")
                    except:
                        message = f"HTTP {response.status}"

                    raise create_exception_from_response(response.status, message)

                content_type = response.headers.get("Content-Type", "application/octet-stream")
                binary_data = await response.read()
                return content_type, binary_data

        except (ClientError, asyncio.TimeoutError) as e:
            raise NetworkError(f"Binary request failed: {e}") from e

    async def _handle_response(self, response: ClientResponse) -> Any:
        """Handle HTTP response with proper error handling.

        Parameters
        ----------
        response : ClientResponse
            HTTP response object.

        Returns
        -------
        Any
            Parsed response data.

        Raises
        ------
        SlideInsightError
            If the response indicates an error.
        """
        if response.status >= 400:
            # Try to parse error response
            try:
                error_data = await response.json()
                message = error_data.get("message", f"HTTP {response.status}")
                raise create_exception_from_response(response.status, message, error_data)
            except Exception as json_error:
                # Fallback to status text if JSON parsing fails
                message = f"HTTP {response.status}: {response.reason}"
                raise create_exception_from_response(response.status, message)

        # Try to parse JSON response
        try:
            return await response.json()
        except Exception:
            # Return raw text if JSON parsing fails
            return await response.text()
