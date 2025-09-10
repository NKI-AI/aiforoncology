"""
Cases resource manager for SlideInsight SDK.

Provides operations for managing cases including creation, updates,
slide management, and soft deletion.
"""

from __future__ import annotations

from typing import Optional

from .base import BaseResourceManager
from ..models import Case, Slide, PaginatedResponse, PaginationInfo


class CasesManager(BaseResourceManager[Case]):
    """Manager for case resources."""

    def __init__(self, http_client):
        super().__init__(
            http_client=http_client, base_endpoint="/api/v1/cases", model_class=Case, list_response_key="cases"
        )

    async def create(
        self,
        name: str,
        metadata: str = "",
    ) -> Case:
        """Create a new case.

        Args:
            name: Case name
            metadata: Case metadata (JSON string)

        Returns:
            Created case
        """
        data = {
            "name": name,
            "metadata": metadata,
        }

        response_data = await self.http_client.request("POST", self.base_endpoint, json=data)
        return Case.from_dict(response_data)

    async def update(
        self,
        case_uid: str,
        name: Optional[str] = None,
        metadata: Optional[str] = None,
    ) -> Case:
        """Update an existing case.

        Args:
            case_uid: Case UID
            name: New case name
            metadata: New case metadata

        Returns:
            Updated case
        """
        data = {}
        if name is not None:
            data["name"] = name
        if metadata is not None:
            data["metadata"] = metadata

        endpoint = self._build_endpoint(case_uid)
        response_data = await self.http_client.request("PUT", endpoint, json=data)
        return Case.from_dict(response_data)

    async def get_slides(
        self,
        case_uid: str,
        page: int = 1,
        limit: int = 100,
        search: Optional[str] = None,
        sort: Optional[str] = None,
        sort_dir: Optional[str] = None,
        **filters,
    ) -> PaginatedResponse[Slide]:
        """Get slides in a case.

        Args:
            case_uid: Case UID
            page: Page number
            limit: Items per page
            search: Search query
            sort: Sort field
            sort_dir: Sort direction
            **filters: Additional filters

        Returns:
            Paginated slides response
        """
        endpoint = f"/api/v1/cases/{case_uid}/slides"
        params = self._build_params(
            page=page,
            limit=limit,
            search=search,
            sort=sort,
            sort_dir=sort_dir,
            **filters,
        )

        response_data = await self.http_client.request("GET", endpoint, params=params)

        # Handle the response which might not follow standard pagination format
        if "slides" in response_data:
            slides_data = response_data["slides"]
            slides = [Slide.from_dict(slide) for slide in slides_data]

            # Create minimal pagination info if not provided
            pagination = PaginationInfo(
                page=1,
                limit=len(slides),
                total=len(slides),
                total_pages=1,
                has_next=False,
                has_prev=False,
            )

            return PaginatedResponse(items=slides, pagination=pagination)
        else:
            # Assume standard paginated response
            slides_data = response_data.get("items", response_data.get("data", []))
            slides = [Slide.from_dict(slide) for slide in slides_data]

            pagination_data = response_data.get("pagination", {})
            pagination = PaginationInfo(
                page=pagination_data.get("page", 1),
                limit=pagination_data.get("limit", len(slides)),
                total=pagination_data.get("total", len(slides)),
                total_pages=pagination_data.get("totalPages", 1),
                has_next=pagination_data.get("hasNext", False),
                has_prev=pagination_data.get("hasPrev", False),
            )

            return PaginatedResponse(items=slides, pagination=pagination)

    async def add_slide(
        self,
        case_uid: str,
        slide_uri: str,
        slide_name: Optional[str] = None,
        slide_id: Optional[str] = None,
    ) -> Slide:
        """Add a slide to a case.

        Args:
            case_uid: Case UID
            slide_uri: URI/path to the slide file
            slide_name: Human-readable slide name
            slide_id: Slide identifier

        Returns:
            Created slide
        """
        endpoint = f"/api/v1/cases/{case_uid}/slides"
        data = {
            "slideUri": slide_uri,
            "caseUid": case_uid,  # Required by Go handler's SlideCreationInput struct
        }

        if slide_name is not None:
            data["slideName"] = slide_name
        if slide_id is not None:
            data["slideId"] = slide_id

        response_data = await self.http_client.request("POST", endpoint, json=data)
        return Slide.from_dict(response_data)

    async def get_neighbors(
        self,
        study_uid: str,
        case_uid: str,
        page: int = 1,
        limit: int = 10,
    ) -> PaginatedResponse[Case]:
        """Get neighboring cases within a study.

        Args:
            study_uid: Study UID
            case_uid: Case UID to find neighbors for
            page: Page number
            limit: Items per page

        Returns:
            Paginated cases response
        """
        endpoint = f"/api/v1/studies/{study_uid}/cases/{case_uid}/neighbors"
        params = self._build_params(page=page, limit=limit)

        response_data = await self.http_client.request("GET", endpoint, params=params)

        # Parse the response
        cases_data = response_data.get("cases", [])
        cases = [Case.from_dict(case) for case in cases_data]

        pagination_data = response_data.get("pagination", {})
        pagination = PaginationInfo(
            page=pagination_data.get("page", 1),
            limit=pagination_data.get("limit", len(cases)),
            total=pagination_data.get("total", len(cases)),
            total_pages=pagination_data.get("totalPages", 1),
            has_next=pagination_data.get("hasNext", False),
            has_prev=pagination_data.get("hasPrev", False),
        )

        return PaginatedResponse(items=cases, pagination=pagination)

    async def soft_delete(self, case_uid: str) -> None:
        """Soft delete a case.

        Args:
            case_uid: Case UID to delete
        """
        await self.delete(case_uid)

    async def restore(self, case_uid: str) -> Case:
        """Restore a soft-deleted case.

        Args:
            case_uid: Case UID to restore

        Returns:
            Restored case
        """
        endpoint = self._build_endpoint(f"{case_uid}/restore")
        response_data = await self.http_client.request("POST", endpoint)
        return Case.from_dict(response_data)

    async def get_deleted(
        self,
        page: int = 1,
        limit: int = 100,
    ) -> PaginatedResponse[Case]:
        """Get soft-deleted cases.

        Args:
            page: Page number
            limit: Items per page

        Returns:
            Paginated deleted cases response
        """
        endpoint = "/api/v1/cases/deleted"
        params = self._build_params(page=page, limit=limit)

        response_data = await self.http_client.request("GET", endpoint, params=params)
        return self._parse_list_response(response_data)
