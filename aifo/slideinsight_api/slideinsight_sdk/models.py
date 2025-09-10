"""
Data models for the SlideInsight SDK.

These models correspond to the domain models in the SlideInsight API and provide
type-safe representations of API responses.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from datetime import datetime
from typing import Any, Generic, Optional, TypeVar, Union
from enum import Enum


T = TypeVar("T")


@dataclass
class PaginationInfo:
    """Pagination information returned by API responses.

    Attributes
    ----------
    page : int
        Current page number (1-based).
    limit : int
        Number of items per page.
    total : int
        Total number of items available.
    total_pages : int
        Total number of pages available.
    has_next : bool
        Whether there is a next page available.
    has_prev : bool
        Whether there is a previous page available.
    """

    page: int
    limit: int
    total: int
    total_pages: int
    has_next: bool
    has_prev: bool


@dataclass
class PaginatedResponse(Generic[T]):
    """Generic paginated response wrapper.

    Attributes
    ----------
    items : list[T]
        List of items in the current page.
    pagination : PaginationInfo
        Pagination metadata.
    """

    items: list[T]
    pagination: PaginationInfo


@dataclass
class Study:
    """Study domain model.

    Attributes
    ----------
    study_uid : str
        Unique identifier for the study.
    tenant_uid : str, optional
        Tenant UID this study belongs to.
    creator_uid : str, optional
        UID of the user who created this study.
    name : str
        Human-readable name of the study.
    description : str
        Detailed description of the study.
    metadata : str
        JSON string containing additional metadata.
    is_published : bool
        Whether the study is published and visible to others.
    deleted_at : str, optional
        ISO timestamp when the study was soft deleted.
    deleted_by : int, optional
        User ID who deleted the study.
    created_at : str, optional
        ISO timestamp when the study was created.
    """

    study_uid: str
    tenant_uid: Optional[str] = None
    creator_uid: Optional[str] = None
    name: str = ""
    description: str = ""
    metadata: str = ""
    is_published: bool = False
    deleted_at: Optional[str] = None
    deleted_by: Optional[int] = None
    created_at: Optional[str] = None

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> "Study":
        """Create Study from API response data.

        Parameters
        ----------
        data : dict[str, Any]
            Raw API response data containing study fields.

        Returns
        -------
        Study
            Study instance created from the response data.
        """
        return cls(
            study_uid=data["studyUid"],
            tenant_uid=data.get("tenantUid"),
            creator_uid=data.get("creatorUid"),
            name=data.get("name", ""),
            description=data.get("description", ""),
            metadata=data.get("metadata", ""),
            is_published=data.get("isPublished", False),
            deleted_at=data.get("deletedAt"),
            deleted_by=data.get("deletedBy"),
            created_at=data.get("createdAt"),
        )


@dataclass
class StudyMetadata:
    """Study metadata with statistics.

    Attributes
    ----------
    study_uid : str
        Unique identifier for the study.
    case_count : int
        Number of cases in this study.
    """

    study_uid: str
    case_count: int

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> "StudyMetadata":
        """Create StudyMetadata from API response data.

        Parameters
        ----------
        data : dict[str, Any]
            Raw API response data containing study metadata.

        Returns
        -------
        StudyMetadata
            StudyMetadata instance created from the response data.
        """
        return cls(
            study_uid=data["studyUid"],
            case_count=data.get("caseCount", 0),
        )


@dataclass
class Case:
    """Case domain model.

    Attributes
    ----------
    case_uid : str
        Unique identifier for the case.
    tenant_uid : str, optional
        Tenant UID this case belongs to.
    creator_uid : str, optional
        UID of the user who created this case.
    name : str
        Human-readable name of the case.
    metadata : str
        JSON string containing additional metadata.
    deleted_at : str, optional
        ISO timestamp when the case was soft deleted.
    deleted_by : int, optional
        User ID who deleted the case.
    created_at : str, optional
        ISO timestamp when the case was created.
    has_vector_annotations : bool, optional
        Whether this case has vector annotations (when retrieved with annotations).
    has_raster_annotations : bool, optional
        Whether this case has raster annotations (when retrieved with annotations).
    """

    case_uid: str
    tenant_uid: Optional[str] = None
    creator_uid: Optional[str] = None
    name: str = ""
    metadata: str = ""
    deleted_at: Optional[str] = None
    deleted_by: Optional[int] = None
    created_at: Optional[str] = None
    # Annotation flags (when retrieved with annotations)
    has_vector_annotations: Optional[bool] = None
    has_raster_annotations: Optional[bool] = None

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> "Case":
        """Create Case from API response data.

        Parameters
        ----------
        data : dict[str, Any]
            Raw API response data containing case fields.

        Returns
        -------
        Case
            Case instance created from the response data.
        """
        return cls(
            case_uid=data["caseUid"],
            tenant_uid=data.get("tenantUid"),
            creator_uid=data.get("creatorUid"),
            name=data.get("name", ""),
            metadata=data.get("metadata", ""),
            deleted_at=data.get("deletedAt"),
            deleted_by=data.get("deletedBy"),
            created_at=data.get("createdAt"),
            has_vector_annotations=data.get("hasVectorAnnotations"),
            has_raster_annotations=data.get("hasRasterAnnotations"),
        )


@dataclass
class Slide:
    """Slide domain model.

    Attributes
    ----------
    slide_uid : str
        Unique identifier for the slide.
    case_uid : str, optional
        Case UID this slide belongs to.
    slide_id : str, optional
        External slide identifier.
    slide_name : str, optional
        Human-readable name of the slide.
    slide_uri : str
        URI or path to the slide file.
    slide_width : int, optional
        Width of the slide in pixels.
    slide_height : int, optional
        Height of the slide in pixels.
    slide_mpp : float, optional
        Microns per pixel resolution.
    metadata : str
        JSON string containing additional metadata.
    deleted_at : str, optional
        ISO timestamp when the slide was soft deleted.
    deleted_by : int, optional
        User ID who deleted the slide.
    created_at : str, optional
        ISO timestamp when the slide was created.
    """

    slide_uid: str
    case_uid: Optional[str] = None
    slide_id: Optional[str] = None
    slide_name: Optional[str] = None
    slide_uri: str = ""
    slide_width: Optional[int] = None
    slide_height: Optional[int] = None
    slide_mpp: Optional[float] = None
    metadata: str = ""
    deleted_at: Optional[str] = None
    deleted_by: Optional[int] = None
    created_at: Optional[str] = None

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> "Slide":
        """Create Slide from API response data.

        Parameters
        ----------
        data : dict[str, Any]
            Raw API response data containing slide fields.

        Returns
        -------
        Slide
            Slide instance created from the response data.
        """
        return cls(
            slide_uid=data["slideUid"],
            case_uid=data.get("caseUid"),
            slide_id=data.get("slideId"),
            slide_name=data.get("slideName"),
            slide_uri=data.get("slideUri", ""),
            slide_width=data.get("slideWidth"),
            slide_height=data.get("slideHeight"),
            slide_mpp=data.get("slideMpp"),
            metadata=data.get("metadata", ""),
            deleted_at=data.get("deletedAt"),
            deleted_by=data.get("deletedBy"),
            created_at=data.get("createdAt"),
        )


@dataclass
class SlideMetadata:
    """Slide metadata with detailed information.

    Attributes
    ----------
    slide_uid : str
        Unique identifier for the slide.
    slide_width : int, optional
        Width of the slide in pixels.
    slide_height : int, optional
        Height of the slide in pixels.
    slide_mpp : float, optional
        Microns per pixel resolution.
    levels : int, optional
        Number of pyramid levels in the slide.
    tile_size : int, optional
        Size of tiles in pixels.
    """

    slide_uid: str
    slide_width: Optional[int] = None
    slide_height: Optional[int] = None
    slide_mpp: Optional[float] = None
    levels: Optional[int] = None
    tile_size: Optional[int] = None

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> "SlideMetadata":
        """Create SlideMetadata from API response data.

        Parameters
        ----------
        data : dict[str, Any]
            Raw API response data containing slide metadata.

        Returns
        -------
        SlideMetadata
            SlideMetadata instance created from the response data.
        """
        return cls(
            slide_uid=data["slideUid"],
            slide_width=data.get("slideWidth"),
            slide_height=data.get("slideHeight"),
            slide_mpp=data.get("slideMpp"),
            levels=data.get("levels"),
            tile_size=data.get("tileSize"),
        )


@dataclass
class SlideTile:
    """Slide tile image data.

    Attributes
    ----------
    content_type : str
        MIME type of the image data.
    image : bytes
        Raw image data in the specified format.
    """

    content_type: str
    image: bytes

    @classmethod
    def from_response(cls, content_type: str, image_data: bytes) -> "SlideTile":
        """Create SlideTile from HTTP response data.

        Parameters
        ----------
        content_type : str
            Content-Type header from the response.
        image_data : bytes
            Raw image data from the response body.

        Returns
        -------
        SlideTile
            SlideTile instance containing the image data.
        """
        return cls(content_type=content_type, image=image_data)


class UserStatus(Enum):
    """User status enumeration.

    Attributes
    ----------
    ACTIVE : str
        User is active and can log in.
    INACTIVE : str
        User is inactive and cannot log in.
    SUSPENDED : str
        User is suspended and cannot log in.
    """

    ACTIVE = "active"
    INACTIVE = "inactive"
    SUSPENDED = "suspended"


@dataclass
class User:
    """User domain model.

    Attributes
    ----------
    user_uid : str
        Unique identifier for the user.
    tenant_uid : str, optional
        Tenant UID this user belongs to.
    email : str
        Email address of the user.
    first_name : str, optional
        First name of the user.
    last_name : str, optional
        Last name of the user.
    is_active : bool
        Whether the user account is active.
    email_verified : bool
        Whether the user's email has been verified.
    must_reset_password : bool
        Whether the user must reset their password on next login.
    created_at : str, optional
        ISO timestamp when the user was created.
    """

    user_uid: str
    tenant_uid: Optional[str] = None
    email: str = ""
    first_name: Optional[str] = None
    last_name: Optional[str] = None
    is_active: bool = True
    email_verified: bool = False
    must_reset_password: bool = False
    created_at: Optional[str] = None

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> "User":
        """Create User from API response data.

        Parameters
        ----------
        data : dict[str, Any]
            Raw API response data containing user fields.

        Returns
        -------
        User
            User instance created from the response data.
        """
        # Handle minimal response from /api/v1/auth/me endpoint
        if "userUid" not in data and "email" in data:
            # Minimal response - just email and possibly scopes
            return cls(
                user_uid="",  # Not available in minimal response
                tenant_uid=None,
                email=data.get("email", ""),
                first_name=None,
                last_name=None,
                is_active=True,
                email_verified=False,
                must_reset_password=False,
                created_at=None,
            )

        # Full user response
        return cls(
            user_uid=data["userUid"],
            tenant_uid=data.get("tenantUid"),
            email=data.get("email", ""),
            first_name=data.get("firstName"),
            last_name=data.get("lastName"),
            is_active=data.get("isActive", True),
            email_verified=data.get("emailVerified", False),
            must_reset_password=data.get("mustResetPassword", False),
            created_at=data.get("createdAt"),
        )


class TenantStatus(Enum):
    """Tenant status enumeration.

    Attributes
    ----------
    ACTIVE : str
        Tenant is active and operational.
    INACTIVE : str
        Tenant is inactive.
    SUSPENDED : str
        Tenant is suspended.
    PENDING : str
        Tenant is pending activation.
    """

    ACTIVE = "active"
    INACTIVE = "inactive"
    SUSPENDED = "suspended"
    PENDING = "pending"


@dataclass
class TenantDomain:
    """Tenant domain model.

    Attributes
    ----------
    domain : str
        Domain name associated with the tenant.
    is_verified : bool
        Whether the domain has been verified.
    is_primary : bool
        Whether this is the primary domain for the tenant.
    created_at : str, optional
        ISO timestamp when the domain was created.
    updated_at : str, optional
        ISO timestamp when the domain was last updated.
    """

    domain: str
    is_verified: bool = False
    is_primary: bool = False
    created_at: Optional[str] = None
    updated_at: Optional[str] = None

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> "TenantDomain":
        """Create TenantDomain from API response data.

        Parameters
        ----------
        data : dict[str, Any]
            Raw API response data containing domain fields.

        Returns
        -------
        TenantDomain
            TenantDomain instance created from the response data.
        """
        return cls(
            domain=data["domain"],
            is_verified=data.get("isVerified", False),
            is_primary=data.get("isPrimary", False),
            created_at=data.get("createdAt"),
            updated_at=data.get("updatedAt"),
        )


@dataclass
class Tenant:
    """Tenant domain model.

    Attributes
    ----------
    tenant_uid : str
        Unique identifier for the tenant.
    name : str
        Human-readable name of the tenant.
    description : str
        Detailed description of the tenant.
    status : str
        Current status of the tenant.
    created_at : str, optional
        ISO timestamp when the tenant was created.
    domains : list[TenantDomain]
        List of domains associated with this tenant.
    """

    tenant_uid: str
    name: str = ""
    description: str = ""
    status: str = "active"
    created_at: Optional[str] = None
    domains: list[TenantDomain] = field(default_factory=list)

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> "Tenant":
        """Create Tenant from API response data.

        Parameters
        ----------
        data : dict[str, Any]
            Raw API response data containing tenant fields.

        Returns
        -------
        Tenant
            Tenant instance created from the response data.
        """
        domains = []
        if "domains" in data:
            domains = [TenantDomain.from_dict(d) for d in data["domains"]]

        return cls(
            tenant_uid=data["tenantUid"],
            name=data.get("name", ""),
            description=data.get("description", ""),
            status=data.get("status", "active"),
            created_at=data.get("createdAt"),
            domains=domains,
        )


class AnnotationActorType(Enum):
    """Annotation actor type enumeration.

    Attributes
    ----------
    USER : str
        Annotation created by a human user.
    MODEL : str
        Annotation created by an AI model or algorithm.
    """

    USER = "user"
    MODEL = "model"


class AnnotationFormat(Enum):
    """Annotation format enumeration.

    Attributes
    ----------
    TIFF : str
        TIFF image format for raster annotations.
    PNG : str
        PNG image format for raster annotations.
    PROTOBUF : str
        Protocol Buffers format for vector annotations.
    GEOJSON : str
        GeoJSON format for vector annotations.
    ZARR : str
        Zarr format for vector annotations.
    """

    # Raster formats
    TIFF = "tiff"
    PNG = "png"
    # Vector formats
    PROTOBUF = "protobuf"
    GEOJSON = "geojson"
    ZARR = "zarr"


@dataclass
class RasterAnnotation:
    """Raster annotation (mask) domain model.

    Attributes
    ----------
    raster_uid : str
        Unique identifier for the raster annotation.
    slide_uid : str
        UID of the slide this annotation belongs to.
    actor_type : str
        Type of actor that created this annotation ('user' or 'model').
    actor_id : int
        ID of the actor that created this annotation.
    name : str, optional
        Human-readable name of the annotation.
    file_uri : str
        URI or path to the annotation file.
    format : str, optional
        Format of the annotation file.
    labels : dict[str, Any], optional
        Labels associated with this annotation.
    metadata : dict[str, Any], optional
        Additional metadata for this annotation.
    mask_width : int, optional
        Width of the mask in pixels.
    mask_height : int, optional
        Height of the mask in pixels.
    mask_mpp : float, optional
        Microns per pixel resolution of the mask.
    version : int
        Version number of the annotation.
    deleted_at : str, optional
        ISO timestamp when the annotation was soft deleted.
    deleted_by : int, optional
        User ID who deleted the annotation.
    created_at : str, optional
        ISO timestamp when the annotation was created.
    """

    raster_uid: str
    slide_uid: str
    actor_type: str
    actor_id: int
    name: Optional[str] = None
    file_uri: str = ""
    format: Optional[str] = None
    labels: Optional[dict[str, Any]] = None
    metadata: Optional[dict[str, Any]] = None
    mask_width: Optional[int] = None
    mask_height: Optional[int] = None
    mask_mpp: Optional[float] = None
    version: int = 1
    deleted_at: Optional[str] = None
    deleted_by: Optional[int] = None
    created_at: Optional[str] = None

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> "RasterAnnotation":
        """Create RasterAnnotation from API response data.

        Parameters
        ----------
        data : dict[str, Any]
            Raw API response data containing raster annotation fields.

        Returns
        -------
        RasterAnnotation
            RasterAnnotation instance created from the response data.
        """
        return cls(
            raster_uid=data["maskUid"],
            slide_uid=data.get("slideId", ""),
            actor_type=data.get("actorType", "user"),
            actor_id=data.get("actorId", 1),
            name=data["maskName"],
            file_uri=data["maskUri"],
            format=data.get("format"),
            labels=data.get("labels"),
            metadata=data.get("metadata"),
            mask_width=data.get("maskWidth"),
            mask_height=data.get("maskHeight"),
            mask_mpp=data.get("maskMpp"),
            version=data.get("version", 1),
            deleted_at=data.get("deletedAt"),
            deleted_by=data.get("deletedBy"),
            created_at=data.get("createdAt"),
        )


@dataclass
class VectorAnnotation:
    """Vector annotation domain model.

    Attributes
    ----------
    vector_uid : str
        Unique identifier for the vector annotation.
    slide_uid : str
        UID of the slide this annotation belongs to.
    actor_type : str
        Type of actor that created this annotation ('user' or 'model').
    actor_id : int
        ID of the actor that created this annotation.
    name : str, optional
        Human-readable name of the annotation.
    file_uri : str
        URI or path to the annotation file.
    format : str, optional
        Format of the annotation file.
    labels : dict[str, Any], optional
        Labels associated with this annotation.
    metadata : dict[str, Any], optional
        Additional metadata for this annotation.
    version : int
        Version number of the annotation.
    deleted_at : str, optional
        ISO timestamp when the annotation was soft deleted.
    deleted_by : int, optional
        User ID who deleted the annotation.
    created_at : str, optional
        ISO timestamp when the annotation was created.
    """

    vector_uid: str
    slide_uid: str
    actor_type: str
    actor_id: int
    name: Optional[str] = None
    file_uri: str = ""
    format: Optional[str] = None
    labels: Optional[dict[str, Any]] = None
    metadata: Optional[dict[str, Any]] = None
    version: int = 1
    deleted_at: Optional[str] = None
    deleted_by: Optional[int] = None
    created_at: Optional[str] = None

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> "VectorAnnotation":
        """Create VectorAnnotation from API response data.

        Parameters
        ----------
        data : dict[str, Any]
            Raw API response data containing vector annotation fields.

        Returns
        -------
        VectorAnnotation
            VectorAnnotation instance created from the response data.
        """
        return cls(
            vector_uid=data["vectorUid"],
            slide_uid=data.get("slideId", ""),
            actor_type=data.get("actorType", "user"),
            actor_id=data.get("actorId", 1),
            name=data["vectorName"],
            file_uri=data["fileUri"],
            format=data.get("format"),
            labels=data.get("labels"),
            metadata=data.get("metadata"),
            version=data.get("version", 1),
            deleted_at=data.get("deletedAt"),
            deleted_by=data.get("deletedBy"),
            created_at=data.get("createdAt"),
        )


@dataclass
class SlideAnnotationsOverview:
    """Overview of annotations for a slide.

    Attributes
    ----------
    slide_uid : str
        Unique identifier for the slide.
    raster_url : str
        URL for accessing raster annotations.
    vector_url : str
        URL for accessing vector annotations.
    raster_count : int
        Number of raster annotations on this slide.
    vector_count : int
        Number of vector annotations on this slide.
    """

    slide_uid: str
    raster_url: str
    vector_url: str
    raster_count: int
    vector_count: int

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> "SlideAnnotationsOverview":
        """Create SlideAnnotationsOverview from API response data.

        Parameters
        ----------
        data : dict[str, Any]
            Raw API response data containing annotation overview.

        Returns
        -------
        SlideAnnotationsOverview
            SlideAnnotationsOverview instance created from the response data.
        """
        return cls(
            slide_uid=data["slideId"],
            raster_url=data["rasterUrl"],
            vector_url=data["vectorUrl"],
            raster_count=data["rasterCount"],
            vector_count=data["vectorCount"],
        )


@dataclass
class AuthTokens:
    """Authentication tokens with expiration information.

    Attributes
    ----------
    access_token : str
        JWT access token for API authentication.
    refresh_token : str
        JWT refresh token for obtaining new access tokens.
    expires_in : int
        Number of seconds until the access token expires.
    refresh_expires_in : int
        Number of seconds until the refresh token expires.
    token_type : str
        Type of token (typically "Bearer").
    """

    access_token: str
    refresh_token: str
    expires_in: int
    refresh_expires_in: int
    token_type: str = "Bearer"

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> "AuthTokens":
        """Create AuthTokens from API response data.

        Parameters
        ----------
        data : dict[str, Any]
            Raw API response data containing token information.

        Returns
        -------
        AuthTokens
            AuthTokens instance created from the response data.
        """
        return cls(
            access_token=data["access_token"],
            refresh_token=data["refresh_token"],
            expires_in=data["expires_in"],
            refresh_expires_in=data["refresh_expires_in"],
            token_type=data.get("token_type", "Bearer"),
        )


# Response wrapper types for pagination
StudiesResponse = PaginatedResponse[Study]
CasesResponse = PaginatedResponse[Case]
SlidesResponse = PaginatedResponse[Slide]
UsersResponse = PaginatedResponse[User]
TenantsResponse = PaginatedResponse[Tenant]
