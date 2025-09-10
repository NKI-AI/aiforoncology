from logging import getLogger
from pathlib import Path
from typing import List

logger = getLogger(__name__)


def find_files_by_extension(directory: Path | str, extensions: str | List[str]) -> List[Path]:
    if isinstance(directory, str):
        directory = Path(directory)
    if isinstance(extensions, str):
        extensions = [extensions]
    if not directory.is_dir():
        raise ValueError(f"{directory} is not a valid directory")

    files = []
    for ext in extensions:
        files.extend(directory.rglob(f"*{ext}"))

    logger.info(f"Found {len(files)} files with extensions {extensions} in {directory}")
    return files
