"""
SlideInsight Python SDK

A client for the SlideInsight API with comprehensive support for
studies, cases, slides, annotations, and user management.

Example usage:
    from slideinsight_sdk import SlideInsightClient

    # Create client and authenticate
    client = SlideInsightClient("http://localhost:3000")
    await client.login("email", "password")

    # Use the client
    studies = await client.studies.list()
    study = await client.studies.create(name="My Study", description="...")

    # Pagination support
    async for page in client.slides.list_paginated(page_size=50):
        for slide in page.items:
            print(f"Slide: {slide.slide_name}")

    # Or get a specific page
    page = await client.slides.list(page=2, limit=25)
"""

from .client import SlideInsightClient
from .models import (
    Study,
    Case,
    Slide,
    User,
    Tenant,
    RasterAnnotation,
    VectorAnnotation,
    PaginatedResponse,
    PaginationInfo,
)
from .exceptions import (
    SlideInsightError,
    AuthenticationError,
    NotFoundError,
    ValidationError,
    APIError,
)

__version__ = "1.0.0"
__all__ = [
    "SlideInsightClient",
    # Models
    "Study",
    "Case",
    "Slide",
    "User",
    "Tenant",
    "RasterAnnotation",
    "VectorAnnotation",
    "PaginatedResponse",
    "PaginationInfo",
    # Exceptions
    "SlideInsightError",
    "AuthenticationError",
    "NotFoundError",
    "ValidationError",
    "APIError",
]
