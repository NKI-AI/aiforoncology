"""
Studies resource manager for SlideInsight SDK.

Provides operations for managing studies including creation, updates,
case management, and metadata retrieval.
"""

from __future__ import annotations

from typing import Optional

from .base import BaseResourceManager
from ..models import Study, StudyMetadata, Case, PaginatedResponse, PaginationInfo


class StudiesManager(BaseResourceManager[Study]):
    """Manager for study resources."""

    def __init__(self, http_client):
        super().__init__(
            http_client=http_client, base_endpoint="/api/v1/studies", model_class=Study, list_response_key="studies"
        )

    async def create(
        self,
        name: str,
        description: str = "",
        metadata: str = "",
        is_published: bool = False,
    ) -> Study:
        """Create a new study.

        Args:
            name: Study name
            description: Study description
            metadata: Study metadata (JSON string)
            is_published: Whether the study is published

        Returns:
            Created study
        """
        data = {
            "name": name,
            "description": description,
            "metadata": metadata,
            "isPublished": is_published,
        }

        response_data = await self.http_client.request("POST", self.base_endpoint, json=data)
        return Study.from_dict(response_data)

    async def update(
        self,
        study_uid: str,
        name: Optional[str] = None,
        description: Optional[str] = None,
        metadata: Optional[str] = None,
        is_published: Optional[bool] = None,
    ) -> Study:
        """Update an existing study.

        Args:
            study_uid: Study UID
            name: New study name
            description: New study description
            metadata: New study metadata
            is_published: New published status

        Returns:
            Updated study
        """
        data = {}
        if name is not None:
            data["name"] = name
        if description is not None:
            data["description"] = description
        if metadata is not None:
            data["metadata"] = metadata
        if is_published is not None:
            data["isPublished"] = is_published

        endpoint = self._build_endpoint(study_uid)
        response_data = await self.http_client.request("PUT", endpoint, json=data)
        return Study.from_dict(response_data)

    async def get_metadata(self, study_uid: str) -> StudyMetadata:
        """Get study metadata including statistics.

        Args:
            study_uid: Study UID

        Returns:
            Study metadata
        """
        endpoint = self._build_endpoint(f"{study_uid}/metadata")
        response_data = await self.http_client.request("GET", endpoint)
        return StudyMetadata.from_dict(response_data)

    async def get_cases(
        self,
        study_uid: str,
        page: int = 1,
        limit: int = 100,
        search: Optional[str] = None,
        sort: Optional[str] = None,
        sort_dir: Optional[str] = None,
        **filters,
    ) -> PaginatedResponse[Case]:
        """Get cases in a study.

        Args:
            study_uid: Study UID
            page: Page number
            limit: Items per page
            search: Search query
            sort: Sort field
            sort_dir: Sort direction
            **filters: Additional filters

        Returns:
            Paginated cases response
        """
        endpoint = self._build_endpoint(f"{study_uid}/cases")
        params = self._build_params(
            page=page,
            limit=limit,
            search=search,
            sort=sort,
            sort_dir=sort_dir,
            **filters,
        )

        response_data = await self.http_client.request("GET", endpoint, params=params)

        # Parse the response manually since it's not the standard studies format
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

    async def add_case(self, study_uid: str, case_uid: str) -> None:
        """Add a case to a study.

        Args:
            study_uid: Study UID
            case_uid: Case UID to add
        """
        endpoint = self._build_endpoint(f"{study_uid}/cases")
        data = {"caseUid": case_uid}
        await self.http_client.request("POST", endpoint, json=data)

    async def remove_case(self, study_uid: str, case_uid: str) -> None:
        """Remove a case from a study.

        Args:
            study_uid: Study UID
            case_uid: Case UID to remove
        """
        endpoint = self._build_endpoint(f"{study_uid}/cases/{case_uid}")
        await self.http_client.request("DELETE", endpoint)

    async def soft_delete(self, study_uid: str) -> None:
        """Soft delete a study.

        Args:
            study_uid: Study UID to delete
        """
        await self.delete(study_uid)

    async def restore(self, study_uid: str) -> Study:
        """Restore a soft-deleted study.

        Args:
            study_uid: Study UID to restore

        Returns:
            Restored study
        """
        endpoint = self._build_endpoint(f"{study_uid}/restore")
        response_data = await self.http_client.request("POST", endpoint)
        return Study.from_dict(response_data)
