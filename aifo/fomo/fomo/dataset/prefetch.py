import threading
import queue
import logging
from torch.utils.data import DataLoader

logger = logging.getLogger(__name__)


class PinnedDataPrefetcher:
    def __init__(self, data_loader: DataLoader, max_prefetch: int = 2):
        self._loader = data_loader
        self._queue = queue.Queue(max_prefetch)
        self._worker = threading.Thread(target=self._worker_func, daemon=True)
        self._worker.start()

    def _worker_func(self) -> None:
        try:
            for data in self._loader:
                pinned_data = {}
                for key, value in data.items():
                    if key == "samples_list" and isinstance(value, list):
                        # Pin each tensor in the list
                        pinned_data[key] = [tensor.pin_memory() for tensor in value]
                    else:
                        pinned_data[key] = value
                self._queue.put(pinned_data)
        except Exception as e:
            logger.error("Prefetch error: %s", e)
        finally:
            self._queue.put(None)

    def next(self):
        return self._queue.get()
