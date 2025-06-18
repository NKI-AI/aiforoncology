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
import argparse
import random
from functools import partial
from pathlib import Path

import numpy as np
from dlup import SlideImage
from dlup.backends import PyVipsSlide, VipsSlide
from python.runfiles import runfiles
from tqdm import tqdm


def generate_random_location_and_scaling(bounds, size):
    """Generate random locations within the slide bounds and a random scaling factor."""
    scaling = random.uniform(0.2, 1.0)
    max_x = max(0, int((bounds[1][0] - size[0]) * scaling))
    max_y = max(0, int((bounds[1][1] - size[1]) * scaling))
    x = random.randint(0, max_x)
    y = random.randint(0, max_y)
    return (x, y), scaling


def main(args):
    # Resolve slide path using Bazel runfiles
    r = runfiles.Create()
    slide_path = r.Rlocation(args.slide_path)
    if not slide_path or not Path(slide_path).exists():
        raise FileNotFoundError(f"Slide file not found at {args.slide_path}")

    # Initialize SlideImage objects with different backends
    slides = {
        "VipsSlide": SlideImage.from_file_path(slide_path, backend=partial(VipsSlide, rgb=True)),
        "PyVipsSlide": SlideImage.from_file_path(slide_path, backend=partial(PyVipsSlide, rgb=True)),
    }

    # Number of random reads and size of regions
    size = (args.width, args.height)
    slide_bounds = slides["VipsSlide"].slide_bounds

    # Perform random reads and compare results
    print("Reading random regions and comparing results...")
    for _ in tqdm(range(args.num_reads)):
        location, scaling = generate_random_location_and_scaling(slide_bounds, size)
        print(location, scaling)
        scaled_size = tuple(int(dim * scaling) for dim in size)
        region_pyvips = slides["PyVipsSlide"].read_region(location, scaling, scaled_size)
        region_vips = slides["VipsSlide"].read_region(location, scaling, scaled_size)
        assert np.array_equal(region_vips.numpy(), region_pyvips.numpy())

    print("Results are consistent between backends.")

    # Close slides
    slides["VipsSlide"].close()
    slides["PyVipsSlide"].close()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Compare SlideImage backends.")
    parser.add_argument("--slide_path", type=str, required=True, help="Path to an image.")
    parser.add_argument("--num_reads", type=int, default=10, help="Number of random regions to read and compare.")
    parser.add_argument("--width", type=int, default=5000, help="Width of the region to read.")
    parser.add_argument("--height", type=int, default=5000, help="Height of the region to read.")

    args = parser.parse_args()
    main(args)
