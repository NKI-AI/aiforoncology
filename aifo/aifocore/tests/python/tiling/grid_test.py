import numpy as np
import pytest
from aifocore.tiling import Grid, GridOrder, TilingMode, tiles_grid_coordinates


class TestTiling:
    @pytest.mark.parametrize("mode", [TilingMode.skip])
    def test_all_zero(self, mode):
        """If all the arguments are zero, an exception should be raised."""
        with pytest.raises(ValueError):
            (basis,) = tiles_grid_coordinates(0, 0, 0, mode=mode)

    @pytest.mark.parametrize("mode", [TilingMode.skip, TilingMode.overflow])
    @pytest.mark.parametrize("tile_overlap", [0, 1, 2])
    def test_tile_bigger_than_size(self, mode, tile_overlap):
        """Check different modes if tile_size is bigger than the size."""
        size = 2
        tile_size = 10
        (basis,) = tiles_grid_coordinates(size, tile_size, tile_overlap=tile_overlap, mode=mode)

        expected_lengths = {TilingMode.skip: 0, TilingMode.overflow: 1}

        for elem in basis:
            assert elem >= 0

        assert len(basis) == expected_lengths[mode]

    @pytest.mark.parametrize(
        "size, tile_size, tile_overlap",
        [(10, 3, 0), (3, 1, 2), (17, 3.2, 2), (53.2, 12.2, 15), (1, 2, 3)],
    )
    @pytest.mark.parametrize("mode", [TilingMode.skip, TilingMode.overflow])
    def test_spanned_basis(self, size, tile_size, tile_overlap, mode):
        """Check the spanned basis behaves as configured for tiles."""
        (basis,) = tiles_grid_coordinates(size, tile_size, tile_overlap=tile_overlap, mode=mode)

        assert np.all(np.diff(basis) >= 0)

        if len(basis) == 0:
            return

        # First coordinate is always zero.
        assert basis[0] == 0

        stride = np.diff(basis)
        tiled_size = basis[-1] + tile_size

        # Grid is uniform
        if len(stride):
            assert np.isclose(stride, stride[0]).all()

        if np.isclose(tiled_size, size):
            return

        if mode == TilingMode.skip:
            assert tiled_size < size

        if mode == TilingMode.overflow:
            assert tiled_size > size

    def test_spanned_basis_multiple_dims(self):
        """Check that multiple dims is the same as a single dim."""
        (basis,) = tiles_grid_coordinates(10, 3, 1.2)
        dbasis, _ = tiles_grid_coordinates((10, 5), (3, 2), (1.2, 1))
        assert all([a == b for a, b in zip(basis, dbasis)])

    @pytest.mark.parametrize("order", ["F", "C", GridOrder.C])
    def test_grid(self, order):
        """Test Grid basic api."""
        grid = Grid([(0, 1), (2, 3, 4)], order=order)

        assert grid.size == (2, 3)
        assert len(grid) == 6

        # First row, first column
        assert grid[0][0] == 0 and grid[0][1] == 2

        if order in ["F", GridOrder.F]:
            # First row, second column
            assert (np.asarray(grid[1]) == (0, 3)).all()
            assert [_ for _ in grid[0:2]] == [(0, 2), (0, 3)]
        else:
            # In C order we need to look at the third element
            assert (np.asarray(grid[2]) == (0, 3)).all()
            assert [_ for _ in grid[0:3]] == [(0, 2), (1, 2), (0, 3)]

    # This does not work well yet, we would need to be able to check == "F" too
    # assert grid.order == order

    @pytest.mark.parametrize("value_a", [(0.0, 0), (1, 1.0), (3.0, 3.0), (0, 0)])
    @pytest.mark.parametrize("value_b", [(0.0, 0, 3), (1, 1.0, 4), (3.0, 3.0), (0, 0)])
    def test_data_type(self, value_a, value_b):
        """Test the data type of the grid."""
        grid = Grid([value_a, value_b], order=GridOrder.C)

        if np.asarray(value_a).dtype == np.float64 or np.asarray(value_b).dtype == np.float64:
            assert isinstance(grid[0][0], float)
            assert isinstance(grid[0][1], float)

        else:
            assert isinstance(grid[0][0], int)
            assert isinstance(grid[0][1], int)

    @pytest.mark.parametrize("tile_overlap", [(0, 0), (1, 1)])
    def test_overlap(self, tile_overlap):
        grid = Grid.from_tiling(offset=(0, 0), size=(100, 200), tile_size=(10, 20), tile_overlap=tile_overlap)
        assert grid[0] == (0, 0)
        assert grid[1] == (10 - tile_overlap[0], 0)

    @pytest.mark.parametrize("tile_size", [(1, 2), (-1, 0)])
    def test_exceptions(self, tile_size: tuple[int, int]):
        """Test exceptions."""

        if tile_size[0] < 0 or tile_size[1] < 0:
            with pytest.raises(ValueError, match="tile_size must be greater than zero."):
                Grid.from_tiling(offset=(0, 0), size=(2, 1), tile_size=tile_size, tile_overlap=(3, 2))
        else:
            with pytest.raises(ValueError, match="size, tile_size, and tile_overlap must have the same dimensions."):
                Grid.from_tiling(offset=(0, 0), size=(2,), tile_size=tile_size, tile_overlap=(3, 2))
