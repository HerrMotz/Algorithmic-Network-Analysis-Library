class Graph:
    def __init__(self):
        self.admat = []
        self.indexlist = []
        self.node_to_index = {}

    # Utility
    def check_node_exists(self, node):
        return node in self.node_to_index

    # Add Functions
    def add_node(self, node):
        if node in self.node_to_index:
            return

        self.indexlist.append(node)
        self.node_to_index[node] = len(self.indexlist) - 1
        new_size = len(self.indexlist)

        for row in self.admat:
            row.append(0)

        self.admat.append([0] * new_size)


    def add_edge(self, u, v, directed=False):
        # if node is not known jet, add it
        if u not in self.node_to_index:
            self.add_node(u)
        if v not in self.node_to_index:
            self.add_node(v)

        # The position of the node in the indexlist is the index in the adjacency matrix
        id_u = self.node_to_index[u]
        id_v = self.node_to_index[v]

        # Add the edge
        self.admat[id_u][id_v] = 1
        if not directed:
            self.admat[id_v][id_u] = 1


    # Remove Functions
    def remove_edge(self, u, v, directed=False):
        if u not in self.node_to_index or v not in self.node_to_index:
            raise IndexError("Node does not exist.")

        # get the index
        id_u = self.node_to_index[u]
        id_v = self.node_to_index[v]

        # remove the edge
        self.admat[id_u][id_v] = 0
        if directed is False:
            self.admat[id_v][id_u] = 0

    def remove_node(self, node):
        if node not in self.indexlist:
            raise IndexError("Knoten existiert nicht.")

        # get the index
        id_node = self.node_to_index(node)

        self.admat.pop(id_node)
        for row in self.admat:
            row.pop(id_node)

        self.indexlist.pop(id_node)
        # pop the node from the node_to_index
        self.node_to_index.pop(node)
        # self.node_to_index = {n: i for i, n in enumerate(self.indexlist)}

    # Get Infos about node
    def neighbors(self, node):
        if node not in self.node_to_index:
            raise IndexError("Knoten existiert nicht.")
    
        id_node = self.node_to_index[node]
        return [self.indexlist[i] for i, val in enumerate(self.admat[id_node]) if val == 1]
    

    def is_neighbor(self, u, v):
        if u not in self.node_to_index or v not in self.node_to_index:
            raise IndexError("Knoten existiert nicht.")

        # get the index
        id_u = self.node_to_index[u]
        id_v = self.node_to_index[v]

        return self.admat[id_u][id_v] == 1

    # Get node degree
    def get_node_degree(self, node):
        if node not in self.node_to_index:
            raise IndexError("Knoten existiert nicht.")
    
        id_node = self.node_to_index[node]
        return sum(self.admat[id_node])

    # Other functions
    def get_num_nodes(self):
        return len(self.indexlist)
    
    def get_num_edges(self, directed=False):
        num_edges = 0
        num_nodes = len(self.indexlist)

        for i in range(num_nodes):
            if directed:
                num_edges += sum(self.admat[i])
            else:
                for j in range(i, num_nodes):
                    if self.admat[i][j] == 1:
                        num_edges += 1
        return num_edges

    def print_graph(self):
        print("Adjazenzmatrix:")
        for row in self.admat:
            print(" ".join(str(val) for val in row))
        print()