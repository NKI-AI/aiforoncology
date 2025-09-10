import os

from sqlalchemy import create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker

DATABASE_PATH = os.environ.get("DATABASE_PATH", None)
if DATABASE_PATH is None:
    raise ValueError("Please set the DATABASE_PATH environment variable.")

# Define the database URL and connect
DATABASE_URL = "sqlite:///{0}".format(DATABASE_PATH)
print("Database URL:", DATABASE_URL)
engine = create_engine(DATABASE_URL)
Session = sessionmaker(bind=engine)
session = Session()

# Define models
Base = declarative_base()

# Create tables if they don't exist
Base.metadata.create_all(engine)
