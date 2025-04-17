from queue import Queue
from graph import Graph
from graphlib.io import read_edge_list

def create_dag(G: Graph, root: int):
    Q = Queue()
    explored = [False] * G.num_nodes
    parent = [None] * G.num_nodes
    children = [set()] * G.num_nodes
    distance = [-1] * G.num_nodes
    Q.put(root)
    distance[root] = 0

    while not Q.empty():
        u = Q.get()
        for v in G.neighbors(u):
            _v = v-1
            if not explored[_v]:
                explored[_v] = True
                parent[_v] = u
                children[_v].add(v)
                distance[_v] = distance[u] + 1
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
    single_source_shortest_paths(graph, 0)