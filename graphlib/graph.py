class Graph:
    def __init__(self):
        self.admat = []
        self.num_nodes = 0

    # Utility
    def check_node_exists(self, node):
        return node < self.num_nodes

    # Add Functions
    def add_node(self, node):
        if node < self.num_nodes:
            return

        new_size = node
        growth = new_size - self.num_nodes

        for row in self.admat:
            row.extend([0] * growth)

        for _ in range(growth):
            self.admat.append([0] * new_size)

        self.num_nodes = new_size

    def add_edge(self, u, v):
        if u >= self.num_nodes:
            self.add_node(u)
        if v >= self.num_nodes:
            self.add_node(v)

        u -= 1
        v -= 1

        self.admat[u][v] = 1
        self.admat[v][u] = 1

    # Remove Functions
    def remove_edge(self, u, v):
        if u > self.num_nodes or v >= self.num_nodes:
            raise IndexError("Knoten existiert nicht.")

        u -= 1
        v -= 1

        self.admat[u][v] = 0
        self.admat[v][u] = 0

    def remove_node(self, node):
        if node > self.num_nodes:
            raise IndexError("Knoten existiert nicht.")

        node -= 1

        self.admat.pop(node)
        for row in self.admat:
            row.pop(node)

        self.num_nodes -= 1

    # Get Infos about node
    def neighbors(self, node):
        if node > self.num_nodes:
            raise IndexError("Knoten existiert nicht.")

        node -= 1
        return [i + 1 for i, val in enumerate(self.admat[node]) if val == 1]

    def is_neighbor(self, u, v):
        if u > self.num_nodes or v > self.num_nodes:
            raise IndexError("Knoten existiert nicht.")

        u -= 1
        v -= 1
        return self.admat[u][v] == 1

    # Get node degree
    def get_node_degree(self, node):
        if node > self.num_nodes:
            raise IndexError("Knoten existiert nicht.")

        return len(self.neighbors(node))

    # Other functions
    def get_num_nodes(self):
        return self.num_nodes
    
    def get_num_edges(self):
        num_edges = 0
        for i in range(self.num_nodes):
            for j in range(i + 1, self.num_nodes):
                if self.admat[i][j] == 1:
                    num_edges += 1
        return num_edges

    def print_graph(self):
        print("Adjazenzmatrix:")
        for row in self.admat:
            print(" ".join(str(val) for val in row))
        print()