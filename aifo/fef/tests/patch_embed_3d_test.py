import pytest
import torch
from fef.models.layers import PatchEmbed3d


class TestPatchEmbed3d:
    embed_dim: int = 768

    @pytest.fixture()
    def layer(self):
        patch_embed = PatchEmbed3d(
            img_size=(16, 224, 224),
            patch_size=(16, 16, 16),
            in_chans=1,
            embed_dim=self.embed_dim,
            norm_layer=None,
            strict_img_size=True,
        )

        return patch_embed

    def test_patch_embed_correctly_creates_patches(self, layer):
        # arrange
        test_tensor = torch.randn(1, 1, 16, 224, 224)

        # act
        patch_embeddings = layer(test_tensor)

        # assert
        assert patch_embeddings.shape == (1, 196, self.embed_dim)

    def test_patch_embed_correctly_creates_patches_dynamic_img_size(self, layer):
        # arrange
        test_tensor = torch.randn(1, 1, 32, 224, 224)
        layer._strict_img_size = False

        # act
        patch_embeddings = layer(test_tensor)

        # assert
        assert patch_embeddings.shape == (1, 2, 14, 14, self.embed_dim)

    def test_patch_embed_height_is_not_multiple_of_patch_size(self, layer):
        # arrange
        test_tensor = torch.randn(1, 1, 16, 226, 224)

        # act & assert
        with pytest.raises(ValueError, match="height"):
            layer(test_tensor)

    def test_patch_embed_width_is_not_multiple_of_patch_size(self, layer):
        # arrange
        test_tensor = torch.randn(1, 1, 16, 224, 222)

        # act & assert
        with pytest.raises(ValueError, match="width"):
            layer(test_tensor)

    def test_patch_embed_depth_is_not_multiple_of_patch_size(self, layer):
        # arrange
        test_tensor = torch.randn(1, 1, 14, 224, 224)

        # act & assert
        with pytest.raises(ValueError, match="depth"):
            layer(test_tensor)
