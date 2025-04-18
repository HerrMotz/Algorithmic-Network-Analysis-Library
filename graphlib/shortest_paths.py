from queue import Queue
from graph import Graph
from graphlib.io import read_edge_list


def create_dag(G: Graph, root: int):
    Q = Queue()
    explored = [False] * G.num_nodes
    parent = [None] * G.num_nodes
    children = [set()] * G.num_nodes
    distance = [float('inf')] * G.num_nodes
    Q.put(root)
    distance[root] = 0

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


def single_source_shortest_paths(G: Graph, root: int):
    parent, children, distance = create_dag(G, root)
    print(parent)
    print(children)
    print(distance)
    print("Finish")


if __name__ == "__main__":
    graph = read_edge_list("../exampleGraphs/edge_01.graph")
    single_source_shortest_paths(graph, 1)
