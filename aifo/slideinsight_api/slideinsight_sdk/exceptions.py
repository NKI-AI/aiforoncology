"""
Exception classes for the SlideInsight SDK.

Provides typed exceptions for different error scenarios with proper HTTP status mapping.
"""

from __future__ import annotations

from typing import Any, Optional


class SlideInsightError(Exception):
    """Base exception for all SlideInsight SDK errors.

    Parameters
    ----------
    message : str
        Error message describing what went wrong.
    status_code : int, optional
        HTTP status code associated with the error.
    response_data : dict[str, Any], optional
        Raw response data from the API.
    """

    def __init__(self, message: str, status_code: Optional[int] = None, response_data: Optional[dict[str, Any]] = None):
        super().__init__(message)
        self.message = message
        self.status_code = status_code
        self.response_data = response_data or {}


class AuthenticationError(SlideInsightError):
    """Raised when authentication fails or tokens are invalid.

    Parameters
    ----------
    message : str, default="Authentication failed"
        Error message describing the authentication failure.
    status_code : int, default=401
        HTTP status code (typically 401).
    response_data : dict[str, Any], optional
        Raw response data from the API.
    """

    def __init__(
        self,
        message: str = "Authentication failed",
        status_code: Optional[int] = 401,
        response_data: Optional[dict[str, Any]] = None,
    ):
        super().__init__(message, status_code, response_data)


class AuthorizationError(SlideInsightError):
    """Raised when user lacks permission for the requested operation.

    Parameters
    ----------
    message : str, default="Insufficient permissions"
        Error message describing the authorization failure.
    status_code : int, default=403
        HTTP status code (typically 403).
    response_data : dict[str, Any], optional
        Raw response data from the API.
    """

    def __init__(
        self,
        message: str = "Insufficient permissions",
        status_code: Optional[int] = 403,
        response_data: Optional[dict[str, Any]] = None,
    ):
        super().__init__(message, status_code, response_data)


class NotFoundError(SlideInsightError):
    """Raised when a requested resource is not found.

    Parameters
    ----------
    message : str, default="Resource not found"
        Error message describing what resource was not found.
    status_code : int, default=404
        HTTP status code (typically 404).
    response_data : dict[str, Any], optional
        Raw response data from the API.
    """

    def __init__(
        self,
        message: str = "Resource not found",
        status_code: Optional[int] = 404,
        response_data: Optional[dict[str, Any]] = None,
    ):
        super().__init__(message, status_code, response_data)


class ValidationError(SlideInsightError):
    """Raised when request data fails validation.

    Parameters
    ----------
    message : str, default="Validation failed"
        Error message describing the validation failure.
    status_code : int, default=400
        HTTP status code (typically 400).
    response_data : dict[str, Any], optional
        Raw response data from the API.
    """

    def __init__(
        self,
        message: str = "Validation failed",
        status_code: Optional[int] = 400,
        response_data: Optional[dict[str, Any]] = None,
    ):
        super().__init__(message, status_code, response_data)


class ConflictError(SlideInsightError):
    """Raised when a resource conflict occurs (e.g., duplicate creation).

    Parameters
    ----------
    message : str, default="Resource conflict"
        Error message describing the conflict.
    status_code : int, default=409
        HTTP status code (typically 409).
    response_data : dict[str, Any], optional
        Raw response data from the API.
    """

    def __init__(
        self,
        message: str = "Resource conflict",
        status_code: Optional[int] = 409,
        response_data: Optional[dict[str, Any]] = None,
    ):
        super().__init__(message, status_code, response_data)


class RateLimitError(SlideInsightError):
    """Raised when API rate limits are exceeded.

    Parameters
    ----------
    message : str, default="Rate limit exceeded"
        Error message describing the rate limit.
    status_code : int, default=429
        HTTP status code (typically 429).
    response_data : dict[str, Any], optional
        Raw response data from the API.
    """

    def __init__(
        self,
        message: str = "Rate limit exceeded",
        status_code: Optional[int] = 429,
        response_data: Optional[dict[str, Any]] = None,
    ):
        super().__init__(message, status_code, response_data)


class ServerError(SlideInsightError):
    """Raised for server-side errors (5xx status codes).

    Parameters
    ----------
    message : str, default="Server error"
        Error message describing the server error.
    status_code : int, default=500
        HTTP status code (5xx range).
    response_data : dict[str, Any], optional
        Raw response data from the API.
    """

    def __init__(
        self,
        message: str = "Server error",
        status_code: Optional[int] = 500,
        response_data: Optional[dict[str, Any]] = None,
    ):
        super().__init__(message, status_code, response_data)


class NetworkError(SlideInsightError):
    """Raised for network connectivity issues.

    Parameters
    ----------
    message : str, default="Network error"
        Error message describing the network issue.
    original_error : Exception, optional
        The original exception that caused this network error.
    """

    def __init__(self, message: str = "Network error", original_error: Optional[Exception] = None):
        super().__init__(message)
        self.original_error = original_error


class APIError(SlideInsightError):
    """Generic API error for unhandled HTTP status codes.

    Parameters
    ----------
    message : str
        Error message describing the API error.
    status_code : int
        HTTP status code.
    response_data : dict[str, Any], optional
        Raw response data from the API.
    """

    def __init__(self, message: str, status_code: int, response_data: Optional[dict[str, Any]] = None):
        super().__init__(message, status_code, response_data)


def create_exception_from_response(
    status_code: int, message: str, response_data: Optional[dict[str, Any]] = None
) -> SlideInsightError:
    """Create appropriate exception based on HTTP status code.

    Parameters
    ----------
    status_code : int
        HTTP status code from the response.
    message : str
        Error message to include in the exception.
    response_data : dict[str, Any], optional
        Raw response data from the API.

    Returns
    -------
    SlideInsightError
        Appropriate exception instance based on the status code.
    """
    if status_code == 400:
        return ValidationError(message, status_code, response_data)
    elif status_code == 401:
        return AuthenticationError(message, status_code, response_data)
    elif status_code == 403:
        return AuthorizationError(message, status_code, response_data)
    elif status_code == 404:
        return NotFoundError(message, status_code, response_data)
    elif status_code == 409:
        return ConflictError(message, status_code, response_data)
    elif status_code == 429:
        return RateLimitError(message, status_code, response_data)
    elif 500 <= status_code < 600:
        return ServerError(message, status_code, response_data)
    else:
        return APIError(message, status_code, response_data)
