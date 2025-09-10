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
from dlup._slide_dataset import SlideDataset
from dlup._slide_image import SlideImage
from dlup.backends._vips_backend import VipsSlide
from aifocore.tiling import Grid

backend = VipsSlide("tests/files/checkerboard.svs")
slide = SlideImage(backend)

grid = Grid.from_tiling((0, 0), slide.dimensions, (4000, 4000), (0, 0), mode="overflow", order="C")

dataset = SlideDataset(slide, grid, slide.mpp, (4000, 4000), False)

print("Length of dataset:", len(dataset))

print("Coordinates of the first sample:")
for idx, sample in enumerate(dataset):
    coordinates = sample.coordinates
    filename = f"sample_{coordinates[0]}_{coordinates[1]}.png"
    print(f"Coordinates: {coordinates} written to {filename}")
    sample.tile.write_to_file(filename)
