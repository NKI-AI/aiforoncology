import torch
import torch.nn as nn


class FullyConnectedNet(nn.Module):
    def __init__(
        self,
        img_dim: int,
        num_classes: int,
        num_channels: int = 3,
        num_layers: int = 3,
        width: int = 64,
    ):
        super().__init__()

        self.num_classes = num_classes

        self.layers = nn.ModuleList()
        self.layers.append(
            nn.Sequential(
                nn.Linear(img_dim * img_dim * num_channels, width),
                nn.ReLU(),
            )
        )

        for i in range(1, num_layers):
            self.layers.append(
                nn.Sequential(
                    nn.Linear(width, width),
                    nn.ReLU(),
                )
            )

        self.backbone = self.layers

        self.head = nn.Linear(width, num_classes)

    def forward(self, x: torch.Tensor) -> dict:
        x = x.view(x.size(0), -1)
        for i, layer in enumerate(self.layers):
            x = layer(x)

        out = x.view(x.size(0), -1)
        out = self.head(out)

        return out

    def freeze_backbone(self):
        for layer in self.backbone:
            for param in layer.parameters():
                param.requires_grad = False

    def reinitialize_head(self):
        torch.nn.init.kaiming_normal_(self.head.weight)
        torch.nn.init.zeros_(self.head.bias)

    def set_linear_head(self):
        pass
