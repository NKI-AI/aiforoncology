"""Declares the models used for creating a database of medical images."""

from sqlalchemy import ForeignKey, Integer, String
from sqlalchemy.orm import Mapped, declarative_base, mapped_column, relationship
from sqlalchemy.types import JSON

Base = declarative_base()


class Image(Base):
    __tablename__ = "images"

    id: Mapped[int] = mapped_column(Integer, primary_key=True)
    filename: Mapped[str] = mapped_column(String, unique=True, nullable=False)
    origin: Mapped[list] = mapped_column(JSON)  # Store origin as a JSON object
    spacing: Mapped[list] = mapped_column(JSON)  # Store spacing as a JSON object
    direction: Mapped[list] = mapped_column(JSON)  # Store direction as a JSON object
    shape: Mapped[list] = mapped_column(JSON)  # Store shape as a JSON object
    mask_id: Mapped[int] = mapped_column(Integer, ForeignKey("masks.id"), nullable=True)  # Allow NULL values
    mask = relationship("Mask", back_populates="images")

    def __repr__(self):
        return f"<Image(id={self.id}, filename={self.filename}), shape={self.shape}>"


class Mask(Base):
    __tablename__ = "masks"

    id: Mapped[int] = mapped_column(Integer, primary_key=True)
    filename: Mapped[str] = mapped_column(String, nullable=False)  # Ensure filename is not NULL
    origin: Mapped[list] = mapped_column(JSON)  # Store origin as a JSON object
    spacing: Mapped[list] = mapped_column(JSON)  # Store spacing as a JSON object
    direction: Mapped[list] = mapped_column(JSON)  # Store direction as a JSON object
    shape: Mapped[list] = mapped_column(JSON)  # Store shape as a JSON object
    bbox: Mapped[list] = mapped_column(JSON)  # Store bounding box as a JSON object

    images = relationship("Image", back_populates="mask")  # One-to-many: a mask can have multiple images

    def __repr__(self):
        return f"<Mask(id={self.id}, filename={self.filename}), shape={self.shape}, bbox={self.bbox}>"
