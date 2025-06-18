import logging
import random
from time import time
from typing import Callable, Optional

import torch
import torch.distributed
from aifocore.shm.vector import DataModifiedError, SharedVector
from torch.utils.data import IterableDataset

logger = logging.getLogger(__name__)


class SharedVectorDataset(IterableDataset):
    def __init__(
        self,
        name: str,
        max_memory_size_bytes: int,
        chunk_size_bytes: int,
        is_3d: bool = False,
        transforms: Optional[Callable] = None,
    ):
        """
        Initialize the SharedVectorDataset.

        Parameters
        ----------
        name : str
            Name of the shared vector in-memory.
        max_memory_size_bytes : int
            Maximum memory of the shared vector in bytes.
        chunk_size_bytes : int
            Size of the chunks of memory in the shared vector in bytes.
        is_3d : bool
            Whether the data in the SharedVector is three dimensional.
        transforms : callable, optional
            Transforms to apply to each slice.
        """
        self.name = name
        self.max_memory_size_bytes = max_memory_size_bytes
        self.chunk_size_bytes = chunk_size_bytes
        self.transforms = transforms
        self._is_3d = is_3d

    def _get_dummy_target(self):
        dummy_target = torch.zeros(1)
        return dummy_target

    def __iter__(self):
        seed = int(time())
        shared_vector = SharedVector(
            self.name,
            max_memory_size=self.max_memory_size_bytes,
            chunk_size=self.chunk_size_bytes,
        )
        print("Consumer: Created SharedVector.")
        while True:
            indices = list(range(len(shared_vector)))
            random.seed(seed)
            random.shuffle(indices)
            for random_index in indices:
                try:
                    item = shared_vector.get(random_index)
                    slice_data = torch.from_numpy(item)
                    if self.transforms:
                        slice_data = self.transforms(slice_data)
                    if self._is_3d:
                        yield slice_data.unsqueeze(0)
                    else:
                        target = self._get_dummy_target()
                        yield slice_data, target

                except DataModifiedError:
                    logger.error("Data modified, skipping this get")
                    continue
                except Exception as e:
                    logger.error(f"Error processing slice: {e}")
                    continue
            seed = int(time())  # new seed for 'epoch'
