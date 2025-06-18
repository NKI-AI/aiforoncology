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
import argparse
from enum import Enum

import torch.multiprocessing as mp
from aifocore.shm.vector import remove_shared_memory
from fomo.ct.dataset.ct_data_producer import CTDataProducer
from fomo.mri.dataset.mri_data_producer import MRIDataProducer

mp.set_start_method("spawn", force=True)


class DataProducerType(Enum):
    CT = "ct"
    MRI = "mri"

    def __str__(self):
        return self.value


def parse_arguments():
    parser = argparse.ArgumentParser(description="Start DataProducer with specified parameters.")

    parser.add_argument("--name", type=str, required=True, help="Name of the vector object")
    parser.add_argument("--chunk-size-mb", type=int, required=True, help="Maximum size of the vector")
    parser.add_argument(
        "--slices-per-chunk",
        type=int,
        default=1,
        help="Number of slices to take in a single chunk (Currently only applicable to CT).",
    )
    parser.add_argument(
        "--max-memory-size-gb",
        type=float,
        required=True,
        help="Size of the vector in GB (converted to bytes)",
    )
    parser.add_argument("--num-workers", type=int, required=True, help="Number of workers")
    parser.add_argument(
        "--database-url",
        type=str,
        required=True,
        help="URL of the database to load data from",
    )
    parser.add_argument(
        "--producer-type",
        type=DataProducerType,
        required=True,
        choices=list(DataProducerType),
        help="Type/modality of the data producer loading data into the vector",
    )

    args = parser.parse_args()
    return args


if __name__ == "__main__":
    args = parse_arguments()
    remove_shared_memory(args.name)

    # Convert queue size from GB to bytes
    memory_size_bytes = int(args.max_memory_size_gb * (1024**3))
    chunk_size_bytes = int(args.chunk_size_mb * (1024**2))

    # Initialize and start the DataProducer
    match args.producer_type:
        case DataProducerType.MRI:
            data_producer = MRIDataProducer(
                name=args.name,
                chunk_size_bytes=chunk_size_bytes,
                max_memory_size_bytes=memory_size_bytes,
                num_workers=args.num_workers,
                database_url=args.database_url,
            )
        case DataProducerType.CT:
            data_producer = CTDataProducer(
                name=args.name,
                chunk_size_bytes=chunk_size_bytes,
                max_memory_size_bytes=memory_size_bytes,
                num_workers=args.num_workers,
                database_url=args.database_url,
                slices_per_chunk=args.slices_per_chunk,
            )
        case _:
            raise NotImplementedError(f"Data producer type of '{args.producer_type}' has not been implemented.")

    data_producer.start_loading()
    data_producer.join()
