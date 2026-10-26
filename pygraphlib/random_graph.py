import random
from math import ceil, floor

import numpy as np


def generate_random_graph_faster(n: int, rho: float, loops: bool = False, directed: bool = False,
                                 seed: int = 42) -> tuple[np.ndarray, int]:
    """
    Create the adjacency matrix of a random graph of desired size and density.
    :param n: the number of nodes in the graph
    :param rho: the graph's density
    :param loops: whether the occurrence loops are permitted (default: False)
    :param directed: whether the network is directed (default: False)
    :param seed: seed for the random number generator
    :return: the adjacency matrix of the graph and its number of edges
    """
    rng = random.Random(seed)

    # calculate the necessary number of edges to get the desired density
    m = (rho * n * (n - 1))
    m /= 2 if not directed else 1
    m += n if loops else 0
    m = ceil(m) if rng.randint(0, 1) else floor(m)

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

    selected_edges = rng.sample(possible_edges, k=m)
    print(f"Number of edges: {len(selected_edges)}")

    for e in selected_edges:
        x, y = e
        A[x, y] = 1
        if not directed:
            A[y, x] = 1

    return A, len(selected_edges)

