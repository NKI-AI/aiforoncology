"""
Main SlideInsight SDK client.

Provides the primary interface for interacting with the SlideInsight API
with proper authentication, resource management, and error handling.
"""

from __future__ import annotations

import logging
from typing import Optional

from .http_client import HTTPClient
from .models import AuthTokens, User
from .resources import (
    StudiesManager,
    CasesManager,
    SlidesManager,
    UsersManager,
    TenantsManager,
)


logger = logging.getLogger(__name__)


class SlideInsightClient:
    """Main SlideInsight API client.

    Provides access to all SlideInsight API resources through a unified interface.
    Handles authentication, token refresh, and provides typed access to all endpoints.

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
    studies : StudiesManager
        Manager for study resources.
    cases : CasesManager
        Manager for case resources.
    slides : SlidesManager
        Manager for slide resources.
    users : UsersManager
        Manager for user resources.
    tenants : TenantsManager
        Manager for tenant resources.

    Examples
    --------
    Async context manager (recommended):

    >>> async with SlideInsightClient("http://localhost:3000") as client:
    ...     await client.login("email", "password")
    ...     # Use resource managers
    ...     studies = await client.studies.list()
    ...     study = await client.studies.create(name="My Study")

    Manual management:

    >>> client = SlideInsightClient("http://localhost:3000")
    >>> try:
    ...     await client.login("email", "password")
    ...     studies = await client.studies.list()
    ... finally:
    ...     await client.close()

    Custom cookie configuration (if server uses different names):

    >>> client = SlideInsightClient(
    ...     "http://localhost:3000",
    ...     auth_cookie_name="slideinsight_token"
    ... )

    Enable debug logging to see HTTP requests:

    >>> client = SlideInsightClient(
    ...     "http://localhost:3000",
    ...     debug=True
    ... )
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
        """Initialize SlideInsight client.

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
        self.base_url = base_url
        self._http_client = HTTPClient(
            base_url=base_url,
            timeout=timeout,
            max_retries=max_retries,
            retry_backoff=retry_backoff,
            auth_cookie_name=auth_cookie_name,
            refresh_cookie_name=refresh_cookie_name,
            debug=debug,
        )

        # Initialize resource managers
        self.studies = StudiesManager(self._http_client)
        self.cases = CasesManager(self._http_client)
        self.slides = SlidesManager(self._http_client)
        self.users = UsersManager(self._http_client)
        self.tenants = TenantsManager(self._http_client)

        logger.info(f"Initialized SlideInsight client for {base_url}")

    async def __aenter__(self) -> "SlideInsightClient":
        """Async context manager entry.

        Returns
        -------
        SlideInsightClient
            The client instance.
        """
        await self._http_client.__aenter__()
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
        await self._http_client.__aexit__(exc_type, exc_val, exc_tb)

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
        logger.info(f"Logging in user: {email}")
        tokens = await self._http_client.login(email, password)
        logger.info("Successfully authenticated")
        return tokens

    async def logout(self) -> None:
        """Logout and clear authentication tokens.

        Calls the logout endpoint and clears stored tokens.
        """
        try:
            await self._http_client.request("POST", "/api/v1/auth/logout")
        except Exception as e:
            logger.warning(f"Logout request failed: {e}")
        finally:
            # Clear tokens regardless of logout success
            self._http_client._tokens = None
            self._http_client._token_expires_at = None
            self._http_client._refresh_expires_at = None

        logger.info("Logged out")

    async def refresh_tokens(self) -> AuthTokens:
        """Refresh authentication tokens.

        Returns
        -------
        AuthTokens
            New authentication tokens.

        Raises
        ------
        AuthenticationError
            If token refresh fails.
        """
        return await self._http_client.refresh_tokens()

    async def get_current_user(self) -> User:
        """Get the current authenticated user.

        Returns
        -------
        User
            Current user information.
        """
        return await self.users.get_current()

    def is_authenticated(self) -> bool:
        """Check if client is currently authenticated.

        Returns
        -------
        bool
            True if authenticated, False otherwise.
        """
        return self._http_client._tokens is not None and self._http_client._is_access_token_valid()

    async def close(self) -> None:
        """Close the client and cleanup resources."""
        await self._http_client.close()
        logger.info("SlideInsight client closed")

    # Convenience methods for common operations

    async def create_study_with_cases(
        self,
        study_name: str,
        study_description: str = "",
        cases_data: Optional[list[dict]] = None,
    ) -> tuple[str, list[str]]:
        """Create a study and optionally add cases to it.

        Parameters
        ----------
        study_name : str
            Name of the study.
        study_description : str, default=""
            Description of the study.
        cases_data : list[dict], optional
            List of case data dictionaries with 'name' and optional 'metadata'.

        Returns
        -------
        tuple[str, list[str]]
            Tuple of (study_uid, list_of_case_uids).
        """
        # Create the study
        study = await self.studies.create(
            name=study_name,
            description=study_description,
        )

        case_uids = []
        if cases_data:
            for case_data in cases_data:
                # Create the case
                case = await self.cases.create(
                    name=case_data["name"],
                    metadata=case_data.get("metadata", ""),
                )

                # Add case to study
                await self.studies.add_case(study.study_uid, case.case_uid)
                case_uids.append(case.case_uid)

        logger.info(f"Created study '{study_name}' with {len(case_uids)} cases")
        return study.study_uid, case_uids

    async def create_case_with_slides(
        self,
        case_name: str,
        slides_data: list[dict],
        study_uid: Optional[str] = None,
        case_metadata: str = "",
    ) -> tuple[str, list[str]]:
        """Create a case with slides and optionally add to a study.

        Parameters
        ----------
        case_name : str
            Name of the case.
        slides_data : list[dict]
            List of slide data dictionaries with 'slide_uri' and optional 'slide_name'.
        study_uid : str, optional
            Optional study to add the case to.
        case_metadata : str, default=""
            Case metadata.

        Returns
        -------
        tuple[str, list[str]]
            Tuple of (case_uid, list_of_slide_uids).
        """
        # Create the case
        case = await self.cases.create(
            name=case_name,
            metadata=case_metadata,
        )

        # Add slides to the case
        slide_uids = []
        for slide_data in slides_data:
            slide = await self.cases.add_slide(
                case_uid=case.case_uid,
                slide_uri=slide_data["slide_uri"],
                slide_name=slide_data.get("slide_name"),
                slide_id=slide_data.get("slide_id"),
            )
            slide_uids.append(slide.slide_uid)

        # Add case to study if specified
        if study_uid:
            await self.studies.add_case(study_uid, case.case_uid)

        logger.info(f"Created case '{case_name}' with {len(slide_uids)} slides")
        return case.case_uid, slide_uids

    async def bulk_add_raster_annotations(
        self,
        annotations_data: list[dict],
    ) -> list[str]:
        """Bulk add raster annotations to slides.

        Parameters
        ----------
        annotations_data : list[dict]
            List of annotation data dictionaries with:
            - slide_uid: Slide UID
            - mask_uri: Path to mask file
            - mask_name: Optional mask name
            - actor_type: Optional actor type
            - labels: Optional labels dict
            - metadata: Optional metadata dict

        Returns
        -------
        list[str]
            List of created annotation UIDs.
        """
        annotation_uids = []

        for annotation_data in annotations_data:
            annotation = await self.slides.add_raster_annotation(
                slide_uid=annotation_data["slide_uid"],
                mask_uri=annotation_data["mask_uri"],
                mask_name=annotation_data.get("mask_name"),
                actor_type=annotation_data.get("actor_type", "user"),
                labels=annotation_data.get("labels"),
                metadata=annotation_data.get("metadata"),
            )
            annotation_uids.append(annotation.raster_uid)

        logger.info(f"Created {len(annotation_uids)} raster annotations")
        return annotation_uids
