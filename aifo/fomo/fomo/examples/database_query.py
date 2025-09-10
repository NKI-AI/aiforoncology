import random

from fomo.database_models import Base, Image
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from sqlalchemy.sql import func


def get_random_image_mask_pair(session, total_images: int):
    """
    Fetch a random image and its associated mask from the database.

    Parameters
    ----------
    session : SQLAlchemy session
        The database session.

    Returns
    -------
    tuple
        A tuple containing the random image and mask pair (Image, Mask).
    """

    # Generate a random offset
    random_offset = random.randint(0, total_images - 1)
    random_image = session.query(Image).offset(random_offset).first()

    # Ensure the image has an associated mask
    if random_image and random_image.mask:
        return random_image, random_image.mask
    else:
        raise Exception("Random image does not have an associated mask.")


# Example usage:
def main():
    engine = create_engine("sqlite:///database.sqlite")
    Base.metadata.create_all(engine)

    Session = sessionmaker(bind=engine)
    session = Session()

    total_images = session.query(func.count(Image.id)).scalar()

    try:
        image, mask = get_random_image_mask_pair(session, total_images)
        print(f"Random Image: {image}")
        print(f"Associated Mask: {mask}")
    except Exception as e:
        print(f"Error fetching random image-mask pair: {e}")


if __name__ == "__main__":
    main()
