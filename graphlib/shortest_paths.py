from queue import Queue
from time import perf_counter

import networkx as nx

from graph import Graph
from graphlib.io import read_metis_list
from graphlib.random_graph import generate_random_graph_faster


def create_dag(G: Graph, root: int):
    """
    Creates a directed acyclic graph (DAG) from the given graph G starting from the root node.
    """
    Q = Queue()
    explored = [False] * G.get_num_nodes()
    parent = [None] * G.get_num_nodes()
    children = [set()] * G.get_num_nodes()
    distance = [float('inf')] * G.get_num_nodes()
    Q.put(root)
    distance[root] = 0
    explored[root] = True

    while not Q.empty():
        u = Q.get()
        children[u] = set(G.neighbors(u))
        for v in G.neighbors(u):
            if not explored[v]:
                explored[v] = True
                parent[v] = u
                distance[v] = distance[u] + 1
                Q.put(v)

    return parent, children, distance

def single_source_shortest_path(G: Graph, root: int):
    """
    Computes the shortest path from a source node to all other nodes in the graph.
    """
    Q = Queue()
    explored = [False] * G.get_num_nodes()
    distance = [float('inf')] * G.get_num_nodes()
    Q.put(root)
    distance[root] = 0

    while not Q.empty():
        u = Q.get()
        for v in G.neighbors(u):
            if not explored[v]:
                explored[v] = True
                distance[v] = distance[u] + 1
                Q.put(v)

    return distance

def all_pairs_shortest_path(G: Graph):
    """
    Computes the shortest paths between all pairs of nodes in the graph.
    """
    distance_matrix = [[float('inf')] * G.get_num_nodes() for _ in range(G.get_num_nodes())]

    for u in range(G.get_num_nodes()):
        Q = Queue()
        explored = [False] * G.get_num_nodes()

        Q.put(u)
        distance_matrix[u][u] = 0
        explored[u] = True

        while not Q.empty():
            v = Q.get()
            for w in G.neighbors(v):
                if not explored[w]:
                    explored[w] = True
                    distance_matrix[w][v] = distance_matrix[u][v] + 1
                    distance_matrix[v][w] = distance_matrix[u][v] + 1
                    Q.put(w)

    return distance_matrix

def shortest_s_t_path(G: Graph, s: int, t: int):
    """
    Computes the shortest path from node s to node t in the graph.
    """
    parent, _, distance = create_dag(G, s)
    # check if t is reachable from s
    if distance[t] == float('inf'):
        return None, float('inf')

    # reconstruct the path that leads to t from s
    path = []
    current = t
    while current is not None:
        path.append(current)
        current = parent[current]
    path.reverse()
    return path, distance[t]

def connected_components(G: Graph):
    """
    Computes the connected components of the graph.
    """
    explored = [False] * G.get_num_nodes()
    components = []

    for u in range(G.get_num_nodes()):
        if not explored[u]:
            component = []
            Q = Queue()
            Q.put(u)
            explored[u] = True

            while not Q.empty():
                v = Q.get()
                component.append(v)
                for w in G.neighbors(v):
                    if not explored[w]:
                        explored[w] = True
                        Q.put(w)

            components.append(component)

    return components

def _example():
    a1 = perf_counter()
    g = read_metis_list("../exampleGraphs/edge_01.graph")

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
    s, t = 1, 2
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
    n = 2000
    rho = 0.5
    # generate a random graph with n nodes and density rho
    A, m = generate_random_graph_faster(n, 0.5, loops=False, directed=False)

    G = nx.Graph(A)
    path = f"../exampleGraphs/random_graph_{n}_{rho}.txt"
    nx.write_adjlist(G, path)
    print("density of graph", nx.density(G))

    a1 = perf_counter()
    g = read_metis_list(path)
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
    # append this to a csv file called benchmark.csv
    with open("benchmark.csv", "a") as f:
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