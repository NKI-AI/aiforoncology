# Copyright 2024 AI for Oncology Research Group. All Rights Reserved.
# Copyright 2024 Jonas Teuwen. All Rights Reserved.
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
from typing import cast

import pyvips
from dlup._types import PathLike
from dlup.backends._vips_backend import VipsSlide as VipsSlideCpp


class VipsSlide(VipsSlideCpp):
    def __init__(self, filename: PathLike, rgb: bool = False, apply_color_profile: bool = False) -> None:
        """
        Initialize a VipsSlide object.

        Parameters
        ----------
        filename : PathLike
            Path to the slide file.
        rgb : bool
            Whether to read the image as RGB, by default openslide will return RGBA, while the A channel usually does not have valuable information.
        apply_color_profile : bool
            Whether to apply the color profile to the image, by default it is False
        """
        super().__init__(str(filename), load_with_openslide=False, rgb=rgb, apply_color_profile=apply_color_profile)
        if self._libvips_version != (pyvips.version(0), pyvips.version(1), pyvips.version(2)):
            raise RuntimeError(
                f"Pyvips library is using a different version of libvips than the one installed. "
                f"Installed compiled version: {self._libvips_version}, "
                f"Pyvips version: {pyvips.version(0), pyvips.version(1), pyvips.version(2)}"
            )

    @property
    def spacing(self) -> tuple[float, float]:
        return super().spacing

    @spacing.setter
    def spacing(self, value: tuple[float, float]) -> None:
        self.set_spacing(value)

    @property
    def _libvips_version(self) -> tuple[int, int, int]:
        return cast(tuple[int, int, int], super()._libvips_version)  # type: ignore
