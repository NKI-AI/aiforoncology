"""
SlideInsight SDK resource managers.

This package contains all the resource managers for different API entities.
"""

from .base import BaseResourceManager, PaginationHelper
from .studies import StudiesManager
from .cases import CasesManager
from .slides import SlidesManager
from .users import UsersManager
from .tenants import TenantsManager

__all__ = [
    "BaseResourceManager",
    "PaginationHelper",
    "StudiesManager",
    "CasesManager",
    "SlidesManager",
    "UsersManager",
    "TenantsManager",
]
