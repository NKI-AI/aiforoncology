"""
Users resource manager for SlideInsight SDK.

Provides operations for managing users including creation, updates,
and user lookup operations.
"""

from __future__ import annotations

from typing import Optional

from .base import BaseResourceManager
from ..models import User


class UsersManager(BaseResourceManager[User]):
    """Manager for user resources."""

    def __init__(self, http_client):
        super().__init__(
            http_client=http_client, base_endpoint="/api/v1/users", model_class=User, list_response_key="users"
        )

    async def create(
        self,
        email: str,
        password: str,
        first_name: Optional[str] = None,
        last_name: Optional[str] = None,
        is_active: bool = True,
        must_reset_password: bool = False,
    ) -> User:
        """Create a new user.

        Args:
            email: Email address
            password: Password
            first_name: First name
            last_name: Last name
            is_active: Whether user is active
            must_reset_password: Whether user must reset password on first login

        Returns:
            Created user
        """
        data = {
            "email": email,
            "password": password,
            "isActive": is_active,
            "mustResetPassword": must_reset_password,
        }

        if first_name is not None:
            data["firstName"] = first_name
        if last_name is not None:
            data["lastName"] = last_name

        response_data = await self.http_client.request("POST", self.base_endpoint, json=data)
        return User.from_dict(response_data)

    async def update(
        self,
        user_uid: str,
        email: Optional[str] = None,
        first_name: Optional[str] = None,
        last_name: Optional[str] = None,
        is_active: Optional[bool] = None,
        must_reset_password: Optional[bool] = None,
    ) -> User:
        """Update an existing user.

        Args:
            user_uid: User UID
            email: New email
            first_name: New first name
            last_name: New last name
            is_active: New active status
            must_reset_password: New password reset requirement

        Returns:
            Updated user
        """
        data = {}

        if email is not None:
            data["email"] = email
        if first_name is not None:
            data["firstName"] = first_name
        if last_name is not None:
            data["lastName"] = last_name
        if is_active is not None:
            data["isActive"] = is_active
        if must_reset_password is not None:
            data["mustResetPassword"] = must_reset_password

        endpoint = self._build_endpoint(user_uid)
        response_data = await self.http_client.request("PUT", endpoint, json=data)
        return User.from_dict(response_data)

    async def get_current(self) -> User:
        """Get the current authenticated user.

        Returns:
            Current user
        """
        endpoint = "/api/v1/auth/me"
        response_data = await self.http_client.request("GET", endpoint)
        return User.from_dict(response_data)

    async def change_password(
        self,
        current_password: str,
        new_password: str,
    ) -> None:
        """Change the current user's password.

        Args:
            current_password: Current password
            new_password: New password
        """
        endpoint = "/api/v1/auth/change-password"
        data = {
            "currentPassword": current_password,
            "newPassword": new_password,
        }
        await self.http_client.request("POST", endpoint, json=data)
