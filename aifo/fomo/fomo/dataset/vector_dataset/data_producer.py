import logging
import random
from abc import ABC, abstractmethod
from multiprocessing import Process
from time import sleep
from typing import Generator

import numpy as np
import SimpleITK as sitk
from aifocore.shm.vector import InUseError, OutOfMemoryError, SharedVector, remove_shared_memory
from fomo.database_models import Image
from fomo.dataset.vector_dataset.data_manager import DataManager
from fomo.utils.types import Orientation

logger = logging.getLogger(__name__)


class DataProducer(ABC):
    """
    Abstract base class that provides the ability of loading and storing/replacing images in a shared vector.
    """

    def __init__(
        self,
        name: str,
        chunk_size_bytes: int,
        max_memory_size_bytes: int,
        database_url: str,
        num_workers: int,
    ):
        """
        Initialize the DataProducer for large-scale datasets.

        Parameters
        ----------
        name : str
            Name of the shared vector in-memory.
        max_memory_size_bytes : int
            Maximum memory of the shared vector in bytes.
        chunk_size_bytes : int
            Size of the chunks of memory in the shared vector in bytes.
        database_url : str
            Database connection URL.
        num_workers : int
            Number of worker processes.
        """
        remove_shared_memory(name)
        self.database_url = database_url
        self.num_workers = num_workers
        self.queue_name = name
        self.chunk_size = chunk_size_bytes
        self.queue_size_bytes = max_memory_size_bytes
        self.workers = []

    def start_loading(self) -> None:
        """Start the data loading process."""
        for worker_id in range(self.num_workers):
            process = Process(target=self._load_data, args=(worker_id,))
            process.start()
            self.workers.append(process)

    def _load_data(self, worker_id: int) -> None:
        """Load and preprocess data, putting results into the shared vector."""
        data_manager = DataManager(self.database_url)
        shared_vector = SharedVector(
            name=self.queue_name,
            chunk_size=self.chunk_size,
            max_memory_size=self.queue_size_bytes,
        )
        logger.info(f"Worker {worker_id} started processing images.")

        image_generator = data_manager.get_random_image_generator(worker_id, self.num_workers)

        print(f"Worker {worker_id}: Starting initial load.")
        self._initial_load(shared_vector, image_generator, worker_id)

        print(f"Worker {worker_id}: Starting replace loop.")
        self._replace_loop(shared_vector, image_generator, worker_id)

    def __getstate__(self):
        """For spawning, we remove the worker variable, which contains weakrefs."""
        state = self.__dict__.copy()
        if "workers" in state:
            del state["workers"]
        return state

    def __setstate__(self, state):
        """For child processes we reinitialize the workers list as empty."""
        self.__dict__.update(state)
        self.workers = []

    def _initial_load(self, shared_vector: SharedVector, image_generator: Generator[Image, None, None], worker_id: int):
        """Initial loading of data into the shared vector."""
        try:
            while True:
                db_entry = next(image_generator)
                image = self._load_and_preprocess_image(db_entry)
                if len(image) > 0:
                    self._append(image, shared_vector, worker_id, db_entry.id)
        except OutOfMemoryError:
            logger.info(f"Worker {worker_id}: Shared vector is full. Stopping initial load.")
        except Exception as e:
            logger.error(f"Worker {worker_id}: Error during initial load: {e}")

    def _replace_loop(self, shared_vector: SharedVector, image_generator: Generator[Image, None, None], worker_id: int):
        """Continuous loop to replace data in the shared vector."""
        logger.info(f"Worker {worker_id} starting replace loop.")
        while True:
            try:
                db_entry = next(image_generator)
                image = self._load_and_preprocess_image(db_entry)
                if len(image) > 0:
                    self._replace(image, shared_vector, worker_id, db_entry.id)
            except InUseError:
                logger.info(f"Worker {worker_id}: Shared vector is in use. Skipping replace.")
                continue
            except Exception as e:
                logger.error(f"Worker {worker_id}: Error in replace loop: {e}")
                continue

    def _append(self, image: list[np.ndarray], shared_vector: SharedVector, worker_id: int, image_id: int):
        for slice_data in image:
            shared_vector.append(slice_data)
            logger.debug(f"Worker {worker_id}: Slice from image ID {image_id} appended.")

    def _replace(
        self, image: list[np.ndarray], shared_vector: SharedVector, worker_id: int, image_id: int, max_retries: int = 3
    ):
        try:
            for slice_data in image:
                random_index = random.randint(0, len(shared_vector) - 1)
                attempts = 0
                while attempts < max_retries:
                    try:
                        shared_vector.replace(random_index, slice_data)
                        logger.debug(
                            f"Worker {worker_id}: Slice from image ID {image_id} replaced at index {random_index}."
                        )
                        break  # Success - exit the retry loop
                    except (InUseError, RuntimeError) as e:
                        attempts += 1
                        if attempts == max_retries:
                            print(f"Worker {worker_id}: Failed to replace slice after {max_retries} attempts: {e}")
                        else:
                            print(f"Worker {worker_id}: Retry attempt {attempts}/{max_retries} for replacing slice")
                        sleep(0.1 * attempts)  # Incremental backoff
                        continue
        except Exception as e:
            print(f"Worker {worker_id}: Error processing image ID {image_id}: {e}")

    def _reorient_image_to_lps(self, image: sitk.Image) -> sitk.Image:
        """Reorient an image to LPS orientation."""
        return sitk.DICOMOrient(image, Orientation.LPS.value)

    @abstractmethod
    def _load_and_preprocess_image(self, db_entry: Image) -> list[np.ndarray]:
        raise NotImplementedError()

    def join(self) -> None:
        """Wait for all worker processes to complete."""
        for worker in self.workers:
            worker.join()
