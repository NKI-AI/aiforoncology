# Copyright 2025 AI for Oncology Research Group. All Rights Reserved.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#     http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
def ensure_tuple(x: int | tuple, n: int) -> tuple:
    """Convert an integer or a tuple to an n-tuple.

    Parameters
    ----------
    x : int or tuple
        Input value which can be an integer or a tuple of n integers.
    n : int
        The length of the tuple to return.
    Returns
    -------
    tuple
        A tuple of n integers.
    """
    if isinstance(x, tuple):
        if len(x) != n:
            raise ValueError(f"Expected a tuple of length {n}, got {len(x)}")
        if not all(isinstance(i, int) for i in x):
            raise ValueError("All elements in the tuple must be integers")
        return tuple(x)

    if not isinstance(x, int):
        raise TypeError(f"Expected int, got {type(x)}")
    return (x,) * n
