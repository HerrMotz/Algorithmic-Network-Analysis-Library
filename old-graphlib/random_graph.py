import random
from math import ceil, floor
from random import randint, sample
from typing import Any

import numpy as np
from numpy import ndarray, dtype, floating
from numpy._typing import _64Bit


def generate_random_graph_faster(n: int, rho: float, loops: bool = False, directed: bool = False, seed: int = 42) -> tuple[
    ndarray[Any, dtype[floating[_64Bit]]], list[tuple[int, int]]]:
    """
    Create the adjacency matrix of a random graph of desired size and density.
    :param n: the number of nodes in the graph
    :param rho: the graph's density
    :param loops: whether the occurrence loops are permitted (default: False)
    :param directed: whether the network is directed (default: False)
    :return: the adjacency matrix of the graph
    """

    # calculate the necessary number of edges to get the desired density
    m = (rho * n * (n - 1))
    m /= 2 if not directed else 1
    m += n if loops else 0
    m = ceil(m) if randint(0, 1) else floor(m)

    A = np.zeros((n, n))

    if directed:
        if loops:
            possible_edges = [(x, y) for x in range(n) for y in range(n)]
        else:
            possible_edges = [(x, y) for x in range(n) for y in range(n) if x != y]
    else:
        if loops:
            possible_edges = [(x, y) for x in range(n) for y in range(n) if x <= y]
        else:
            possible_edges = [(x, y) for x in range(n) for y in range(n) if x < y]

    random.seed(seed)
    selected_edges = sample(possible_edges, k=m)
    print(f"Number of edges: {len(selected_edges)}")

    for e in selected_edges:
        x, y = e
        A[x, y] = 1
        if not directed:
            A[y, x] = 1

    return A, len(selected_edges)

