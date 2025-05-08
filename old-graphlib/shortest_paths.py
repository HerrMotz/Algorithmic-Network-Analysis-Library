from queue import Queue
from graph import Graph

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
    explored[root] = True
    distance = [float('inf')] * G.get_num_nodes()
    distance[root] = 0
    Q.put(root)

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
        explored[u] = True
        distance = single_source_shortest_path(G, u)
        distance_matrix[u] = distance

    return distance_matrix

def shortest_s_t_path(G: Graph, s: int, t: int):
    """
    Computes the shortest path from node s to node t in the graph.
    """
    parent, _, distance = create_dag(G, s)
    # check if t is reachable from s
    if distance[t] == float('inf') or type(distance[t]) == float:
        return None, float('inf')

    # reconstruct the path that leads to t from s
    path = [None] * (int(distance[t]) + 1)
    current = t
    while current is not None:
        path[int(distance[current])] = current
        current = parent[current]
    return path, distance[t]

def connected_components(G: Graph):
    """
    Computes the connected components of the graph.
    """
    explored = [False] * G.get_num_nodes()
    components = set()

    for u in range(G.get_num_nodes()):
        if not explored[u]:
            component = set()
            Q = Queue()
            Q.put(u)
            explored[u] = True

            while not Q.empty():
                v = Q.get()
                component.add(v)
                for w in G.neighbors(v):
                    if not explored[w]:
                        explored[w] = True
                        Q.put(w)

            components.add(frozenset(component))

    return components