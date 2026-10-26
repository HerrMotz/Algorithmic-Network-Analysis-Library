"""Benchmarks of the Python prototype; run them from the repository root with

    python -m benchmarks.shortest_paths_benchmark
"""
import random
import tempfile
from pathlib import Path
from time import perf_counter
import networkx as nx
from pygraphlib.graph_io import read_edge_list
from pygraphlib.random_graph import generate_random_graph_faster
from pygraphlib.shortest_paths import (
    create_dag,
    single_source_shortest_path,
    all_pairs_shortest_path,
    shortest_s_t_path,
    connected_components,
)

HERE = Path(__file__).resolve().parent


def _example():
    a1 = perf_counter()
    g = read_edge_list(HERE.parent / "exampleGraphs" / "edge_01.graph")

    print(g.admat)

    a2 = perf_counter()

    b1 = perf_counter()
    print("\nCreate DAG (parents, children, distances)")
    p, c, d = create_dag(g, 1)
    print(p, c, d)
    b2 = perf_counter()

    c1 = perf_counter()
    print("\nSingle source shortest path (distances)")
    d = single_source_shortest_path(g, 1)
    print(d)
    c2 = perf_counter()

    d1 = perf_counter()
    print("\nAll pairs shortest path")
    distance_matrix = all_pairs_shortest_path(g)
    print(distance_matrix)
    d2 = perf_counter()

    e1 = perf_counter()
    print("\nShortest s-t path")
    s, t = 0, 4
    p, d = shortest_s_t_path(g, s, t)
    print(f"Path: {p}, Distance: {d}")
    e2 = perf_counter()

    f1 = perf_counter()
    print("\nConnected components")
    components = connected_components(g)
    print(components)
    f2 = perf_counter()

    print("\nFinish")

    # report all timings
    print(f"Read graph: {a2 - a1:.4f} seconds")
    print(f"Create DAG: {b2 - b1:.4f} seconds")
    print(f"Single source shortest path: {c2 - c1:.4f} seconds")
    print(f"All pairs shortest path: {d2 - d1:.4f} seconds")
    print(f"Shortest s-t path: {e2 - e1:.4f} seconds")
    print(f"Connected components: {f2 - f1:.4f} seconds")
    print(f"Total time: {f2 - a1:.4f} seconds")

def _benchmark():
    random.seed(42)

    # ask for a number of nodes and density
    n = int(input("Enter the number of nodes: "))
    rho = float(input("Enter the density (0-1): "))

    # check if the density is between 0 and 1
    if rho < 0 or rho > 1:
        print("Density must be between 0 and 1.")
        return

    # n = 500
    # rho = 0.5

    # generate a random graph with n nodes and density rho
    A, m = generate_random_graph_faster(n, rho, loops=False, directed=False)

    G = nx.Graph(A)
    path = Path(tempfile.mkdtemp(prefix="pygraphlib-")) / f"random_graph_{n}_{rho}.edgelist"
    nx.write_edgelist(G, path, data=False)
    print("density of graph", nx.density(G))

    a1 = perf_counter()
    g = read_edge_list(path)
    a2 = perf_counter()

    c1 = perf_counter()
    print("\nSingle source shortest path (distances)")
    d = single_source_shortest_path(g, 1)
    c2 = perf_counter()

    d1 = perf_counter()
    print("\nAll pairs shortest path")
    all_pairs = all_pairs_shortest_path(g)
    d2 = perf_counter()

    # report all timings
    print(f"Read graph: {a2 - a1:.4f} seconds")
    print(f"Single source shortest path: {c2 - c1:.4f} seconds")
    print(f"All pairs shortest path: {d2 - d1:.4f} seconds")
    print(f"Total time: {d2 - a1:.4f} seconds")

    print(f"{n}, {rho}, {m}, {a2-a1:.4f}, {c2-c1:.4f}, {d2-d1:.4f}, {d2-a1:.4f}")
    # append this to benchmark.csv next to this script
    with open(HERE / "benchmark.csv", "a") as f:
        f.write(f"{n}, {rho}, {m}, {a2-a1:.4f}, {c2-c1:.4f}, {d2-d1:.4f}, {d2-a1:.4f}\n")

if __name__ == "__main__":
    # ask the user whether he wants to run the benchmark or example
    choice = input("Do you want to run the example (e) or the benchmark (b)? ").strip().lower()
    if choice == 'e':
        _example()
    elif choice == 'b':
        _benchmark()
    else:
        print("Invalid choice. Please enter 'e' for example or 'b' for benchmark.")