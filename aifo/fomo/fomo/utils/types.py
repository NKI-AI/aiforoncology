from enum import Enum


class FomoEnum(Enum):
    @classmethod
    def from_value(cls, value):
        for orientation in cls:
            if orientation.value == value:
                return orientation
        raise ValueError(f"Invalid orientation value: {value}")


class Orientation(FomoEnum):
    LPS = "LPS"
    PSL = "PSL"


class Dimensionality(FomoEnum):
    D2 = "2D"
    D3 = "3D"
