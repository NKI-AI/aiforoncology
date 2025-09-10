import numpy as np
import SimpleITK as sitk
from fomo.database_models import Image
from fomo.dataset.vector_dataset.data_producer import DataProducer


class CTDataProducer(DataProducer):
    def __init__(
        self,
        name: str,
        chunk_size_bytes: int,
        max_memory_size_bytes: int,
        database_url: str,
        num_workers: int,
        slices_per_chunk: int = 1,
    ):
        """
        Initialize a CT DataProducer for large-scale datasets.

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
        slices_per_chunk : int
            Slices contained within a single chunk. If 1, DataProducer can be used for 2d training.
            If 0, entire array will be stored within the chunk.
        num_workers : int
            Number of worker processes.
        """
        self._slices_per_chunk = slices_per_chunk
        super().__init__(name, chunk_size_bytes, max_memory_size_bytes, database_url, num_workers)

    # override, method-annotation only available in python 3.12
    def _load_and_preprocess_image(self, db_entry: Image) -> list[np.ndarray]:
        """
        Load and preprocess a medical image from the database entry.

        This method reads an image file specified by the database entry, reorients it to the
        Left-Posterior-Superior (LPS) anatomical orientation, and then processes it into chunks
        of slices for further analysis or training. The method supports both 2D and 3D image
        processing based on the `slices_per_chunk` parameter.

        Parameters
        ----------
        db_entry : Image
            A database entry containing metadata and the filename of the image to be processed.

        Returns
        -------
        list[np.ndarray]
            A list of numpy arrays, each representing a contiguous chunk of image slices. If
            `slices_per_chunk` is set to 0, the entire image is returned as a single chunk.
        """
        image = sitk.ReadImage(db_entry.filename)
        image = self._reorient_image_to_lps(image)
        p = np.random.rand()

        view = sitk.GetArrayFromImage(image)

        if self._slices_per_chunk == 0:
            return [np.ascontiguousarray(view)]

        # start chunking from the other side
        if p >= 0.5:
            view = view[::-1]

        num_full_chunks = len(view) // self._slices_per_chunk
        # if the chunk size is too big for this particular scan, short circuit
        if num_full_chunks == 0:
            return []

        view = view[: num_full_chunks * self._slices_per_chunk]

        chunks = np.array_split(view, num_full_chunks)

        contiguous_chunks = []
        for chunk in chunks:
            # expand color channel
            contiguous_chunks.append(np.ascontiguousarray(chunk))

        return contiguous_chunks
