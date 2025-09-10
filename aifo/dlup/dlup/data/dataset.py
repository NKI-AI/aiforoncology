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
import bisect
import collections.abc
import random
import warnings
from collections import OrderedDict
from dataclasses import dataclass, field
from typing import Any, Callable, Generic, Iterable, Optional, Sequence, TypeVar, Union, overload

import numpy as np
import numpy.typing as npt
import pyvips
from dlup import BoundaryMode, SlideImage
from dlup._geometry import AnnotationRegion
from dlup._types import PathLike
from dlup.annotations import SlideAnnotations
from dlup.background import compute_masked_indices
from dlup.tiling import Grid, GridOrder, TilingMode
from dlup.utils.backends import ImageBackend

# Type aliases
MaskTypes = Union[SlideImage, npt.NDArray[np.int_], SlideAnnotations]
LabelType = Union[str, bool, int, float]
AnnotationData = dict[str, Any]
PointType = tuple[float, float]
BoundingBoxType = tuple[tuple[int, int], tuple[int, int]]

T_co = TypeVar("T_co", covariant=True)
T = TypeVar("T")


@dataclass
class TileSample:
    """
    A sample from a dataset, representing a tile extracted from a slide image.
    """

    image: pyvips.Image
    coordinates: tuple[float, float]
    mpp: float
    path: PathLike
    region_index: int
    labels: Optional[dict[str, Any]] = None
    annotations: Optional[AnnotationRegion] = None


@dataclass
class RegionFromSlideDataset(TileSample):
    """
    A tile sample with additional information about its position in the grid.
    """

    grid_local_coordinates: tuple[int, int] = field(default_factory=lambda: (0, 0))
    grid_index: int = 0


class Dataset(Generic[T_co], collections.abc.Sequence[T_co]):
    """An abstract class representing a :class:`Dataset`.

    All datasets that represent a map from keys to data samples should subclass
    it. All subclasses should overwrite :meth:`__getitem__`, supporting fetching a
    data sample for a given key. Subclasses could also optionally overwrite
    :meth:`__len__`, which is expected to return the size of the dataset by many
    :class:`~torch.utils.data.Sampler` implementations and the default options
    of :class:`~torch.utils.data.DataLoader`.

    Notes
    -----
    Taken and adapted from pytorch 1.8.0 torch.utils.data.Dataset under BSD license.
    :class:`~torch.utils.data.DataLoader` by default constructs a index
    sampler that yields integral indices.  To make it work with a map-style
    dataset with non-integral indices/keys, a custom sampler must be provided.

    """

    def __add__(self, other: "Dataset[T_co]") -> "ConcatDataset[T_co]":
        return ConcatDataset([self, other])

    def __getitem__(self, index: int) -> T_co:  # type: ignore
        raise IndexError


class ConcatDataset(Dataset[T_co]):
    """
    A dataset that concatenates multiple datasets.
    """

    def __init__(self, datasets: Iterable[Dataset[T_co]]) -> None:
        self.datasets = list(datasets)
        assert len(self.datasets) > 0, "datasets should not be an empty iterable"
        for dataset in self.datasets:
            if not hasattr(dataset, "__getitem__"):
                raise ValueError("ConcatDataset requires datasets to be indexable.")
        self.cumulative_sizes = self._compute_cumulative_sizes(self.datasets)

    @staticmethod
    def _compute_cumulative_sizes(datasets: list[Dataset[T_co]]) -> list[int]:
        cumulative_sizes = []
        total = 0
        for dataset in datasets:
            total += len(dataset)
            cumulative_sizes.append(total)
        return cumulative_sizes

    def __len__(self) -> int:
        return self.cumulative_sizes[-1]

    def index_to_dataset(self, index: int) -> tuple[Dataset[T_co], int]:
        if index < 0:
            if -index > len(self):
                raise ValueError("Absolute value of index should not exceed dataset length")
            index = len(self) + index
        dataset_idx = bisect.bisect_right(self.cumulative_sizes, index)
        sample_idx = index if dataset_idx == 0 else index - self.cumulative_sizes[dataset_idx - 1]
        return self.datasets[dataset_idx], sample_idx

    @overload
    def __getitem__(self, index: int) -> T_co: ...

    @overload
    def __getitem__(self, index: slice) -> list[T_co]: ...

    def __getitem__(self, index: Union[int, slice]) -> Union[T_co, list[T_co]]:
        if isinstance(index, slice):
            indices = range(*index.indices(len(self)))
            return [self[i] for i in indices]
        dataset, sample_idx = self.index_to_dataset(index)
        return dataset[sample_idx]


class SlideCache:
    """
    A cache for storing SlideImage instances to optimize memory usage.
    """

    def __init__(self, max_size: int = 32):
        self.cache: OrderedDict[PathLike, SlideImage] = OrderedDict()
        self.max_size = max_size

    def get_slide(
        self,
        path: PathLike,
        backend: ImageBackend,
        apply_color_profile: bool,
        **kwargs: Any,
    ) -> SlideImage:
        if path in self.cache:
            self.cache.move_to_end(path)
            return self.cache[path]
        if len(self.cache) >= self.max_size:
            _, old_slide = self.cache.popitem(last=False)
            old_slide.close()
        slide = SlideImage.from_file_path(path, backend=backend, apply_color_profile=apply_color_profile, **kwargs)
        self.cache[path] = slide
        return slide


slide_cache = SlideCache()


class MaskMixin:
    """
    Mixin class for applying masks to datasets.
    """

    def _apply_mask(
        self,
        regions: Sequence[tuple[float, float, int, int, float]],
        mask: MaskTypes,
        mask_threshold: Optional[float],
        slide_image: SlideImage,
    ) -> npt.NDArray[np.int64]:
        masked_indices = compute_masked_indices(slide_image, mask, regions, mask_threshold)
        return masked_indices


class AnnotationMixin:
    """
    Mixin class for handling annotations in datasets.
    """

    _annotations: Optional[SlideAnnotations] = None

    def _load_annotations(
        self,
        coordinates: tuple[float, float],
        scaling: float,
        region_size: tuple[int, int],
    ) -> Optional[AnnotationRegion]:
        if self._annotations:
            return self._annotations.read_region(coordinates, scaling, region_size)
        return None


class LabelMixin:
    """
    Mixin class for handling labels in datasets.
    """

    _labels: Optional[dict[str, Any]] = None

    def _assign_labels(self) -> Optional[dict[str, Any]]:
        return self._labels


class SlideDataset(Dataset[TileSample], MaskMixin, AnnotationMixin, LabelMixin):
    """
    Dataset class for iterating over tiles extracted from a slide image.
    """

    def __init__(
        self,
        path: PathLike,
        grid: tuple[Grid, tuple[int, int], float],
        crop: bool = False,
        mask: Optional[MaskTypes] = None,
        mask_threshold: Optional[float] = 0.0,
        random_sample_in_grid: bool = False,
        annotations: Optional[SlideAnnotations] = None,
        labels: Optional[dict[str, Any]] = None,
        transform: Optional[Callable[[TileSample], TileSample]] = None,
        backend: ImageBackend = ImageBackend.OPENSLIDE,
        apply_color_profile: bool = False,
        **kwargs: Any,
    ):
        self._path = path
        self._grid = grid
        self._regions = self._compute_regions()
        self._crop = crop
        self._random_sample_in_grid = random_sample_in_grid
        self._annotations = annotations
        self._labels = labels
        self._transform = transform
        self._backend = backend
        self._apply_color_profile = apply_color_profile
        self.kwargs = kwargs

        self._masked_indices: Optional[npt.NDArray[np.int64]] = None
        if mask is not None:
            self._masked_indices = self._apply_mask(self._regions, mask, mask_threshold, self.slide_image)

    @property
    def grids(self) -> list[tuple[Grid, tuple[int, int], float]]:
        """Retrieve the grids used to generate the regions."""
        warnings.warn(
            "The 'grids' attribute is deprecated and will be removed in a future version. "
            "Use the 'grid' attribute instead.",
            DeprecationWarning,
        )
        return [self._grid]

    @property
    def grid(self) -> Grid:
        """Retrieve the grid used to generate the regions."""
        return self._grid[0]

    @property
    def tile_size(self) -> tuple[int, int]:
        """Retrieve the tile size used to generate the regions."""
        return self._grid[1]

    @property
    def mpp(self) -> float:
        """Retrieve the target microns per pixel used to generate the regions."""
        return self._grid[2]

    @property
    def slide_image(self) -> SlideImage:
        """Retrieve the SlideImage instance for this dataset."""
        return slide_cache.get_slide(
            self._path, backend=self._backend, apply_color_profile=self._apply_color_profile, **self.kwargs
        )

    def __len__(self) -> int:
        if self._masked_indices is not None:
            return len(self._masked_indices)
        return len(self._regions)

    @overload
    def __getitem__(self, index: int) -> TileSample: ...

    @overload
    def __getitem__(self, index: slice) -> list[TileSample]: ...

    def __getitem__(self, index: Union[int, slice]) -> Union[TileSample, list[TileSample]]:
        if isinstance(index, slice):
            indices = range(*index.indices(len(self)))
            return [self._get_sample(i) for i in indices]
        return self._get_sample(index)

    def _get_sample(self, index: int) -> TileSample:
        region_index = self._get_region_index(index)
        x, y, w, h, mpp = self._regions[region_index]
        coordinates = (x, y)
        region_size = (w, h)
        scaling = self.slide_image.mpp / mpp

        region_view = self.slide_image.get_scaled_view(scaling)
        region_view.boundary_mode = BoundaryMode.crop if self._crop else BoundaryMode.zero

        if self._random_sample_in_grid:
            x_rand = random.uniform(x, min(x + w, region_view.size[0]))
            y_rand = random.uniform(y, min(y + h, region_view.size[1]))
            coordinates = (int(x_rand), int(y_rand))  # TODO: A float would work here too, but not for mypy

        image = region_view.read_region(coordinates, region_size)
        annotations = self._load_annotations(coordinates, scaling, region_size)
        labels = self._assign_labels()

        sample = TileSample(
            image=image,
            coordinates=coordinates,
            mpp=mpp,
            path=self._path,
            region_index=region_index,
            labels=labels,
            annotations=annotations,
        )

        if self._transform:
            sample = self._transform(sample)

        return sample

    def _get_region_index(self, index: int) -> int:
        if self._masked_indices is not None:
            return int(self._masked_indices[index])
        return index

    def _compute_regions(self) -> list[tuple[int, int, int, int, float]]:
        """Compute regions from grids."""
        regions = []
        for grid, tile_size, grid_mpp in self.grids:
            for coords in grid:
                regions.append(_coords_to_region(tile_size, grid_mpp, key="", coords=coords.tolist()))
        return regions

    @classmethod
    def from_standard_tiling(
        cls,
        path: PathLike,
        mpp: Optional[float],
        tile_size: tuple[int, int],
        tile_overlap: tuple[int, int],
        random_sample_in_grid: bool = False,
        tile_mode: TilingMode = TilingMode.overflow,
        grid_order: GridOrder = GridOrder.C,
        crop: bool = False,
        mask: Optional[MaskTypes] = None,
        mask_threshold: Optional[float] = 0.0,
        annotations: Optional[SlideAnnotations] = None,
        labels: Optional[dict[str, Any]] = None,
        transform: Optional[Callable[[TileSample], TileSample]] = None,
        backend: ImageBackend = ImageBackend.OPENSLIDE,
        apply_color_profile: bool = False,
        limit_bounds: bool = True,
        **kwargs: Any,
    ) -> "SlideDataset":
        with SlideImage.from_file_path(path, backend=backend, **kwargs) as slide_image:
            scaling = slide_image.get_scaling(mpp)
            slide_mpp = slide_image.mpp

            if limit_bounds:
                offset, size = slide_image.get_scaled_slide_bounds(scaling)
            else:
                size = slide_image.get_scaled_size(scaling, limit_bounds=False)
                offset = (0, 0)

        grid = Grid.from_tiling(
            offset=offset,
            size=size,
            tile_size=tile_size,
            tile_overlap=tile_overlap,
            mode=tile_mode,
            order=grid_order,
        )
        grid_mpp = mpp if mpp else slide_mpp

        return cls(
            path=path,
            grid=(grid, tile_size, grid_mpp),
            crop=crop,
            mask=mask,
            mask_threshold=mask_threshold,
            random_sample_in_grid=random_sample_in_grid,
            annotations=annotations,
            labels=labels,
            transform=transform,
            backend=backend,
            apply_color_profile=apply_color_profile,
            **kwargs,
        )


def _coords_to_region(
    tile_size: tuple[int, int], target_mpp: float, key: str, coords: tuple[int, int]
) -> tuple[int, int, int, int, float]:
    """Return the necessary tuple that represents a region."""
    return *coords, *tile_size, target_mpp
