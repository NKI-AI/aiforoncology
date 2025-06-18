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

import logging

from ._image import SlideImage
from ._region import BoundaryMode, RegionView

from .annotations import SlideAnnotations
from .data.dataset import SlideDataset

pyvips_logger = logging.getLogger("pyvips")
pyvips_logger.setLevel(logging.CRITICAL)

__author__ = """dlup contributors"""
__email__ = "j.teuwen@nki.nl"
__version__ = "0.7.0"

__all__ = ("SlideImage", "SlideDataset", "SlideAnnotations", "RegionView", "BoundaryMode")
