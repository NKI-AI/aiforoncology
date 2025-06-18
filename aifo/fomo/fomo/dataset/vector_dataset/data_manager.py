import random

from fomo.database_models import Image
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from sqlalchemy.sql.expression import func


class DataManager:
    def __init__(self, database_url: str):
        self.engine = create_engine(database_url)
        Session = sessionmaker(bind=self.engine)
        self.session = Session()
        # I assume we have no gaps in the IDs
        self.min_id = self.session.query(func.min(Image.id)).scalar()
        self.max_id = self.session.query(func.max(Image.id)).scalar()

    def get_worker_id_range(self, worker_id: int, num_workers: int):
        """Calculate the ID range assigned to a worker."""
        total_range = self.max_id - self.min_id + 1
        range_per_worker = total_range // num_workers
        start_id = self.min_id + worker_id * range_per_worker
        # Last worker takes any remaining IDs
        if worker_id == num_workers - 1:
            end_id = self.max_id
        else:
            end_id = start_id + range_per_worker - 1
        return start_id, end_id

    def get_random_image_generator(self, worker_id: int, num_workers: int):
        """Generator to yield random images assigned to a worker."""
        start_id, end_id = self.get_worker_id_range(worker_id, num_workers)
        print(f"Worker {worker_id}: {start_id} to {end_id}")
        # Initialize a random generator with a unique seed per worker
        random_seed = worker_id
        random_generator = random.Random(random_seed)
        while True:
            random_id = random_generator.randint(start_id, end_id)
            db_entry = self.session.query(Image).filter(Image.id == random_id).first()
            if db_entry:
                yield db_entry
