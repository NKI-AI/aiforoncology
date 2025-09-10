import datetime
from typing import List, Optional

from sqlalchemy import DateTime, Float, ForeignKey, Index, Integer, Text
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship


class Base(DeclarativeBase):
    pass


class Instances(Base):
    __tablename__ = "instances"
    __table_args__ = (
        Index("idx_instance_uid", "sop_instance_uid", unique=True),
        Index("idx_instances_deleted_at", "deleted_at"),
        Index("idx_instances_modality", "modality"),
        Index("idx_instances_series_instance_uid", "series_instance_uid"),
        Index("idx_series_instance", "series_id"),
    )

    id: Mapped[Optional[int]] = mapped_column(Integer, primary_key=True)
    created_at: Mapped[Optional[datetime.datetime]] = mapped_column(DateTime)
    updated_at: Mapped[Optional[datetime.datetime]] = mapped_column(DateTime)
    deleted_at: Mapped[Optional[datetime.datetime]] = mapped_column(DateTime)
    columns: Mapped[Optional[int]] = mapped_column(Integer)
    content_date_time: Mapped[Optional[datetime.datetime]] = mapped_column(DateTime)
    series_id: Mapped[Optional[int]] = mapped_column(Integer)
    sop_instance_uid: Mapped[Optional[str]] = mapped_column(Text)
    sop_class_uid: Mapped[Optional[str]] = mapped_column(Text)
    acquisition_number: Mapped[Optional[int]] = mapped_column(Integer)
    instance_number: Mapped[Optional[int]] = mapped_column(Integer)
    rows: Mapped[Optional[int]] = mapped_column(Integer)
    slice_location: Mapped[Optional[str]] = mapped_column(Text)
    dicom_file_path: Mapped[Optional[str]] = mapped_column(Text)
    acquisition_date_time: Mapped[Optional[datetime.datetime]] = mapped_column(DateTime)
    instance_creation_date_time: Mapped[Optional[datetime.datetime]] = mapped_column(DateTime)
    modality: Mapped[Optional[str]] = mapped_column(Text)
    series_instance_uid: Mapped[Optional[str]] = mapped_column(ForeignKey("series.series_instance_uid"))

    series: Mapped["Series"] = relationship("Series", foreign_keys=[series_instance_uid], back_populates="instances")
    series_: Mapped[List["Series"]] = relationship(
        "Series", foreign_keys="[Series.series_instance_uid]", back_populates="instances_"
    )


class Patients(Base):
    __tablename__ = "patients"
    __table_args__ = (
        Index("idx_patient_mrn", "patient_mrn", unique=True),
        Index("idx_patients_deleted_at", "deleted_at"),
        Index("idx_patients_patient_birth_date", "patient_birth_date"),
        Index("idx_patients_patient_name", "patient_name"),
    )

    id: Mapped[Optional[int]] = mapped_column(Integer, primary_key=True)
    created_at: Mapped[Optional[datetime.datetime]] = mapped_column(DateTime)
    updated_at: Mapped[Optional[datetime.datetime]] = mapped_column(DateTime)
    deleted_at: Mapped[Optional[datetime.datetime]] = mapped_column(DateTime)
    patient_mrn: Mapped[Optional[str]] = mapped_column(Text)
    patient_name: Mapped[Optional[str]] = mapped_column(Text)
    patient_birth_date: Mapped[Optional[str]] = mapped_column(Text)

    studies: Mapped[List["Studies"]] = relationship("Studies", back_populates="patient")


class Series(Base):
    __tablename__ = "series"
    __table_args__ = (
        Index("idx_series_deleted_at", "deleted_at"),
        Index("idx_series_modality", "modality"),
        Index("idx_series_study_instance_uid", "study_instance_uid"),
        Index("idx_series_uid", "series_instance_uid", unique=True),
        Index("idx_study_series", "study_id"),
    )

    id: Mapped[Optional[int]] = mapped_column(Integer, primary_key=True)
    created_at: Mapped[Optional[datetime.datetime]] = mapped_column(DateTime)
    updated_at: Mapped[Optional[datetime.datetime]] = mapped_column(DateTime)
    deleted_at: Mapped[Optional[datetime.datetime]] = mapped_column(DateTime)
    study_id: Mapped[Optional[int]] = mapped_column(Integer)
    series_instance_uid: Mapped[Optional[str]] = mapped_column(ForeignKey("instances.series_instance_uid"))
    series_number: Mapped[Optional[str]] = mapped_column(Text)
    series_description: Mapped[Optional[str]] = mapped_column(Text)
    modality: Mapped[Optional[str]] = mapped_column(Text)
    protocol_name: Mapped[Optional[str]] = mapped_column(Text)
    body_part_examined: Mapped[Optional[str]] = mapped_column(Text)
    contrast_bolus_agent: Mapped[Optional[str]] = mapped_column(Text)
    frame_of_reference_uid: Mapped[Optional[str]] = mapped_column(Text)
    planar_configuration: Mapped[Optional[str]] = mapped_column(Text)
    patient_position: Mapped[Optional[str]] = mapped_column(Text)
    institution_name: Mapped[Optional[str]] = mapped_column(Text)
    instance_count: Mapped[Optional[int]] = mapped_column(Integer)
    acquisition_duration: Mapped[Optional[str]] = mapped_column(Text)
    manufacturer: Mapped[Optional[str]] = mapped_column(Text)
    manufacturer_model_name: Mapped[Optional[str]] = mapped_column(Text)
    series_date_time: Mapped[Optional[datetime.datetime]] = mapped_column(DateTime)
    slice_thickness: Mapped[Optional[str]] = mapped_column(Text)
    study_instance_uid: Mapped[Optional[str]] = mapped_column(ForeignKey("studies.study_instance_uid"))

    instances: Mapped[List["Instances"]] = relationship(
        "Instances", foreign_keys="[Instances.series_instance_uid]", back_populates="series"
    )
    instances_: Mapped["Instances"] = relationship(
        "Instances", foreign_keys=[series_instance_uid], back_populates="series_"
    )
    studies: Mapped["Studies"] = relationship("Studies", foreign_keys=[study_instance_uid], back_populates="series")
    studies_: Mapped[List["Studies"]] = relationship(
        "Studies", foreign_keys="[Studies.study_instance_uid]", back_populates="series_"
    )
    mri_specifics: Mapped[List["MriSpecifics"]] = relationship("MriSpecifics", back_populates="series")


class Studies(Base):
    __tablename__ = "studies"
    __table_args__ = (
        Index("idx_patient_study", "patient_id"),
        Index("idx_studies_deleted_at", "deleted_at"),
        Index("idx_studies_patient_sex", "patient_sex"),
        Index("idx_studies_patient_weight", "patient_weight"),
        Index("idx_study_uid", "study_instance_uid", unique=True),
    )

    id: Mapped[Optional[int]] = mapped_column(Integer, primary_key=True)
    created_at: Mapped[Optional[datetime.datetime]] = mapped_column(DateTime)
    updated_at: Mapped[Optional[datetime.datetime]] = mapped_column(DateTime)
    deleted_at: Mapped[Optional[datetime.datetime]] = mapped_column(DateTime)
    patient_id: Mapped[Optional[int]] = mapped_column(ForeignKey("patients.id"))
    patient_sex: Mapped[Optional[str]] = mapped_column(Text)
    patient_weight: Mapped[Optional[str]] = mapped_column(Text)
    study_instance_uid: Mapped[Optional[str]] = mapped_column(ForeignKey("series.study_instance_uid"))
    study_description: Mapped[Optional[str]] = mapped_column(Text)
    study_id: Mapped[Optional[str]] = mapped_column(Text)
    study_date_time: Mapped[Optional[datetime.datetime]] = mapped_column(DateTime)
    accession_number: Mapped[Optional[str]] = mapped_column(Text)

    series: Mapped[List["Series"]] = relationship(
        "Series", foreign_keys="[Series.study_instance_uid]", back_populates="studies"
    )
    patient: Mapped["Patients"] = relationship("Patients", back_populates="studies")
    series_: Mapped["Series"] = relationship("Series", foreign_keys=[study_instance_uid], back_populates="studies_")


class MriSpecifics(Base):
    __tablename__ = "mri_specifics"
    __table_args__ = (
        Index("idx_mri_series", "series_id", unique=True),
        Index("idx_mri_specifics_deleted_at", "deleted_at"),
    )

    id: Mapped[Optional[int]] = mapped_column(Integer, primary_key=True)
    created_at: Mapped[Optional[datetime.datetime]] = mapped_column(DateTime)
    updated_at: Mapped[Optional[datetime.datetime]] = mapped_column(DateTime)
    deleted_at: Mapped[Optional[datetime.datetime]] = mapped_column(DateTime)
    series_id: Mapped[Optional[int]] = mapped_column(ForeignKey("series.id"))
    receive_coil_name: Mapped[Optional[str]] = mapped_column(Text)
    number_of_frames: Mapped[Optional[int]] = mapped_column(Integer)
    samples_per_pixel: Mapped[Optional[int]] = mapped_column(Integer)
    photometric_interpretation: Mapped[Optional[str]] = mapped_column(Text)
    bits_stored: Mapped[Optional[int]] = mapped_column(Integer)
    image_type: Mapped[Optional[str]] = mapped_column(Text)
    acquisition_matrix: Mapped[Optional[str]] = mapped_column(Text)
    acquisition_duration: Mapped[Optional[str]] = mapped_column(Text)
    angio_flag: Mapped[Optional[str]] = mapped_column(Text)
    beat_rejection_flag: Mapped[Optional[str]] = mapped_column(Text)
    bits_allocated: Mapped[Optional[int]] = mapped_column(Integer)
    db_dt: Mapped[Optional[str]] = mapped_column(Text)
    echo_number: Mapped[Optional[int]] = mapped_column(Integer)
    echo_planer_pulse_sequence: Mapped[Optional[str]] = mapped_column(Text)
    echo_time: Mapped[Optional[float]] = mapped_column(Float)
    echo_train_length: Mapped[Optional[int]] = mapped_column(Integer)
    flip_angle: Mapped[Optional[str]] = mapped_column(Text)
    high_bit: Mapped[Optional[str]] = mapped_column(Text)
    high_rr_value: Mapped[Optional[str]] = mapped_column(Text)
    imaged_nucleus: Mapped[Optional[str]] = mapped_column(Text)
    images_in_acquisition: Mapped[Optional[int]] = mapped_column(Integer)
    imaging_frequency: Mapped[Optional[str]] = mapped_column(Text)
    in_plane_phase_encoding_direction: Mapped[Optional[str]] = mapped_column(Text)
    intervals_acquired: Mapped[Optional[str]] = mapped_column(Text)
    intervals_rejected: Mapped[Optional[str]] = mapped_column(Text)
    inversion_time: Mapped[Optional[float]] = mapped_column(Float)
    low_rr_value: Mapped[Optional[str]] = mapped_column(Text)
    magnetic_field_strength: Mapped[Optional[str]] = mapped_column(Text)
    multi_planar_excitation: Mapped[Optional[str]] = mapped_column(Text)
    multiple_spin_echo: Mapped[Optional[str]] = mapped_column(Text)
    nominal_interval: Mapped[Optional[str]] = mapped_column(Text)
    number_of_averages: Mapped[Optional[float]] = mapped_column(Float)
    number_of_phase_encoding_steps: Mapped[Optional[int]] = mapped_column(Integer)
    number_of_temporal_positions: Mapped[Optional[int]] = mapped_column(Integer)
    percent_phase_field_of_view: Mapped[Optional[str]] = mapped_column(Text)
    percent_sampling: Mapped[Optional[str]] = mapped_column(Text)
    pixel_aspect_ratio: Mapped[Optional[str]] = mapped_column(Text)
    pixel_bandwidth: Mapped[Optional[str]] = mapped_column(Text)
    pixel_spacing: Mapped[Optional[str]] = mapped_column(Text)
    planar_configuration: Mapped[Optional[str]] = mapped_column(Text)
    pulse_sequence_name: Mapped[Optional[str]] = mapped_column(Text)
    reconstruction_diameter: Mapped[Optional[str]] = mapped_column(Text)
    repetition_time: Mapped[Optional[float]] = mapped_column(Float)
    saturation_recovery: Mapped[Optional[str]] = mapped_column(Text)
    scanning_sequence: Mapped[Optional[str]] = mapped_column(Text)
    scan_options: Mapped[Optional[str]] = mapped_column(Text)
    sequence_name: Mapped[Optional[str]] = mapped_column(Text)
    sequence_variant: Mapped[Optional[str]] = mapped_column(Text)
    slice_thickness: Mapped[Optional[float]] = mapped_column(Float)
    spacing_between_slices: Mapped[Optional[float]] = mapped_column(Float)
    steady_state_pulse_sequence: Mapped[Optional[str]] = mapped_column(Text)
    temporal_position_identifier: Mapped[Optional[str]] = mapped_column(Text)
    temporal_resolution: Mapped[Optional[float]] = mapped_column(Float)
    transmit_coil_name: Mapped[Optional[str]] = mapped_column(Text)
    variable_flip_angle_flag: Mapped[Optional[str]] = mapped_column(Text)
    pulse_sequence: Mapped[Optional[int]] = mapped_column(Integer)
    acquisition_contrast: Mapped[Optional[str]] = mapped_column(Text)
    water_fat_shift: Mapped[Optional[float]] = mapped_column(Float)

    series: Mapped["Series"] = relationship("Series", back_populates="mri_specifics")
