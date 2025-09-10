import pytest
import torch

from dinov2.models.vision_transformer import DinoVisionTransformer3d


class TestDinoVisionTransformer3d:
    @pytest.fixture
    def model(self) -> DinoVisionTransformer3d:
        """Create a minimal DinoVisionTransformer3d instance"""
        model = DinoVisionTransformer3d(
            img_size=(224, 224, 224),
            patch_size=(16, 16, 16),
            in_chans=1,
            embed_dim=128,
            depth=2,
            num_heads=2,
        )

        # Create a class token with ones
        class_token = torch.ones(1, 1, model.embed_dim)

        patch_depth, patch_height, patch_width = model.patch_size
        num_patches_d, num_patches_h, num_patches_w = model.patch_embed.patches_resolution

        # Create patch embeddings with values that reflect their 3D position
        # Initialize with zeros
        patch_embed = torch.zeros(1, num_patches_d * num_patches_h * num_patches_w, model.embed_dim)

        # Fill in the first few dimensions of each embedding with the normalized d,h,w coordinates
        # This creates a predictable 3D gradient pattern in the embeddings
        for d in range(num_patches_d):
            for h in range(num_patches_h):
                for w in range(num_patches_w):
                    idx = d * num_patches_d**2 + h * num_patches_h + w
                    # Use first 3 dimensions to store d,h,w position (0 to 1)
                    patch_embed[0, idx, 0] = d / (num_patches_d - 1)  # d position (normalized)
                    patch_embed[0, idx, 1] = h / (num_patches_h - 1)  # h position (normalized)
                    patch_embed[0, idx, 2] = w / (num_patches_w - 1)  # w position (normalized)

        model.pos_embed = torch.nn.Parameter(torch.cat([class_token, patch_embed], dim=1))

        return model

    def test_no_interpolation_needed(self, model: DinoVisionTransformer3d) -> None:
        """Test when interpolation is not needed (first if-statement is true)"""

        patch_depth, patch_height, patch_width = model.patch_size
        num_patches_d, num_patches_h, num_patches_w = model.patch_embed.patches_resolution

        # Arrange
        img_depth = patch_depth * num_patches_d
        img_height = patch_height * num_patches_h
        img_width = patch_width * num_patches_w
        embeddings = torch.randn(1, 1 + num_patches_d * num_patches_h * num_patches_w, model.embed_dim)

        # Act
        result = model._interpolate_pos_encoding(embeddings, (img_width, img_height, img_depth))

        # Assert
        assert torch.equal(result, model.pos_embed)

    @pytest.mark.parametrize(
        "scale_factor",
        [
            # Upsampling: double the dimensions
            2.0,
            # Downsampling: halve the dimensions
            0.5,
        ],
        ids=["upsampling", "downsampling"],
    )
    def test_interpolation(self, model: DinoVisionTransformer3d, scale_factor: float) -> None:
        """Test both up- and downsampling of position embeddings with verification of interpolated values"""

        patch_depth, patch_height, patch_width = model.patch_size
        num_patches_d, num_patches_h, num_patches_w = model.patch_embed.patches_resolution

        # Arrange
        original_depth = patch_depth * num_patches_d
        original_height = patch_height * num_patches_h
        original_width = patch_width * num_patches_w

        # Calculate new dimensions
        img_depth = int(original_depth * scale_factor)
        img_height = int(original_height * scale_factor)
        img_width = int(original_width * scale_factor)

        # Create embeddings that would match the scaled dimensions
        new_num_patches_d, new_num_patches_h, new_num_patches_w = (
            (img_depth // patch_depth),
            (img_height // patch_height),
            (img_width // patch_width),
        )
        new_patch_count = new_num_patches_d * new_num_patches_h * new_num_patches_w
        embeddings = torch.randn(1, 1 + new_patch_count, model.embed_dim)

        # Act
        result = model._interpolate_pos_encoding(embeddings, (img_width, img_height, img_depth))

        # Assert
        # 1. Verify shape
        expected_shape = (1, 1 + new_patch_count, model.embed_dim)
        assert result.shape == expected_shape

        # 2. Verify class token remains unchanged
        assert torch.equal(result[:, 0:1, :], model.pos_embed[:, 0:1, :])

        # 3. Verify the interpolation behavior by looking at specific 3D positions
        # Convert flattened patch embeddings back to 3D grid for easier testing
        patches_3d = result[:, 1:, :].reshape(
            1, new_num_patches_d, new_num_patches_h, new_num_patches_w, model.embed_dim
        )

        # Test corners and middle to verify 3D interpolation worked correctly

        # Origin corner (0,0,0) should be close to the original (0,0,0) position
        corner_000 = patches_3d[0, 0, 0, 0, :3]
        assert torch.allclose(corner_000, torch.tensor([0.0, 0.0, 0.0]), atol=0.1)

        # Far corner (max,max,max) should be close to the original (1,1,1) position
        corner_111 = patches_3d[0, -1, -1, -1, :3]
        assert torch.allclose(corner_111, torch.tensor([1.0, 1.0, 1.0]), atol=0.1)

        # Middle position should be close to (0.5,0.5,0.5)
        mid_d, mid_h, mid_w = new_num_patches_d // 2, new_num_patches_h // 2, new_num_patches_w // 2
        middle = patches_3d[0, mid_d, mid_h, mid_w, :3]
        assert torch.allclose(middle, torch.tensor([0.5, 0.5, 0.5]), atol=0.1)

        # Test intermediate positions along each axis
        # These should show gradual change in the corresponding dimension value

        # Verify d-axis interpolation (changing only d coordinate)
        d_positions = [
            patches_3d[0, d, 0, 0, 0].item() for d in range(0, new_num_patches_d, max(1, new_num_patches_d // 5))
        ]
        assert all(d_positions[i] < d_positions[i + 1] for i in range(len(d_positions) - 1))

        # Verify h-axis interpolation (changing only h coordinate)
        h_positions = [
            patches_3d[0, 0, h, 0, 1].item() for h in range(0, new_num_patches_h, max(1, new_num_patches_h // 5))
        ]
        assert all(h_positions[i] < h_positions[i + 1] for i in range(len(h_positions) - 1))

        # Verify w-axis interpolation (changing only w coordinate)
        w_positions = [
            patches_3d[0, 0, 0, w, 2].item() for w in range(0, new_num_patches_w, max(1, new_num_patches_w // 5))
        ]
        assert all(w_positions[i] < w_positions[i + 1] for i in range(len(w_positions) - 1))

    @pytest.mark.parametrize("img_size", [(1, 1, 64, 64, 64), (1, 1, 32, 64, 128)])
    def test_forward_features_returns_expected_shape(
        self, model: DinoVisionTransformer3d, img_size: tuple[int, int, int]
    ) -> None:
        patch_depth, patch_height, patch_width = model.patch_size
        num_patches_d, num_patches_h, num_patches_w = model.patch_embed.patches_resolution
        # arrange
        img = torch.randn(img_size)
        img_d, img_h, img_w = img_size[-3:]
        num_patches_d, num_patches_h, num_patches_w = (
            (img_d // patch_depth),
            (img_h // patch_height),
            (img_w // patch_width),
        )
        # Add one for the CLS-token
        total_patches = num_patches_d * num_patches_h * num_patches_w

        # act
        result_dict = model.forward_features(img, masks=None)

        # assert
        # Patch tokens have shape of B x N_PATCHES x C
        assert result_dict["x_norm_patchtokens"].shape == (1, total_patches, model.embed_dim)
        # CLS tokens have shape of B x C
        assert result_dict["x_norm_clstoken"].shape == (1, model.embed_dim)
