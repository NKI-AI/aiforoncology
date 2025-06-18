import pytest
import torch
from dinov2.layers.patch_embed import PatchEmbed3d


class TestPatchEmbed3d:
    @pytest.mark.parametrize("depth", [16, 32])
    def test_forward_correct_shape(self, depth):
        # arrange
        batch_size = 1
        channels = 1
        height, width = 224, 224
        input_tensor = torch.rand(batch_size, channels, depth, height, width)
        patch_embed = PatchEmbed3d(
            img_size=(depth, height, width), patch_size=(16, 16, 16), in_chans=channels, embed_dim=768
        )

        # act
        output = patch_embed(input_tensor)

        # assert
        expected_num_patches = (depth // 16) * (height // 16) * (width // 16)
        expected_shape = (batch_size, expected_num_patches, 768)
        assert output.shape == expected_shape

    def test_forward_raises_assertion_error_on_invalid_depth(self):
        # arrange
        batch_size = 1
        channels = 1
        depth, height, width = 15, 224, 224  # Depth is not a multiple of 16
        input_tensor = torch.rand(batch_size, channels, depth, height, width)
        patch_embed = PatchEmbed3d(
            img_size=(depth, height, width), patch_size=(16, 16, 16), in_chans=channels, embed_dim=768
        )

        # act & assert
        with pytest.raises(AssertionError, match="Input image depth 15 is not a multiple of patch height 16"):
            patch_embed(input_tensor)
