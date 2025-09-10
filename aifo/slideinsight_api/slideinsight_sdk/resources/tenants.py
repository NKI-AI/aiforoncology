"""
Tenants resource manager for SlideInsight SDK.

Provides operations for managing tenants including creation, updates,
and domain management.
"""

from __future__ import annotations

from typing import Optional

from .base import BaseResourceManager
from ..models import Tenant, TenantDomain


class TenantsManager(BaseResourceManager[Tenant]):
    """Manager for tenant resources."""

    def __init__(self, http_client):
        super().__init__(
            http_client=http_client, base_endpoint="/api/v1/tenants", model_class=Tenant, list_response_key="tenants"
        )

    async def create(
        self,
        name: str,
        description: str = "",
        status: str = "active",
    ) -> Tenant:
        """Create a new tenant.

        Args:
            name: Tenant name
            description: Tenant description
            status: Tenant status

        Returns:
            Created tenant
        """
        data = {
            "name": name,
            "description": description,
            "status": status,
        }

        response_data = await self.http_client.request("POST", self.base_endpoint, json=data)
        return Tenant.from_dict(response_data)

    async def update(
        self,
        tenant_uid: str,
        name: Optional[str] = None,
        description: Optional[str] = None,
        status: Optional[str] = None,
    ) -> Tenant:
        """Update an existing tenant.

        Args:
            tenant_uid: Tenant UID
            name: New tenant name
            description: New tenant description
            status: New tenant status

        Returns:
            Updated tenant
        """
        data = {}

        if name is not None:
            data["name"] = name
        if description is not None:
            data["description"] = description
        if status is not None:
            data["status"] = status

        endpoint = self._build_endpoint(tenant_uid)
        response_data = await self.http_client.request("PUT", endpoint, json=data)
        return Tenant.from_dict(response_data)

    async def get_domains(self, tenant_uid: str) -> list[TenantDomain]:
        """Get domains for a tenant.

        Args:
            tenant_uid: Tenant UID

        Returns:
            List of tenant domains
        """
        endpoint = self._build_endpoint(f"{tenant_uid}/domains")
        response_data = await self.http_client.request("GET", endpoint)

        domains_data = response_data.get("domains", [])
        return [TenantDomain.from_dict(domain) for domain in domains_data]

    async def add_domain(
        self,
        tenant_uid: str,
        domain: str,
        is_primary: bool = False,
    ) -> None:
        """Add a domain to a tenant.

        Args:
            tenant_uid: Tenant UID
            domain: Domain name
            is_primary: Whether this is the primary domain
        """
        endpoint = self._build_endpoint(f"{tenant_uid}/domains")
        data = {
            "domain": domain,
            "isPrimary": is_primary,
        }
        await self.http_client.request("POST", endpoint, json=data)

    async def update_domain(
        self,
        tenant_uid: str,
        domain_id: int,
        is_verified: Optional[bool] = None,
        is_primary: Optional[bool] = None,
    ) -> None:
        """Update a tenant domain.

        Args:
            tenant_uid: Tenant UID
            domain_id: Domain ID
            is_verified: New verification status
            is_primary: New primary status
        """
        endpoint = self._build_endpoint(f"{tenant_uid}/domains/{domain_id}")
        data = {}

        if is_verified is not None:
            data["isVerified"] = is_verified
        if is_primary is not None:
            data["isPrimary"] = is_primary

        await self.http_client.request("PUT", endpoint, json=data)

    async def remove_domain(self, tenant_uid: str, domain_id: int) -> None:
        """Remove a domain from a tenant.

        Args:
            tenant_uid: Tenant UID
            domain_id: Domain ID to remove
        """
        endpoint = self._build_endpoint(f"{tenant_uid}/domains/{domain_id}")
        await self.http_client.request("DELETE", endpoint)
