# Copyright 2025 Joren Brunekreef. All Rights Reserved.
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
from aifocore.pathology import inference_engine as ie
from pathlib import Path
from tqdm import tqdm
import argparse


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--input_dir", type=Path, required=True)
    parser.add_argument("--output_dir", type=Path, required=True)
    parser.add_argument("--model_path", type=Path, required=True)
    parser.add_argument("--device", type=str, required=True)
    parser.add_argument("--batch_size", type=int, required=False, default=4)
    parser.add_argument("--create_thumbnail", action="store_true", required=False, default=False)
    args = parser.parse_args()

    input_dir = args.input_dir
    output_dir = args.output_dir
    model_path = args.model_path
    device = args.device
    batch_size = args.batch_size
    create_thumbnail = args.create_thumbnail

    config = ie.InferenceConfig(
        output_dir=output_dir,
        model_path=model_path,
        device=device,
        batch_size=batch_size,
        create_thumbnail=create_thumbnail,
    )

    engine = ie.InferenceEngine(config)

    image_paths = list(input_dir.glob("*.mrxs")) + list(input_dir.glob("*.svs"))

    # Single progress bar over images; description text shows per-image batch progress
    pbar = tqdm(total=len(image_paths), desc="processing images")
    for i, image_path in enumerate(image_paths):

        def progress_cb(current_batch: int, total_batches: int, tile_index: int):
            # Update description to reflect per-image batch progress (formatted in Python)
            pbar.set_description(
                f"image {i + 1}/{len(image_paths)}: {image_path.name} batch {current_batch}/{total_batches}"
            )
            pbar.set_postfix_str(f"tile {tile_index}")

        engine.set_progress_callback(progress_cb)
        try:
            engine.process_image(image_path=image_path)
        finally:
            engine.set_progress_callback(None)
            pbar.update(1)
    pbar.close()


if __name__ == "__main__":
    main()
