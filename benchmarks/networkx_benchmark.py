"""networkx reference timings; run from the repository root with

    python -m benchmarks.networkx_benchmark
"""
import csv
import tempfile
import time
from pathlib import Path

import networkx as nx

from pygraphlib.random_graph import generate_random_graph_faster

HERE = Path(__file__).resolve().parent


# benchmarking reading from a file, single-source shortest path and all-pairs shortest path with networkx
def networkx_benchmark(number_of_nodes, rho, filename=HERE / "benchmarks.csv"):
    graph, m = generate_random_graph_faster(number_of_nodes, rho)
    g = nx.Graph(graph) #g is a networkx graph

    path = Path(tempfile.mkdtemp(prefix="pygraphlib-")) / "random_graph.txt"
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


if __name__ == '__main__':
    networkx_benchmark(10, 0.5)
    networkx_benchmark(50, 0.5)
    networkx_benchmark(100, 0.5)
    networkx_benchmark(500, 0.1)
    networkx_benchmark(500, 0.5)
    networkx_benchmark(1000, 0.1)
    networkx_benchmark(1000, 0.5)
    networkx_benchmark(2000, 0.5)