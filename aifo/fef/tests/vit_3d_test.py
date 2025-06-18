from fef.models import ViT3d
import torch.nn as nn
import torch
import pytest
from unittest.mock import MagicMock
import types


class TestViT3d:
    embed_dim: int = 768

    @pytest.fixture
    def vit_model(self):
        # Create a mock ViT3d model
        model = MagicMock(spec=ViT3d)
        model.dynamic_img_size = True
        model.cls_token = None
        model.reg_token = None
        model.pos_embed = nn.Parameter(torch.randn(1, 4 * 7 * 7, self.embed_dim))
        model.pos_drop = lambda x: x
        model.patch_embed = MagicMock()
        model.patch_embed.grid_size = (4, 7, 7)
        model.num_prefix_tokens = 0
        model.no_embed_class = True

        # Only bind the instance method, not the static method
        bound_pos_embed = types.MethodType(ViT3d._pos_embed, model)
        model._pos_embed = bound_pos_embed
        model.resample_positional_embedding_3d = ViT3d.resample_positional_embedding_3d

        return model

    def test_resample_positional_embedding_3d_upsample(self):
        # Arrange
        old_size = (4, 7, 7)
        new_size = (8, 14, 14)
        pos_embed = nn.Parameter(torch.randn(1, 4 * 7 * 7, self.embed_dim))
        expected_shape = (1, 8 * 14 * 14, self.embed_dim)

        # Act
        resampled = ViT3d.resample_positional_embedding_3d(pos_embed=pos_embed, new_size=new_size, old_size=old_size)

        # Assert
        assert resampled.shape == expected_shape

    def test_resample_positional_embedding_3d_downsample(self):
        # Arrange
        old_size = (8, 14, 14)
        new_size = (4, 7, 7)
        pos_embed = nn.Parameter(torch.randn(1, 8 * 14 * 14, self.embed_dim))
        expected_shape = (1, 4 * 7 * 7, self.embed_dim)

        # Act
        resampled = ViT3d.resample_positional_embedding_3d(pos_embed=pos_embed, new_size=new_size, old_size=old_size)

        # Assert
        assert resampled.shape == expected_shape

    def test_resample_positional_embedding_3d_with_prefix_tokens(self):
        # Arrange
        old_size = (4, 7, 7)
        new_size = (8, 14, 14)
        num_prefix_tokens = 1
        pos_embed = nn.Parameter(torch.randn(1, num_prefix_tokens + 4 * 7 * 7, self.embed_dim))
        expected_shape = (1, num_prefix_tokens + 8 * 14 * 14, self.embed_dim)

        # Act
        resampled = ViT3d.resample_positional_embedding_3d(
            pos_embed=pos_embed, new_size=new_size, old_size=old_size, number_of_prefix_tokens=num_prefix_tokens
        )

        # Assert
        assert resampled.shape == expected_shape

    def test_resample_positional_embedding_3d_same_size(self):
        # Arrange
        size = (4, 7, 7)
        pos_embed = nn.Parameter(torch.randn(1, 4 * 7 * 7, self.embed_dim))

        # Act
        resampled = ViT3d.resample_positional_embedding_3d(pos_embed=pos_embed, new_size=size, old_size=size)

        # Assert
        assert torch.allclose(resampled, pos_embed)

    def test_pos_embed_dynamic_sizing(self, vit_model):
        # Arrange
        batch_size = 2
        # Create input with different spatial dimensions than the model's default
        input_tensor = torch.randn(batch_size, 8, 14, 14, self.embed_dim)
        expected_shape = (batch_size, 8 * 14 * 14, self.embed_dim)

        # Act
        output = vit_model._pos_embed(input_tensor)

        # Assert
        assert output.shape == expected_shape
        # Verify that the positional embedding was applied (output should be different from input)
        assert not torch.allclose(output, input_tensor.view(batch_size, -1, self.embed_dim))

    def test_pos_embed_with_class_token(self, vit_model):
        # Arrange
        vit_model.cls_token = nn.Parameter(torch.randn(1, 1, self.embed_dim))
        vit_model.num_prefix_tokens = 1
        batch_size = 2
        input_tensor = torch.randn(batch_size, 4, 7, 7, self.embed_dim)
        expected_shape = (batch_size, 1 + 4 * 7 * 7, self.embed_dim)

        # Act
        output = vit_model._pos_embed(input_tensor)

        # Assert
        assert output.shape == expected_shape
        # Verify that the class token is different from spatial tokens
        assert not torch.allclose(output[:, 0], output[:, 1])

    def test_pos_embed_with_reg_token(self, vit_model):
        # Arrange
        vit_model.reg_token = nn.Parameter(torch.randn(1, 1, self.embed_dim))
        vit_model.num_prefix_tokens = 1
        batch_size = 2
        input_tensor = torch.randn(batch_size, 4, 7, 7, self.embed_dim)
        expected_shape = (batch_size, 1 + 4 * 7 * 7, self.embed_dim)

        # Act
        output = vit_model._pos_embed(input_tensor)

        # Assert
        assert output.shape == expected_shape
        # Verify that the reg token is different from spatial tokens
        assert not torch.allclose(output[:, 0], output[:, 1])

    def test_pos_embed_with_both_tokens(self, vit_model):
        # Arrange
        vit_model.cls_token = nn.Parameter(torch.randn(1, 1, self.embed_dim))
        vit_model.reg_token = nn.Parameter(torch.randn(1, 1, self.embed_dim))
        vit_model.num_prefix_tokens = 2
        batch_size = 2
        input_tensor = torch.randn(batch_size, 4, 7, 7, self.embed_dim)
        expected_shape = (batch_size, 2 + 4 * 7 * 7, self.embed_dim)

        # Act
        output = vit_model._pos_embed(input_tensor)

        # Assert
        assert output.shape == expected_shape
        # Verify that all tokens are different from each other
        assert not torch.allclose(output[:, 0], output[:, 1])  # cls vs reg
        assert not torch.allclose(output[:, 0], output[:, 2])  # cls vs spatial
        assert not torch.allclose(output[:, 1], output[:, 2])  # reg vs spatial
