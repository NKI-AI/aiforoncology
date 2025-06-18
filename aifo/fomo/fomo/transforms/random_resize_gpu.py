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
import torch
from typing import Optional, Tuple
from torch import vmap


# We implement our own bilinear since it improves compiled performance by at least 2x
def bilinear_interpolate(input: torch.Tensor, grid: torch.Tensor, align_corners: bool = False) -> torch.Tensor:
    """
    Manually performs bilinear interpolation on an input tensor given a normalized grid.

    Parameters
    ----------
    input : torch.Tensor
        Tensor of shape (N, C, H, W).
    grid : torch.Tensor
        Tensor of shape (N, H_out, W_out, 2) with normalized coordinates in [-1, 1].
    align_corners : bool, optional
        Whether to align corners (False by default, matching torch grid_sample).

    Returns
    -------
    torch.Tensor
        Tensor of shape (N, C, H_out, W_out) resulting from bilinear interpolation.
    """
    N, C, H, W = input.shape
    H_out, W_out = grid.shape[1:3]

    # Convert normalized coordinates to pixel indices
    if align_corners:
        ix = (grid[..., 0] + 1) * (W - 1) / 2
        iy = (grid[..., 1] + 1) * (H - 1) / 2
    else:
        ix = ((grid[..., 0] + 1) * W - 1) / 2
        iy = ((grid[..., 1] + 1) * H - 1) / 2

    # Coordinates
    ix0 = torch.floor(ix).long()
    iy0 = torch.floor(iy).long()
    ix1 = ix0 + 1
    iy1 = iy0 + 1

    # weights
    wx = ix - ix0.float()
    wy = iy - iy0.float()
    w00 = (1 - wx) * (1 - wy)
    w01 = (1 - wx) * wy
    w10 = wx * (1 - wy)
    w11 = wx * wy

    # out of bounds mask
    mask00 = ((ix0 >= 0) & (ix0 < W) & (iy0 >= 0) & (iy0 < H)).float()
    mask01 = ((ix0 >= 0) & (ix0 < W) & (iy1 >= 0) & (iy1 < H)).float()
    mask10 = ((ix1 >= 0) & (ix1 < W) & (iy0 >= 0) & (iy0 < H)).float()
    mask11 = ((ix1 >= 0) & (ix1 < W) & (iy1 >= 0) & (iy1 < H)).float()

    # Clamp indices so that they lie within [0, W-1] and [0, H-1].
    ix0_clamped = torch.clamp(ix0, 0, W - 1)
    iy0_clamped = torch.clamp(iy0, 0, H - 1)
    ix1_clamped = torch.clamp(ix1, 0, W - 1)
    iy1_clamped = torch.clamp(iy1, 0, H - 1)

    # Create batch indices for advanced indexing. This is needed because when gathering
    # pixel values, we need to specify which batch item to sample from. The resulting
    # tensor has shape (N, H_out, W_out) where each row i contains all i's, like:
    # [[0,0,0...], [1,1,1...], [2,2,2...], ...] ensuring we sample from the correct batch item.
    batch_indices = torch.arange(N, device=input.device).view(N, 1, 1).expand(N, H_out, W_out)

    # Channels to last dim, purely for more intuitive indexing
    input_perm = input.permute(0, 2, 3, 1)  # shape (N, H, W, C)

    Ia = input_perm[batch_indices, iy0_clamped, ix0_clamped, :]  # top-left
    Ib = input_perm[batch_indices, iy1_clamped, ix0_clamped, :]  # bottom-left
    Ic = input_perm[batch_indices, iy0_clamped, ix1_clamped, :]  # top-right
    Id = input_perm[batch_indices, iy1_clamped, ix1_clamped, :]  # bottom-right

    Ia = Ia * mask00.unsqueeze(-1)
    Ib = Ib * mask01.unsqueeze(-1)
    Ic = Ic * mask10.unsqueeze(-1)
    Id = Id * mask11.unsqueeze(-1)

    w00 = w00.unsqueeze(-1)
    w01 = w01.unsqueeze(-1)
    w10 = w10.unsqueeze(-1)
    w11 = w11.unsqueeze(-1)

    # Compute the weighted sum
    output = Ia * w00 + Ib * w01 + Ic * w10 + Id * w11  # shape (N, H_out, W_out, C)

    # Permute back to (N, C, H_out, W_out)
    output = output.permute(0, 3, 1, 2)
    return output


def select_z_slices(
    input_tensor: torch.Tensor, out_z_dim: int, start_z: Optional[int] = None
) -> Tuple[torch.Tensor, list[int]]:
    """
    Selects different contiguous blocks of z-slices for each item in the batch.

    Parameters
    ----------
    input_tensor : torch.Tensor
        Tensor of shape (N, Z, H, W)
    out_z_dim : int
        Number of contiguous z-slices to select
    start_z : int, optional
        Optional starting z-index. If None, random starting indices
        are chosen for each batch item.

    Returns
    -------
    torch.Tensor
        Selected tensor with shape (N, out_z_dim, H, W)
    list of int
        List of starting z-indices used for selection
    """
    N, Z, H, W = input_tensor.shape
    out_z_dim = min(out_z_dim, Z)

    if start_z is None and Z > out_z_dim:
        max_start = Z - out_z_dim
        start_z_tensor = torch.randint(
            0, max_start + 1, (N,), device=input_tensor.device
        )  # different random start for each batch item

        batch_idx = torch.arange(N, device=input_tensor.device)
        z_idx = torch.arange(out_z_dim, device=input_tensor.device)
        z_indices = (start_z_tensor[:, None] + z_idx[None, :]).view(N, out_z_dim)
        selected_tensor = input_tensor[batch_idx[:, None], z_indices, :, :]

    else:
        start_z = start_z or 0
        if isinstance(start_z, (int, float)):
            start_z = [start_z] * N
        selected_tensor = input_tensor[:, start_z[0] : start_z[0] + out_z_dim, :, :]

    return selected_tensor, start_z


def random_resize_gpu_single(
    padded: torch.Tensor,  # shape: (N, Z, H, W) where Z is the z-dimension
    crop_coords: torch.Tensor,
    output_size: Tuple[int, int],
    out_z_dim: Optional[int] = None,
    start_z: Optional[int] = None,
) -> torch.Tensor:
    """
    Crop and resize each image in the padded tensor according to the crop coordinates.

    Parameters
    ----------
    padded : torch.Tensor
        A tensor of shape (N, Z, H, W) on GPU, where Z is the z-dimension
    crop_coords : torch.Tensor
        A tensor of shape (N, 7), where each row is
        (batch_idx, x1, y1, x2, y2, H_i, W_i)
    output_size : tuple of int
        A tuple (H_out, W_out) specifying the desired output size
    out_z_dim : int, optional
        Optional number of z-slices to select
    start_z : int, optional
        Optional starting z-index for selection

    Returns
    -------
    torch.Tensor
        Cropped and resized tensor
    """
    N, Z, H, W = padded.shape
    H_out, W_out = output_size

    if out_z_dim is not None and Z != out_z_dim:
        padded, start_z = select_z_slices(padded, out_z_dim, start_z)

    # Create grid
    xs = torch.linspace(-1, 1, W_out, device=padded.device)
    ys = torch.linspace(-1, 1, H_out, device=padded.device)
    grid_y, grid_x = torch.meshgrid(ys, xs, indexing="ij")
    base_grid = torch.stack([grid_x, grid_y], dim=-1)  # (H_out, W_out, 2)
    base_grid = base_grid.unsqueeze(0).expand(N, -1, -1, -1)  # (N, H_out, W_out, 2)

    x1 = crop_coords[:, 1].view(N, 1, 1)
    y1 = crop_coords[:, 2].view(N, 1, 1)
    x2 = crop_coords[:, 3].view(N, 1, 1)
    y2 = crop_coords[:, 4].view(N, 1, 1)

    w_scale = (x2 - x1) / (2 * W)
    h_scale = (y2 - y1) / (2 * H)
    w_offset = (x1 + x2) / (2 * W)
    h_offset = (y1 + y2) / (2 * H)

    # transform grid, note this maps it within [0,1]
    grid_x = base_grid[..., 0] * w_scale + w_offset
    grid_y = base_grid[..., 1] * h_scale + h_offset

    # Back to [-1,1]
    grid_x = grid_x * 2 - 1
    grid_y = grid_y * 2 - 1
    grid = torch.stack([grid_x, grid_y], dim=-1)

    return bilinear_interpolate(padded, grid, align_corners=False)


def random_resize_gpu(
    padded: torch.Tensor,  # shape: (N, Z, H, W)
    crop_coords: torch.Tensor,  # shape: (k*N, 7)
    output_size: Tuple[int, int],
    out_z_dim: Optional[int] = None,
    start_z: Optional[int] = None,
) -> torch.Tensor:
    """
    Vectorized version of `random_resize_gpu_single` that applies the operation
    for each group of crop coordinates.

    Parameters
    ----------
    padded : torch.Tensor
        Tensor of shape (N, Z, H, W) on GPU.
    crop_coords : torch.Tensor
        Tensor of shape (k*N, 7) containing crop coordinates.
    output_size : tuple of int
        Desired output spatial size as (H_out, W_out).
    out_z_dim : int, optional
        Optional number of z-slices to select.
    start_z : int, optional
        Optional starting z-index for z-slice selection.

    Returns
    -------
    torch.Tensor
        A tensor of shape (k*N, Z, H_out, W_out) resulting from the batched
        execution of random_resize_gpu.

    Notes
    -----
    Assumes that the first axis of `crop_coords` can be divided into groups of N
    (i.e. shape becomes (k, N, 7)). This is true for dinov2 crops, where we have
    e.g. k=8 local crops.

    Compile this function.
    """
    # Determine group count k based on padded batch size N.
    kN, _ = crop_coords.shape
    N = padded.shape[0]
    if kN % N != 0:
        raise ValueError("crop_coords first dimension must be a multiple of padded batch size N.")
    k: int = kN // N

    # Reshape crop_coords to (k, N, 7) so that each group corresponds to one
    # batch of crops for the padded tensor.
    crop_coords_grouped: torch.Tensor = crop_coords.reshape(k, N, 7)

    # Create a vectorized version of random_resize_gpu.
    # in_dims=(None, 0, None, None, None) means:
    #   - padded is shared across all groups (not mapped),
    #   - crop_coords_grouped is mapped along its first dimension,
    #   - output_size, out_z_dim, and start_z are static.
    # randomness="different" means that each group gets its own random number stream (for z dim in our case)
    v_random_resize = vmap(random_resize_gpu_single, in_dims=(None, 0, None, None, None), randomness="different")

    # Apply vmap: output has shape (k, N, C, H_out, W_out)
    output: torch.Tensor = v_random_resize(padded, crop_coords_grouped, output_size, out_z_dim, start_z)
    output = output.reshape(-1, *output.shape[2:]).contiguous()
    return output
