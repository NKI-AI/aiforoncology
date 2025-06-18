import torch
import torch.nn as nn


class SimpleConvNet(nn.Module):
    def __init__(
        self,
        img_dim: int,
        num_classes: int,
        num_channels: int = 3,
        num_layers: int = 3,
        num_filters: int = 32,
        pool_method: str = "avg",
    ):
        super().__init__()

        self.num_classes = num_classes

        pool_methods = {
            "max": nn.MaxPool2d,
            "avg": nn.AvgPool2d,
        }

        self.layers = nn.ModuleList()
        self.layers.append(
            nn.Sequential(
                nn.Conv2d(num_channels, num_filters, kernel_size=3, stride=1, padding=1),
                nn.ReLU(),
                pool_methods[pool_method](kernel_size=2, stride=2),
            )
        )

        for i in range(1, num_layers):
            self.layers.append(
                nn.Sequential(
                    nn.Conv2d(num_filters, num_filters, kernel_size=3, stride=1, padding=1),
                    nn.ReLU(),
                    pool_methods[pool_method](kernel_size=2, stride=2),
                )
            )

        self.backbone = self.layers

        final_dim = num_filters * (img_dim // 2**num_layers) ** 2

        self.head = nn.Linear(final_dim, num_classes)

    def forward(self, x: torch.Tensor) -> dict:
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
