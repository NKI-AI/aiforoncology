import math
import torch
from typing import List, Tuple
from logging import getLogger

logger = getLogger(__name__)


def pad_tensors_and_compute_crop_coords(
    tensors: List[torch.Tensor],
    global_scale_range: Tuple[float, float] = (0.2, 0.5),
    local_scale_range: Tuple[float, float] = (0.05, 0.2),
    ratio_range: Tuple[float, float] = (3 / 4, 4 / 3),
    n_global_crops: int = 2,
    n_local_crops: int = 6,
    pad_size: Tuple[int, int] = (512, 512),
    custom_stream: torch.cuda.Stream | None = None,
    device: torch.device = torch.device("cuda"),
) -> Tuple[torch.Tensor, torch.Tensor]:
    """
    Pads tensors and computes multiple crop coordinates for both global and local views.

    Parameters
    ----------
    tensors : list of torch.Tensor
        List of input tensors
    global_scale_range : tuple of float
        Scale range for global crops
    local_scale_range : tuple of float
        Scale range for local crops
    ratio_range : tuple of float
        Aspect ratio range for all crops
    n_global_crops : int
        Number of global crops to generate
    n_local_crops : int
        Number of local crops to generate
    pad_size : tuple of int
        Size to pad tensors to
    custom_stream : torch.cuda.Stream or None
        CUDA stream to use for non-blocking operations
    device : torch.device
        Device to run the operations on

    Returns
    -------
    tuple
        Tuple of (padded_tensor, crop_coordinates)
        crop_coordinates shape: (n_images * (n_global_crops + n_local_crops), 7)

    Notes
    -----
    All ops are done on the GPU, including the random number generation.
    """
    n = len(tensors)
    total_crops_per_image = n_global_crops + n_local_crops

    # Initialize CUDA tensors
    z_dim = tensors[0].shape[1]
    padded = torch.zeros((n, z_dim, *pad_size), dtype=tensors[0].dtype, device=device)
    heights = torch.tensor([t.shape[2] for t in tensors], device=device)
    widths = torch.tensor([t.shape[3] for t in tensors], device=device)

    # Copy into padded tensor
    # Non blocking (only works with pinned memory) ensures that the data is copied to the GPU asynchronously
    if custom_stream is not None:
        with torch.cuda.stream(custom_stream):
            for i, t in enumerate(tensors):
                H, W = t.shape[2:]
                t_cpu = t.squeeze(0)
                padded[i, :, :H, :W].copy_(t_cpu.to(device), non_blocking=True)
    else:
        for i, t in enumerate(tensors):
            H, W = t.shape[2:]
            t_cpu = t.squeeze(0)
            padded[i, :, :H, :W].copy_(t_cpu.to(device), non_blocking=True)

    heights = heights.repeat_interleave(total_crops_per_image)
    widths = widths.repeat_interleave(total_crops_per_image)
    scales = torch.empty(n * total_crops_per_image, device=device)

    # Fill scales and compute dimensions
    global_idx = torch.arange(n, device=device).repeat_interleave(n_global_crops)
    global_offsets = torch.arange(n_global_crops, device=device).repeat(n)
    scales[global_idx * total_crops_per_image + global_offsets] = torch.empty(
        n * n_global_crops, device=device
    ).uniform_(global_scale_range[0], global_scale_range[1])

    local_idx = torch.arange(n, device=device).repeat_interleave(n_local_crops)
    local_offsets = torch.arange(n_local_crops, device=device).repeat(n)
    scales[local_idx * total_crops_per_image + n_global_crops + local_offsets] = torch.empty(
        n * n_local_crops, device=device
    ).uniform_(local_scale_range[0], local_scale_range[1])

    # Generate crop dimensions (exp log for non biased uniform)
    ratios = torch.exp(
        torch.empty(n * total_crops_per_image, device=device).uniform_(
            math.log(ratio_range[0]), math.log(ratio_range[1])
        )
    )
    target_areas = heights * widths * scales
    w = (torch.sqrt(target_areas * ratios)).round().long()
    h = (torch.sqrt(target_areas / ratios)).round().long()
    h = torch.minimum(h, heights)
    w = torch.minimum(w, widths)

    # Generate positions and stack coordinates
    max_x = (widths - w).clamp(min=0)
    max_y = (heights - h).clamp(min=0)
    x1 = (torch.rand(n * total_crops_per_image, device=device) * max_x).long()
    y1 = (torch.rand(n * total_crops_per_image, device=device) * max_y).long()
    batch_indices = torch.arange(n, device=device).repeat_interleave(total_crops_per_image)

    crop_coords = torch.stack(
        [
            batch_indices.float(),
            x1.float(),
            y1.float(),
            (x1 + w).float(),
            (y1 + h).float(),
            heights.float(),
            widths.float(),
        ],
        dim=1,
    )

    # Reshape to align with dinov2 format
    crop_coords = crop_coords.view(n, total_crops_per_image, 7)
    crop_coords_global = crop_coords[:, :n_global_crops, :].transpose(0, 1).reshape(-1, 7)
    crop_coords_local = crop_coords[:, n_global_crops:, :].transpose(0, 1).reshape(-1, 7)

    return padded, crop_coords_global, crop_coords_local


def pad_tensor_to_gpu(
    tensor_list: list[torch.Tensor],
    pad_size: Tuple[int, int, int],
    n_global_crops: int,
    n_local_crops: int,
    device: torch.device,
    global_scale_range: Tuple[float, float],
    local_scale_range: Tuple[float, float],
    ratio_range: Tuple[float, float] = (3 / 4, 4 / 3),
    custom_stream: torch.cuda.Stream | None = None,
) -> dict[str, torch.Tensor]:
    """
    Pads a list of tensors to the specified size and prepares crop coordinates.

    Parameters
    ----------
    tensor_list : list of torch.Tensor
        List of tensors to be padded
    pad_size : tuple of int
        Size to pad tensors to (D, H, W)
    n_global_crops : int
        Number of global crops to generate
    n_local_crops : int
        Number of local crops to generate
    device : torch.device
        Device to run operations on
    global_scale_range : tuple of float
        Scale range for global crops
    local_scale_range : tuple of float
        Scale range for local crops
    ratio_range : tuple of float, optional
        Aspect ratio range for all crops, default is (3/4, 4/3)
    custom_stream : torch.cuda.Stream or None, optional
        CUDA stream to use for non-blocking operations

    Returns
    -------
    dict
        Dictionary containing:
        - 'padded_tensor': Padded tensor on GPU
        - 'global_crop_coords': Coordinates for global crops
        - 'local_crop_coords': Coordinates for local crops
    """
    padded_tensor, global_crop_coords, local_crop_coords = pad_tensors_and_compute_crop_coords(
        tensors=tensor_list,
        pad_size=pad_size[1:],
        n_global_crops=n_global_crops,
        n_local_crops=n_local_crops,
        global_scale_range=global_scale_range,
        local_scale_range=local_scale_range,
        ratio_range=ratio_range,
        custom_stream=custom_stream,
        device=device,
    )

    return {
        "padded_tensor": padded_tensor,
        "global_crop_coords": global_crop_coords,
        "local_crop_coords": local_crop_coords,
    }
