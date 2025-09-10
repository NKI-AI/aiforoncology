import torch
import torch.nn as nn


class DummyModel(nn.Module):
    """
    A dummy model that generates a one-hot-encoded segmentation map for an input image.

    The output segmentation map assigns:
    - Label 0 to the background (everything outside the image).
    - Label 1 to the interior of the image.
    - Label 2 to the border of the image.

    The output is a one-hot-encoded tensor with shape (N, 3, H, W), where:
    - The first channel (axis 0) corresponds to the background (class 0).
    - The second channel (axis 1) corresponds to the interior (class 1).
    - The third channel (axis 2) corresponds to the border (class 2).

    Parameters
    ----------
    border_width : int, optional
        Width of the border to be labeled as 2. Default is 1.

    Examples
    --------
    >>> model = DummyModel(border_width=2)
    >>> input_image = torch.randn(1, 3, 8, 8)  # (N, C, H, W)
    >>> segmentation_map = model(input_image)
    >>> print(segmentation_map.shape)
    torch.Size([1, 3, 8, 8])
    """

    def __init__(self, border_width: int = 1):
        super(DummyModel, self).__init__()
        self.border_width = border_width

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        """
        Generate the one-hot-encoded segmentation map for the input image.

        Parameters
        ----------
        x : torch.Tensor
            Input tensor of shape (N, C, H, W), where:
            - N is the batch size.
            - C is the number of channels.
            - H and W are the spatial dimensions.

        Returns
        -------
        torch.Tensor
            One-hot-encoded segmentation map of shape (N, 3, H, W), where:
            - Channel 0 corresponds to the background (class 0).
            - Channel 1 corresponds to the interior (class 1).
            - Channel 2 corresponds to the border (class 2).

        Raises
        ------
        AssertionError
            If the input tensor does not have 4 dimensions (N, C, H, W).
        """
        assert x.ndim == 4, "Input must have shape (N, C, H, W)"

        # Get spatial dimensions
        N, _, H, W = x.shape

        # Initialize one-hot channels for background, interior, and border
        background = torch.ones((N, H, W), device=x.device, dtype=torch.float32)
        interior = torch.zeros((N, H, W), device=x.device, dtype=torch.float32)
        border = torch.zeros((N, H, W), device=x.device, dtype=torch.float32)

        # Mark the interior with 1s
        interior[:, self.border_width : -self.border_width, self.border_width : -self.border_width] = 1

        # Mark the border with 1s
        border[:, : self.border_width, :] = 1  # Top border
        border[:, -self.border_width :, :] = 1  # Bottom border
        border[:, :, : self.border_width] = 1  # Left border
        border[:, :, -self.border_width :] = 1  # Right border

        # Update background: subtract interior and border
        background -= (interior + border).clamp(max=1)

        # Stack channels to create the final one-hot tensor
        one_hot_map = torch.stack([background, interior, border], dim=1)

        return one_hot_map
