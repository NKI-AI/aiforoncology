from dataclasses import dataclass
from typing import Optional


@dataclass
class MriParameters:
    echo_time: float
    repetition_time: float
    inversion_time: float
    flip_angle: float
    pixel_bandwidth: Optional[float] = None
    image_type: Optional[str] = None


class MriSequenceClassifier:
    def __init__(self):
        self.T1_TR_THRESHOLD = 700  # ms
        self.T2_TR_THRESHOLD = 2000  # ms
        self.T1_TE_THRESHOLD = 30  # ms
        self.T2_TE_THRESHOLD = 80  # ms
        self.DWI_TE_THRESHOLD_MIN = 50  # ms
        self.DWI_TE_THRESHOLD_MAX = 150  # ms
        self.DWI_TR_THRESHOLD = 3000  # ms

    def guess_sequence_type(self, params: MriParameters) -> str:
        if params.inversion_time > 0:
            return "Inversion Recovery (e.g., STIR or FLAIR)"

        if (
            params.repetition_time > self.DWI_TR_THRESHOLD
            and self.DWI_TE_THRESHOLD_MIN < params.echo_time < self.DWI_TE_THRESHOLD_MAX
            and params.flip_angle == 90
        ):
            if params.image_type and "EPI" in params.image_type.upper():
                return "Diffusion Weighted Imaging (DWI)"

        if params.repetition_time < self.T1_TR_THRESHOLD and params.echo_time < self.T1_TE_THRESHOLD:
            return "T1-Weighted"

        if params.repetition_time > self.T2_TR_THRESHOLD and params.echo_time > self.T2_TE_THRESHOLD:
            return "T2-Weighted"

        if params.repetition_time > self.T2_TR_THRESHOLD and params.echo_time < self.T2_TE_THRESHOLD:
            return "Proton Density (PD)-Weighted"

        if (
            params.repetition_time < self.T1_TR_THRESHOLD
            and params.echo_time < self.T1_TE_THRESHOLD
            and params.flip_angle < 30
        ):
            return "Gradient Echo (T1-weighted GRE)"

        if params.pixel_bandwidth and params.pixel_bandwidth > 500:
            if params.image_type and "EPI" in params.image_type.upper():
                return "Echo Planar Imaging (EPI) - Likely DWI or fMRI"
            return "High-bandwidth Gradient Echo"

        if params.image_type and "DERIVED" in params.image_type.upper():
            return "Derived Image (e.g., post-processed sequence)"

        return "Unknown or Custom MRI Sequence"
