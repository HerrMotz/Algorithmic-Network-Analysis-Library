from queue import Queue
from graph import Graph
from graphlib.io import read_edge_list


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
    all_pairs = []
    for u in range(G.get_num_nodes()):
        Q = Queue()
        explored = [False] * G.get_num_nodes()
        parent = [None] * G.get_num_nodes()
        distance = [float('inf')] * G.get_num_nodes()

        Q.put(u)
        distance[u] = 0
        parent[u] = None
        explored[u] = True

        while not Q.empty():
            v = Q.get()
            for w in G.neighbors(v):
                if not explored[w]:
                    explored[w] = True
                    parent[w] = v
                    distance[w] = distance[v] + 1
                    Q.put(w)

        all_pairs.append((parent, distance))
    return all_pairs

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

if __name__ == "__main__":
    g = read_edge_list("../exampleGraphs/edge_01.graph")

    print("\nCreate DAG (parents, children, distances)")
    p, c, d = create_dag(g, 1)
    print(p, c, d)

    print("\nSingle source shortest path (distances)")
    d = single_source_shortest_path(g, 1)
    print(d)

    print("\nAll pairs shortest path")
    all_pairs = all_pairs_shortest_path(g)
    for i, (p, d) in enumerate(all_pairs):
        print(f"Node {i}: Parent: {p}, Distance: {d}")

    print("\nShortest s-t path")
    s, t = 1, 2
    p, d = shortest_s_t_path(g, s, t)
    print(f"Path: {p}, Distance: {d}")

    print("\nConnected components")
    components = connected_components(g)
    print(components)

    print("\nFinish")
