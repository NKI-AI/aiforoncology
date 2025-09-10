"""
Slides resource manager for SlideInsight SDK.

Provides operations for managing slides including creation, metadata retrieval,
tile access, and annotations management.
"""

from __future__ import annotations

from typing import Optional

from .base import BaseResourceManager
from ..models import (
    Slide,
    SlideMetadata,
    SlideTile,
    SlideAnnotationsOverview,
    RasterAnnotation,
    VectorAnnotation,
    PaginatedResponse,
    PaginationInfo,
)


class SlidesManager(BaseResourceManager[Slide]):
    """Manager for slide resources."""

    def __init__(self, http_client):
        super().__init__(
            http_client=http_client, base_endpoint="/api/v1/slides", model_class=Slide, list_response_key="slides"
        )

    async def create(
        self,
        slide_uri: str,
        slide_name: Optional[str] = None,
        slide_id: Optional[str] = None,
        case_uid: Optional[str] = None,
    ) -> Slide:
        """Create a new slide.

        Args:
            slide_uri: URI/path to the slide file
            slide_name: Human-readable slide name
            slide_id: Slide identifier
            case_uid: Case UID to associate with (optional)

        Returns:
            Created slide
        """
        data = {
            "slideUri": slide_uri,
        }

        if slide_name is not None:
            data["slideName"] = slide_name
        if slide_id is not None:
            data["slideId"] = slide_id
        if case_uid is not None:
            data["caseUid"] = case_uid

        response_data = await self.http_client.request("POST", self.base_endpoint, json=data)
        return Slide.from_dict(response_data)

    async def get_metadata(self, slide_uid: str) -> SlideMetadata:
        """Get slide metadata including dimensions and properties.

        Args:
            slide_uid: Slide UID

        Returns:
            Slide metadata
        """
        endpoint = self._build_endpoint(f"{slide_uid}/metadata")
        response_data = await self.http_client.request("GET", endpoint)
        return SlideMetadata.from_dict(response_data)

    async def get_tile(
        self,
        slide_uid: str,
        z: int,
        x: int,
        y: int,
        format: str = "jpeg",
    ) -> SlideTile:
        """Get a tile from a slide.

        Args:
            slide_uid: Slide UID
            z: Zoom level
            x: Tile X coordinate
            y: Tile Y coordinate
            format: Image format ('jpeg' or 'png')

        Returns:
            Slide tile with image data
        """
        endpoint = f"/api/v1/slides/{slide_uid}/tiles/{z}/{x}/{y}.{format}"
        content_type, image_data = await self.http_client.request_binary("GET", endpoint)
        return SlideTile.from_response(content_type, image_data)

    async def get_annotations_overview(self, slide_uid: str) -> SlideAnnotationsOverview:
        """Get overview of annotations for a slide.

        Args:
            slide_uid: Slide UID

        Returns:
            Annotations overview with counts and URLs
        """
        endpoint = self._build_endpoint(f"{slide_uid}/annotations")
        response_data = await self.http_client.request("GET", endpoint)
        return SlideAnnotationsOverview.from_dict(response_data)

    async def get_raster_annotations(
        self,
        slide_uid: str,
        page: int = 1,
        limit: int = 100,
    ) -> list[RasterAnnotation]:
        """Get raster annotations (masks) for a slide.

        Args:
            slide_uid: Slide UID
            page: Page number
            limit: Items per page

        Returns:
            List of raster annotations
        """
        endpoint = f"/api/v1/slides/{slide_uid}/annotations/raster"
        params = self._build_params(page=page, limit=limit)

        response_data = await self.http_client.request("GET", endpoint, params=params)

        # Handle response format (could be different structures)
        if "masks" in response_data:
            masks_data = response_data["masks"]
        elif "rasterAnnotations" in response_data:
            masks_data = response_data["rasterAnnotations"]
        else:
            masks_data = response_data.get("items", response_data.get("data", []))

        return [RasterAnnotation.from_dict(mask) for mask in masks_data]

    async def add_raster_annotation(
        self,
        slide_uid: str,
        mask_uri: str,
        mask_name: Optional[str] = None,
        actor_type: str = "user",
        actor_id: Optional[int] = None,
        labels: Optional[dict] = None,
        metadata: Optional[dict] = None,
    ) -> RasterAnnotation:
        """Add a raster annotation (mask) to a slide.

        Args:
            slide_uid: Slide UID
            mask_uri: URI/path to the mask file
            mask_name: Human-readable mask name
            actor_type: Type of actor ('user' or 'model')
            actor_id: ID of the actor
            labels: Annotation labels
            metadata: Additional metadata

        Returns:
            Created raster annotation
        """
        endpoint = f"/api/v1/slides/{slide_uid}/annotations/raster"
        data = {
            "maskUri": mask_uri,
            "slideUid": slide_uid,
            "actorType": actor_type,
        }

        if mask_name is not None:
            data["maskName"] = mask_name
        if actor_id is not None:
            data["actorId"] = actor_id
        if labels is not None:
            data["labels"] = labels
        if metadata is not None:
            data["metadata"] = metadata

        response_data = await self.http_client.request("POST", endpoint, json=data)
        return RasterAnnotation.from_dict(response_data)

    async def get_vector_annotations(
        self,
        slide_uid: str,
        page: int = 1,
        limit: int = 100,
    ) -> list[VectorAnnotation]:
        """Get vector annotations for a slide.

        Args:
            slide_uid: Slide UID
            page: Page number
            limit: Items per page

        Returns:
            List of vector annotations
        """
        endpoint = f"/api/v1/slides/{slide_uid}/annotations/vector"
        params = self._build_params(page=page, limit=limit)

        response_data = await self.http_client.request("GET", endpoint, params=params)

        # Handle response format
        if "vectorAnnotations" in response_data:
            annotations_data = response_data["vectorAnnotations"]
        else:
            annotations_data = response_data.get("items", response_data.get("data", []))

        return [VectorAnnotation.from_dict(annotation) for annotation in annotations_data]

    async def add_vector_annotation(
        self,
        slide_uid: str,
        file_uri: str,
        name: Optional[str] = None,
        format: str = "geojson",
        actor_type: str = "user",
        actor_id: Optional[int] = None,
        labels: Optional[dict] = None,
        metadata: Optional[dict] = None,
    ) -> VectorAnnotation:
        """Add a vector annotation to a slide.

        Args:
            slide_uid: Slide UID
            file_uri: URI/path to the annotation file
            name: Human-readable annotation name
            format: Annotation format ('geojson', 'protobuf', 'zarr')
            actor_type: Type of actor ('user' or 'model')
            actor_id: ID of the actor
            labels: Annotation labels
            metadata: Additional metadata

        Returns:
            Created vector annotation
        """
        endpoint = f"/api/v1/slides/{slide_uid}/annotations/vector"
        data = {
            "fileUri": file_uri,
            "slideUid": slide_uid,
            "format": format,
            "actorType": actor_type,
        }

        if name is not None:
            data["name"] = name
        if actor_id is not None:
            data["actorId"] = actor_id
        if labels is not None:
            data["labels"] = labels
        if metadata is not None:
            data["metadata"] = metadata

        response_data = await self.http_client.request("POST", endpoint, json=data)
        return VectorAnnotation.from_dict(response_data)

    async def get_raster_annotation_tile(
        self,
        slide_uid: str,
        mask_id: str,
        z: int,
        x: int,
        y: int,
        format: str = "png",
    ) -> SlideTile:
        """Get a tile from a raster annotation.

        Args:
            slide_uid: Slide UID
            mask_id: Mask/annotation ID
            z: Zoom level
            x: Tile X coordinate
            y: Tile Y coordinate
            format: Image format ('png' recommended for masks)

        Returns:
            Annotation tile with image data
        """
        endpoint = f"/api/v1/slides/{slide_uid}/annotations/raster/{mask_id}/tiles/{z}/{x}/{y}.{format}"
        content_type, image_data = await self.http_client.request_binary("GET", endpoint)
        return SlideTile.from_response(content_type, image_data)

    async def get_vector_annotation_file(
        self,
        slide_uid: str,
        vector_id: str,
    ) -> bytes:
        """Get the raw file data for a vector annotation.

        Args:
            slide_uid: Slide UID
            vector_id: Vector annotation ID

        Returns:
            Raw file data
        """
        endpoint = f"/api/v1/slides/{slide_uid}/annotations/vector/{vector_id}/file"
        content_type, file_data = await self.http_client.request_binary("GET", endpoint)
        return file_data
