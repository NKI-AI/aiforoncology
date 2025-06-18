import os
import random
import time
from multiprocessing import Process
from time import sleep

import numpy as np
import psutil
import torch
from aifocore.shm.vector import DataModifiedError, OutOfMemoryError, SharedVector, remove_shared_memory


def producer(name, chunk_size_bytes, max_memory_size_bytes):
    shared_vector = SharedVector(
        name=name,
        chunk_size=chunk_size_bytes,
        max_memory_size=max_memory_size_bytes,
    )
    print("Producer: Starting append phase.")

    while True:
        num_elements = 10  # Number of elements to append
        data_shape = (1, 512, 512)  # Shape of the random arrays

        # Append random data
        try:
            for _ in range(num_elements):
                data = np.random.rand(*data_shape).astype(np.float32)
                shared_vector.append(data)
        except OutOfMemoryError:
            print("Producer: Shared vector is full during append.")
            break
        except Exception as e:
            print(f"Producer: Error during append: {e}")
            break

    print("Producer: Append phase complete.")

    # Replace loop
    print("Producer: Starting replace loop.")
    while True:
        if len(shared_vector) == 0:
            sleep(0.1)
            continue
        index = random.randint(0, len(shared_vector) - 1)
        data = np.random.rand(*data_shape).astype(np.float32)
        try:
            shared_vector.replace(index, data)
            # sleep(0.01)
        except Exception as e:
            print(f"Producer: Error during replace: {e}")
            continue


def consumer(name, chunk_size_bytes, max_memory_size_bytes):
    shared_vector = SharedVector(
        name=name,
        chunk_size=chunk_size_bytes,
        max_memory_size=max_memory_size_bytes,
    )
    print("Consumer: Starting consumption.")

    process = psutil.Process(os.getpid())
    iteration = 0
    elapsed = []
    while True:
        iteration += 1
        if iteration % 1000 == 0:
            mem_info = process.memory_info()
            print(f"Iteration {iteration}, Memory usage: {mem_info.rss / (1024 * 1024):.2f} MB")
            print(f"Average vector get over 1000 iterations: '{np.mean(elapsed) * 1000:.2f} ms'")
            elapsed = []
        if len(shared_vector) == 0:
            sleep(0.1)
            continue
        try:
            index = random.randint(0, len(shared_vector) - 1)
            start = time.time()
            item = shared_vector.get(index)
            end = time.time() - start
            elapsed.append(end)
            # Simulate processing data
            tensor = torch.from_numpy(item).unsqueeze(0)
            # Clean up
            del item
            del tensor
        except DataModifiedError:
            print("Consumer: Data modified, skipping this get.")
            continue
        except Exception as e:
            print(f"Consumer: Error during get: {e}")
            continue


if __name__ == "__main__":
    # Parameters
    name = "shared_vector_test"
    chunk_size_bytes = 512 * 512 * 4
    max_memory_size_bytes = chunk_size_bytes * 2000
    num_producers = 4
    num_consumers = 1

    # Clean up any existing shared memory
    remove_shared_memory(name)

    # Start producers
    producers = [
        Process(
            target=producer,
            args=(name, chunk_size_bytes, max_memory_size_bytes),
        )
        for _ in range(num_producers)
    ]

    # Start consumers
    consumers = [
        Process(
            target=consumer,
            args=(name, chunk_size_bytes, max_memory_size_bytes),
        )
        for _ in range(num_consumers)
    ]

    # Start all processes
    for p in producers:
        p.start()
    sleep(5)
    for c in consumers:
        c.start()

    # Wait for all processes
    for p in producers:
        p.join()
    for c in consumers:
        c.join()
