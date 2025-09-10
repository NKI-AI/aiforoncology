# Copyright 2025 AI for Oncology Research Group. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
from enum import Enum


class Orientation(Enum):
    """Orientation enum with DICOM orientation codes.

    LPS = Left-Posterior-Superior (standard radiological convention)
    PSL = Posterior-Superior-Left (alternative orientation)
    PLI = Posterior-Left-Inferior (sagittal orientation in LPS-equivalent space)
    """

    LPS = "LPS"  # Standard for axial views
    PSL = "PSL"  # Alternative orientation
    PLI = "PLI"  # For sagittal views
    RAS = "RAS"  # Right-Anterior-Superior

    @classmethod
    def from_value(cls, value: str) -> "Orientation":
        """Get the Orientation enum value from a string value.

        Args:
            value: String value to convert to Orientation enum

        Returns:
            The corresponding Orientation enum value

        Raises:
            ValueError: If the value is not a valid orientation
        """
        try:
            return cls(value)
        except ValueError:
            raise ValueError(f"Invalid orientation value: {value}. Must be one of {[o.value for o in cls]}")

    @classmethod
    def get_orientation_for_plane(cls, plane: str, base_orientation: str = "LPS"):
        """Get the appropriate orientation code for a given acquisition plane.

        Args:
            plane: Acquisition plane ("axial", "sagittal", or "coronal")
            base_orientation: Base orientation to use (default: "LPS")

        Returns:
            Orientation enum value appropriate for the plane
        """
        base = cls.from_value(base_orientation)

        if plane.lower() == "axial":
            return base
        elif plane.lower() == "sagittal":
            return cls.PLI  # Sagittal orientation in LPS-equivalent space
        elif plane.lower() == "coronal":
            return base
        else:
            return base
