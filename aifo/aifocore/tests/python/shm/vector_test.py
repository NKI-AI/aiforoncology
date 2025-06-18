from multiprocessing import Manager, Process

import numpy as np
import pytest
from aifocore.shm.vector import SharedVector, remove_shared_memory

MAX_MEMORY_SIZE = 1024 * 1024 * 1240


def test_reference_count_and_pointer():
    memory_name = "shared_memory_reference_count"
    remove_shared_memory(memory_name)
    manager = SharedVector(memory_name, max_memory_size=1024 * 1024 * 1240, chunk_size=10 * 10 * 10)
    arr = np.random.random((2, 2)).astype(np.float32)
    manager.append(arr)
    manager.append(arr * 2)

    x = manager.get(0)
    assert not x.flags.owndata

    pointer_id = manager.get_chunk_pointer(0)
    assert pointer_id == x.__array_interface__["data"][0]

    assert (x == arr).all()
    assert manager.get_chunk_ref_count(0) == 1

    y = manager.get(0)
    assert manager.get_chunk_ref_count(0) == 2

    del x
    assert manager.get_chunk_ref_count(0) == 1
    del y
    assert manager.get_chunk_ref_count(0) == 0

    manager.get(1)
    assert manager.get_chunk_ref_count(1) == 0
    remove_shared_memory(memory_name)


def producer(memory_name, process_index, num_items):
    manager = SharedVector(memory_name, max_memory_size=MAX_MEMORY_SIZE, chunk_size=10 * 10 * 10)
    for i in range(num_items):
        value = float(process_index * num_items + i)
        manager.append(np.array([value], dtype=np.float32))


def consumer(memory_name, results):
    manager = SharedVector(memory_name, max_memory_size=MAX_MEMORY_SIZE, chunk_size=10 * 10 * 10)
    idx = 0
    while True:
        try:
            value = manager.get(idx)[0]
            results.append(value)
            idx += 1
        except IndexError:
            if len(results) == len(manager):
                break


def test_multi_producer_consumer():
    memory_name = "shared_memory_multi_producer_consumer"
    remove_shared_memory(memory_name)
    manager = SharedVector(memory_name, max_memory_size=MAX_MEMORY_SIZE, chunk_size=10 * 10 * 10)
    assert manager.ref_count == 1

    num_producers = 3
    items_per_producer = 5
    total_items = num_producers * items_per_producer

    mp_manager = Manager()
    results = mp_manager.list()

    producers = [Process(target=producer, args=(memory_name, i, items_per_producer)) for i in range(num_producers)]
    consumer_process = Process(target=consumer, args=(memory_name, results))

    for p in producers:
        p.start()

    consumer_process.start()

    for p in producers:
        p.join()

    consumer_process.join()

    assert len(results) == total_items, f"Expected {total_items} items, but got {len(results)}"
    assert sorted(results) == list(range(total_items)), "Results do not match expected values"
    remove_shared_memory(memory_name)


def test_replace():
    memory_name = "shared_memory_replace"
    remove_shared_memory(memory_name)
    manager = SharedVector(memory_name, max_memory_size=MAX_MEMORY_SIZE, chunk_size=10 * 10 * 10)
    arr = np.random.random((2, 2)).astype(np.float32)
    arr2 = np.random.random((3, 3)).astype(np.float32)
    manager.append(arr)
    manager.append(arr * 2)
    manager.append(arr * 3)

    x = manager.get(0)
    assert (x == arr).all()
    assert manager.get_chunk_ref_count(0) == 1

    x = manager.get(1)
    assert (x == arr * 2).all()
    del x
    assert manager.get_chunk_ref_count(1) == 0
    manager.replace(1, arr2)
    x = manager.get(1)
    assert (x == arr2).all()
    del x
    remove_shared_memory(memory_name)


def test_non_contiguous_array():
    memory_name = "shared_memory_non_contiguous"
    remove_shared_memory(memory_name)
    manager = SharedVector(memory_name, max_memory_size=MAX_MEMORY_SIZE, chunk_size=10 * 10 * 10)
    arr = np.random.random((2, 2)).astype(np.float32)

    # Let's make arr non-contiguous
    arr = arr.T
    assert not arr.flags.contiguous
    with pytest.raises(RuntimeError):
        manager.append(arr)

    with pytest.raises(RuntimeError):
        manager.replace(0, arr)

    remove_shared_memory(memory_name)


def test_arbitrary_shapes():
    memory_name = "shared_memory_arbitrary_shapes"
    remove_shared_memory(memory_name)
    manager = SharedVector(memory_name, max_memory_size=MAX_MEMORY_SIZE, chunk_size=10 * 10 * 10)
    arr = np.random.random((2, 2)).astype(np.float32)
    arr2 = np.random.random((3, 3)).astype(np.float32)
    arr3 = np.random.random((4, 4)).astype(np.float32)
    manager.append(arr)
    manager.append(arr2)
    manager.append(arr3)

    x = manager.get(0)
    assert (x == arr).all()
    assert manager.get_chunk_ref_count(0) == 1
    assert manager.get_chunk_ref_count(1) == 0
    assert manager.get_chunk_ref_count(2) == 0

    y = manager.get(1)
    assert (y == arr2).all()
    assert manager.get_chunk_ref_count(0) == 1
    assert manager.get_chunk_ref_count(1) == 1
    assert manager.get_chunk_ref_count(2) == 0

    z = manager.get(2)
    assert (z == arr3).all()
    assert manager.get_chunk_ref_count(0) == 1
    assert manager.get_chunk_ref_count(1) == 1
    assert manager.get_chunk_ref_count(2) == 1

    del x
    del y
    del z

    assert manager.get_chunk_ref_count(0) == 0
    assert manager.get_chunk_ref_count(1) == 0
    assert manager.get_chunk_ref_count(2) == 0

    del manager
    remove_shared_memory(memory_name)


def test_manager_ref_count():
    memory_name = "shared_memory_manager_ref_count"
    remove_shared_memory(memory_name)
    manager = SharedVector(memory_name, max_memory_size=MAX_MEMORY_SIZE, chunk_size=10 * 10 * 10)

    assert manager.ref_count == 1

    managers = []
    for idx in range(4):
        new_manager = SharedVector(memory_name, max_memory_size=MAX_MEMORY_SIZE, chunk_size=10 * 10 * 10)
        assert new_manager.ref_count == idx + 2
        managers.append(new_manager)

    del new_manager

    for idx in range(4):
        del managers[-1]
        expected_ref_count = 5 - (idx + 1)
        assert manager.ref_count == expected_ref_count, f"Expected {expected_ref_count}, but got {manager.ref_count}"

    assert managers == []

    assert manager.ref_count == 1
    remove_shared_memory(memory_name)


def test_chunk_size():
    memory_name = "shared_memory_chunk_size"
    remove_shared_memory(memory_name)
    manager = SharedVector(memory_name, max_memory_size=MAX_MEMORY_SIZE, chunk_size=10 * 10 * 10)

    arr = np.random.random((2, 2)).astype(np.float32)
    manager.append(arr)

    assert manager.get_chunk_shape(0) == (2, 2)

    manager.append(arr.reshape(4))
    assert manager.get_chunk_shape(1) == (4,)
    remove_shared_memory(memory_name)
