import random
from math import ceil, floor
from random import randint, sample
from typing import Any

import numpy as np
from numpy import ndarray, dtype, floating
from numpy._typing import _64Bit

import networkx as nx
import time
import csv

#Copied from radom_graph.py (only number of edges won't be returned)
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

#benchmarking reading from a file, single-source shortest path and all-pairs shortest path with networkx
def networkx_benchmark(number_of_nodes, rho, filename = "benchmarks.csv"):
    graph, m = generate_random_graph_faster(number_of_nodes, rho)
    g = nx.Graph(graph) #g is a networkx graph

    path = f"../algorithmische-netzwerkanalyse-lab/exampleGraphs/random_graph.txt"
    nx.write_adjlist(g, path)

    #benchmarking reading the graph from a file
    start_read = time.perf_counter()
    g_read = nx.read_adjlist(path, create_using=nx.Graph())
    end_read = time.perf_counter()
    read_time = end_read - start_read

    #benchmark single-source shortest path
    start_single = time.perf_counter()
    _ = nx.single_source_shortest_path(g, 0) #this calculates the actual paths and not only their length
    end_single = time.perf_counter()
    single_time = end_single - start_single

    #benchmark all-pairs shortest path
    start_all_pairs = time.perf_counter()
    _ = list(nx.all_pairs_shortest_path(g))
    end_all_pairs = time.perf_counter()
    all_pairs_time = end_all_pairs - start_all_pairs

    total_time = read_time + single_time + all_pairs_time

    #writing results into a csv-file
    with open(filename, mode='a', newline='') as f:
        writer = csv.writer(f)
        f.seek(0, 2)
        #write header if file is empty
        if f.tell() == 0:
            writer.writerow(["n", "rho", "m", "read", "single", "all-pairs", "total"])
        writer.writerow([number_of_nodes, rho, m, read_time, single_time, all_pairs_time, total_time])

    print(f"benchmark saved for n={number_of_nodes}, rho={rho}")


networkx_benchmark(10, 0.5)
networkx_benchmark(50, 0.5)
networkx_benchmark(100, 0.5)
networkx_benchmark(500, 0.1)
networkx_benchmark(500, 0.5)
networkx_benchmark(1000, 0.1)
networkx_benchmark(1000, 0.5)
networkx_benchmark(2000, 0.5)