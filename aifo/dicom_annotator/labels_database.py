from sqlalchemy import Column, Integer, String, create_engine
from sqlalchemy.orm import declarative_base, sessionmaker

# Define a new Base for the labels database
LabelsBase = declarative_base()


class SeriesLabel(LabelsBase):
    """
    SQLAlchemy ORM model for the series_labels table.
    """

    __tablename__ = "series_labels"

    id = Column(Integer, primary_key=True, autoincrement=True)
    series_instance_uid = Column(String, unique=True, nullable=False)
    study_instance_uid = Column(String, nullable=False)
    series_description = Column(String)
    study_description = Column(String)
    series_type = Column(String)


# Function to initialize the labels database
def init_labels_db(db_url):
    """
    Initializes the labels database.

    Args:
        db_url (str): Database connection string. Defaults to SQLite.

    Returns:
        session: SQLAlchemy session for the labels database.
    """
    # Create the engine
    labels_engine = create_engine(db_url, echo=False)

    # Create all tables in the labels database (if they don't exist)
    LabelsBase.metadata.create_all(labels_engine)

    # Create a sessionmaker bound to the labels engine
    LabelsSession = sessionmaker(bind=labels_engine)

    # Create and return a session
    return LabelsSession()
