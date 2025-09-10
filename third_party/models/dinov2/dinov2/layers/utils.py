from omegaconf import ListConfig


def make_tuple(x: int | tuple | ListConfig, n: int) -> tuple:
    """Convert an integer or a tuple/ListConfig to an n-tuple.

    Parameters
    ----------
    x : int or tuple or ListConfig
        Input value which can be an integer or a tuple/ListConfig of n integers.
    n : int
        The length of the tuple to return.

    Returns
    -------
    tuple
        A tuple of n integers.
    """
    if isinstance(x, tuple) or isinstance(x, ListConfig):
        if len(x) != n:
            raise ValueError(f"Expected a tuple of length {n}, got {len(x)}")
        if not all(isinstance(i, int) for i in x):
            raise ValueError("All elements in the tuple must be integers")
        return tuple(x)

    if not isinstance(x, int):
        raise TypeError(f"Expected int, got {type(x)}")
    return (x,) * n


def make_2tuple(x: int | tuple[int, int] | ListConfig) -> tuple[int, int]:
    """Convert an integer or a tuple to a 2-tuple.

    Parameters
    ----------
    x : int or tuple[int, int] or ListConfig
        Input value which can be an integer or a tuple of two integers or a ListConfig with two elements.

    Returns
    -------
    tuple[int, int]
        A tuple of two integers.
    """
    return make_tuple(x, 2)


def make_3tuple(x: int | tuple[int, int, int] | ListConfig) -> tuple[int, int, int]:
    """Convert an integer or a tuple to a 3-tuple.

    Parameters
    ----------
    x : int or tuple[int, int, int] or ListConfig
        Input value which can be an integer or a tuple of three integers or a ListConfig with three elements.

    Returns
    -------
    tuple[int, int, int]
        A tuple of three integers.
    """
    return make_tuple(x, 3)
