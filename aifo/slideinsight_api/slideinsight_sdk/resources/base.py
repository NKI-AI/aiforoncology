"""
Base resource manager for SlideInsight SDK resources.

Provides common functionality like pagination, search, and CRUD operations.
"""

from __future__ import annotations

import asyncio
from typing import Any, AsyncIterator, Generic, Optional, TypeVar, Union, Type, Callable
from urllib.parse import urlencode

from ..exceptions import ValidationError
from ..models import PaginatedResponse, PaginationInfo

# Generic type for resource models
T = TypeVar("T")


class BaseResourceManager(Generic[T]):
    """Base class for resource managers.

    Parameters
    ----------
    http_client : HTTPClient
        HTTP client instance for making API requests.
    base_endpoint : str
        Base API endpoint for this resource.
    model_class : Type[T]
        Model class for this resource.
    list_response_key : str, optional
        Key in response containing list of items (if different from resource name).
    """

    def __init__(
        self,
        http_client,
        base_endpoint: str,
        model_class: Type[T],
        list_response_key: str = None,
    ):
        """Initialize resource manager.

        Parameters
        ----------
        http_client : HTTPClient
            HTTP client instance for making API requests.
        base_endpoint : str
            Base API endpoint for this resource.
        model_class : Type[T]
            Model class for this resource.
        list_response_key : str, optional
            Key in response containing list of items (if different from resource name).
        """
        self._http_client = http_client
        self._base_endpoint = base_endpoint.rstrip("/")
        self._model_class = model_class
        self._list_response_key = list_response_key

    @property
    def http_client(self):
        """Get the HTTP client instance.

        Returns
        -------
        HTTPClient
            The HTTP client instance.
        """
        return self._http_client

    @property
    def base_endpoint(self) -> str:
        """Get the base endpoint for this resource.

        Returns
        -------
        str
            The base API endpoint.
        """
        return self._base_endpoint

    @property
    def model_class(self) -> Type[T]:
        """Get the model class for this resource.

        Returns
        -------
        Type[T]
            The model class.
        """
        return self._model_class

    def _build_endpoint(self, path: str = "") -> str:
        """Build full endpoint path.

        Parameters
        ----------
        path : str, default=""
            Additional path to append to base endpoint.

        Returns
        -------
        str
            Complete endpoint path.
        """
        if not path:
            return self._base_endpoint
        return f"{self._base_endpoint}/{path.lstrip('/')}"

    def _build_params(
        self,
        page: Optional[int] = None,
        limit: Optional[int] = None,
        search: Optional[str] = None,
        sort: Optional[str] = None,
        sort_dir: Optional[str] = None,
        **filters,
    ) -> dict[str, Any]:
        """Build query parameters for requests.

        Parameters
        ----------
        page : int, optional
            Page number for pagination.
        limit : int, optional
            Number of items per page.
        search : str, optional
            Search query string.
        sort : str, optional
            Field to sort by.
        sort_dir : str, optional
            Sort direction ('asc' or 'desc').
        **filters : Any
            Additional filter parameters.

        Returns
        -------
        dict[str, Any]
            Query parameters dictionary.
        """
        params = {}

        if page is not None:
            params["page"] = page
        if limit is not None:
            params["limit"] = limit
        if search:
            params["q"] = search
        if sort:
            params["sort"] = sort
        if sort_dir:
            params["dir"] = sort_dir

        # Add any additional filters
        for key, value in filters.items():
            if value is not None:
                params[key] = value

        return params

    def _parse_list_response(self, response_data: dict[str, Any]) -> PaginatedResponse[T]:
        """Parse paginated list response.

        Parameters
        ----------
        response_data : dict[str, Any]
            Raw API response data.

        Returns
        -------
        PaginatedResponse[T]
            Parsed paginated response with items and pagination info.

        Raises
        ------
        ValueError
            If the response format is not recognized.
        """
        # Determine the key containing the list of items
        if self._list_response_key:
            items_key = self._list_response_key
        else:
            # Try common keys
            for key in ["items", "data", "results"]:
                if key in response_data:
                    items_key = key
                    break
            else:
                # Fallback: look for keys that contain lists
                for key, value in response_data.items():
                    if isinstance(value, list) and key != "pagination":
                        items_key = key
                        break
                else:
                    raise ValueError(f"Could not find items list in response: {list(response_data.keys())}")

        # Parse items
        items_data = response_data.get(items_key, [])
        items = [self._model_class.from_dict(item) for item in items_data]

        # Parse pagination info
        pagination_data = response_data.get("pagination", {})
        pagination = PaginationInfo(
            page=pagination_data.get("page", 1),
            limit=pagination_data.get("limit", len(items)),
            total=pagination_data.get("total", len(items)),
            total_pages=pagination_data.get("totalPages", 1),
            has_next=pagination_data.get("hasNext", False),
            has_prev=pagination_data.get("hasPrev", False),
        )

        return PaginatedResponse(items=items, pagination=pagination)

    async def list(
        self,
        page: int = 1,
        limit: int = 100,
        search: Optional[str] = None,
        sort: Optional[str] = None,
        sort_dir: Optional[str] = None,
        **filters,
    ) -> PaginatedResponse[T]:
        """List resources with pagination and filtering.

        Parameters
        ----------
        page : int, default=1
            Page number (1-based).
        limit : int, default=100
            Number of items per page.
        search : str, optional
            Search query string.
        sort : str, optional
            Field to sort by.
        sort_dir : str, optional
            Sort direction ('asc' or 'desc').
        **filters : Any
            Additional filter parameters.

        Returns
        -------
        PaginatedResponse[T]
            Paginated response with items and pagination info.

        Raises
        ------
        ValidationError
            If page or limit parameters are invalid.
        """
        if page < 1:
            raise ValidationError("Page must be >= 1")
        if limit < 1 or limit > 1000:
            raise ValidationError("Limit must be between 1 and 1000")

        params = self._build_params(
            page=page,
            limit=limit,
            search=search,
            sort=sort,
            sort_dir=sort_dir,
            **filters,
        )

        response_data = await self._http_client.request("GET", self._base_endpoint, params=params)
        return self._parse_list_response(response_data)

    async def list_all(
        self,
        page_size: int = 100,
        search: Optional[str] = None,
        sort: Optional[str] = None,
        sort_dir: Optional[str] = None,
        **filters,
    ) -> list[T]:
        """List all resources by automatically handling pagination.

        Parameters
        ----------
        page_size : int, default=100
            Number of items per page.
        search : str, optional
            Search query string.
        sort : str, optional
            Field to sort by.
        sort_dir : str, optional
            Sort direction ('asc' or 'desc').
        **filters : Any
            Additional filter parameters.

        Returns
        -------
        list[T]
            List of all resources matching the criteria.
        """
        all_items = []
        page = 1

        while True:
            response = await self.list(
                page=page,
                limit=page_size,
                search=search,
                sort=sort,
                sort_dir=sort_dir,
                **filters,
            )

            all_items.extend(response.items)

            if not response.pagination.has_next:
                break

            page += 1

        return all_items

    async def list_paginated(
        self,
        page_size: int = 100,
        search: Optional[str] = None,
        sort: Optional[str] = None,
        sort_dir: Optional[str] = None,
        **filters,
    ) -> AsyncIterator[PaginatedResponse[T]]:
        """Async iterator for paginated results.

        Parameters
        ----------
        page_size : int, default=100
            Number of items per page.
        search : str, optional
            Search query string.
        sort : str, optional
            Field to sort by.
        sort_dir : str, optional
            Sort direction ('asc' or 'desc').
        **filters : Any
            Additional filter parameters.

        Yields
        ------
        PaginatedResponse[T]
            Paginated responses, one per page.
        """
        page = 1

        while True:
            response = await self.list(
                page=page,
                limit=page_size,
                search=search,
                sort=sort,
                sort_dir=sort_dir,
                **filters,
            )

            yield response

            if not response.pagination.has_next:
                break

            page += 1

    async def get(self, resource_id: str) -> T:
        """Get a single resource by ID.

        Parameters
        ----------
        resource_id : str
            Unique identifier for the resource.

        Returns
        -------
        T
            The requested resource instance.
        """
        endpoint = self._build_endpoint(resource_id)
        response_data = await self._http_client.request("GET", endpoint)
        return self._model_class.from_dict(response_data)

    async def create(self, data: dict[str, Any]) -> T:
        """Create a new resource.

        Parameters
        ----------
        data : dict[str, Any]
            Resource data for creation.

        Returns
        -------
        T
            The created resource instance.
        """
        response_data = await self._http_client.request("POST", self._base_endpoint, json=data)
        return self._model_class.from_dict(response_data)

    async def update(self, resource_id: str, data: dict[str, Any]) -> T:
        """Update an existing resource.

        Parameters
        ----------
        resource_id : str
            Unique identifier for the resource.
        data : dict[str, Any]
            Updated resource data.

        Returns
        -------
        T
            The updated resource instance.
        """
        endpoint = self._build_endpoint(resource_id)
        response_data = await self._http_client.request("PUT", endpoint, json=data)
        return self._model_class.from_dict(response_data)

    async def delete(self, resource_id: str) -> None:
        """Delete a resource.

        Parameters
        ----------
        resource_id : str
            Unique identifier for the resource to delete.
        """
        endpoint = self._build_endpoint(resource_id)
        await self._http_client.request("DELETE", endpoint)

    async def count(self) -> int:
        """Get total count of resources.

        Returns
        -------
        int
            Total number of resources.
        """
        endpoint = self._build_endpoint("count")
        response_data = await self._http_client.request("GET", endpoint)
        return response_data.get("count", 0)


class PaginationHelper:
    """Helper class for pagination operations."""

    @staticmethod
    def iter_pages(
        current_page: int,
        total_pages: int,
        edge_pages: int = 2,
        center_pages: int = 5,
    ) -> list[Union[int, None]]:
        """Generate page numbers for pagination UI.

        Parameters
        ----------
        current_page : int
            Current page number.
        total_pages : int
            Total number of pages.
        edge_pages : int, default=2
            Number of pages to show at start/end.
        center_pages : int, default=5
            Number of pages to show around current page.

        Returns
        -------
        list[Union[int, None]]
            List of page numbers (None represents ellipsis).
        """
        if total_pages <= edge_pages * 2 + center_pages:
            return list(range(1, total_pages + 1))

        pages = []

        # Start pages
        for i in range(1, min(edge_pages + 1, total_pages + 1)):
            pages.append(i)

        # Calculate center range
        center_start = max(current_page - center_pages // 2, edge_pages + 1)
        center_end = min(current_page + center_pages // 2, total_pages - edge_pages)

        # Add ellipsis if needed
        if center_start > edge_pages + 1:
            pages.append(None)

        # Add center pages
        for i in range(center_start, center_end + 1):
            if i not in pages:
                pages.append(i)

        # Add ellipsis if needed
        if center_end < total_pages - edge_pages:
            pages.append(None)

        # End pages
        for i in range(max(total_pages - edge_pages + 1, center_end + 1), total_pages + 1):
            if i not in pages:
                pages.append(i)

        return pages
