#improved versions of diameter() and all_pairs_shortest_paths

from queue import Queue

#simple diameter algorithm - calculate the entire all-pairs shortest path matrix and look for a biggest entry.


test_adj_list_0 = {
    0: [1],
    1: [0]
}

test_adj_list_1 = {
    0: [1, 2],
    1: [0],
    2: [0],
    3: []
}

test_adj_list_2 = {
    0: [3, 5],
    1: [2, 4],
    2: [1, 4],
    3: [0, 4],
    4: [1, 2, 3],
    5: [0]
}


def get_cut_vertices(adj_list):
    """
    returns all cut-vertices of a graph using tarjan's algorithm
    """
    time = [0]
    disc = {} #time the node was discovered in the DFS
    low = {} #the lowest possible discovery time of a node or his descendants possibly using back nodes
    parent = {}
    visited = set()
    cut_vertices = set()

    def dfs(u):
        visited.add(u)
        disc[u] = low[u] = time[0]
        time[0] += 1
        num_children = 0

        for v in adj_list.get(u, []):
            if v not in visited:
                parent[v] = u
                num_children += 1
                dfs(v)

                low[u] = min(low[u], low[v])

                if parent.get(u) is None and num_children > 1: #u is the root and has more than 1 child
                    cut_vertices.add(u)
                if parent.get(u) is not None and low[v] >= disc[u]: #u is not the root and v and his descendants are only connected to the root through u
                    cut_vertices.add(u)
            elif v != parent.get(u):
                low[u] = min(low[u], disc[v])

    for node in adj_list:
        if node not in visited:
            dfs(node)

    return cut_vertices

#print(get_cut_vertices(test_adj_list_2)) #seems to work

def get_neighborhood_subsets(adj_list):
    """
    returns a set of vertices u such that there exists another vertex v in the graph whose (open or closed) neighborhood is a subset of the (open or closed) neighborhood of u.
    """
    result = set()
    open_neighbor_sets = {u: set(neigh) for u, neigh in adj_list.items()} #dictionary comprehension (using sets instead of lists)
    closed_neighbor_sets = {u: open_neighbor_sets[u] | {u} for u in adj_list} #add vertex itself for closed neighborhood

    for u in adj_list:
        candidates = set()
        for v in open_neighbor_sets[u]:
            candidates.add(v)
            candidates.update(open_neighbor_sets.get(v, []))
        candidates.discard(u)

        for v in candidates:
            if v in result:
                continue #to not add a AND b to the result if N(a) = N(b) (or N[a] = N[b])
            if (open_neighbor_sets[v].issubset(open_neighbor_sets[u]) or closed_neighbor_sets[v].issubset(closed_neighbor_sets[u])):
                result.add(u)
                break

    return result

#print(get_neighborhood_subsets(test_adj_list_2)) #works aswell


def bfs_max_distance(graph, root):
    """
    returns the maximum distance from a given vertex to any other vertex in the graph using BFS
    """
    Q = Queue()
    explored = set([root])
    Q.put((root, 0))
    max_dist = 0

    while not Q.empty():
        node, dist = Q.get()
        max_dist = dist
        for neighbor in graph.get(node, []):
            if neighbor not in explored:
                explored.add(neighbor)
                Q.put((neighbor, dist + 1))

    if len(explored) == len(graph):
        return max_dist
    else:
        return float('inf')


def improved_diameter(adj_list):
    """
    returns the diameter of a given graph using bfs
    does not start bfs from cut-vertices or multiple vertices with the same neighborhood
    the max. distances from the nodes each give upper (d*2) and lower (d) bounds for the diameter
    """
    skip_nodes = get_cut_vertices(adj_list)
    skip_nodes.update(get_neighborhood_subsets(adj_list)) #skip_nodes is now a set of vertices where we do not need to start a BFS from
    all_nodes = set(adj_list.keys())
    start_nodes = all_nodes - skip_nodes
    lower_bound = 0
    upper_bound = float('inf')

    for node in start_nodes:
        max_dist = bfs_max_distance(adj_list, node)
        if max_dist == float('inf'):
            return float('inf')
        if max_dist > lower_bound:
            lower_bound = max_dist
        if max_dist * 2 < upper_bound:
            upper_bound = max_dist * 2
        if upper_bound == lower_bound:
            return lower_bound

    return lower_bound


#print(improved_diameter(test_adj_list_2)) #works


def single_source_shortest_paths(graph, root):
    """
    Computes the shortest path from a source node to all other nodes in the graph.
    """
    Q = Queue()
    explored = set([root])
    distance = [float('inf')] * len(graph)
    distance[root] = 0
    Q.put(root)

    while not Q.empty():
        node = Q.get()
        for neighbor in graph.get(node, []):
            if neighbor not in explored:
                explored.add(neighbor)
                distance[neighbor] = distance[node] + 1
                Q.put(neighbor)

    return distance

def get_same_neighborhood_nodes(node, open_neighborhood, closed_neighborhood): #works
    """
    returns a set of vertices with the same (open or closed) neighborhood as a given vertex
    """
    same_neighborhood = set()

    closed_candidates = open_neighborhood[node]
    for candidate in closed_candidates:
        if closed_neighborhood[candidate] == closed_neighborhood[node]:
            same_neighborhood.add(candidate)

    all_candidates = set(closed_candidates)
    for u in closed_candidates:
        all_candidates.update(open_neighborhood.get(u, [])) #added neighbors of neighbors
    open_candidates = all_candidates - closed_neighborhood[node] #delete all direct neighbors and the node itself from the candidates
    for candidate in open_candidates:
        if open_neighborhood[candidate] == open_neighborhood[node]:
            same_neighborhood.add(candidate)

    return same_neighborhood


def improved_all_pairs_shortest_paths(adj_list):
    """
    returns the distance matrix of a given graph
    uses BFS to compute the distances from one vertex to all other vertices
    copies the distances for vertices with the same (open or closed) neighborhood
    for degrees lower than a threshold t computes the distances using the distances of its neighbors
    """
    distance_matrix = [[float('inf')] * len(adj_list) for _ in range(len(adj_list))]
    distances_known = [False] * len(adj_list) #only works if the nodes are 0,1,2 and so on

    open_neighbor_sets = {u: set(neigh) for u, neigh in adj_list.items()} #dictionary comprehension (using sets instead of lists)
    closed_neighbor_sets = {u: open_neighbor_sets[u] | {u} for u in adj_list} #add vertex itself for closed neighborhood
    degree_threshold = int(sum(len(x) for x in adj_list.values()) / len(adj_list)) #(m/n)

    for node in set(adj_list.keys()):
        if distances_known[node]:
            continue

        degree = len(open_neighbor_sets[node])
        if degree == 0:
            distances_known[node] = True
            distance_matrix[node][node] = 0 #in Zeile 212 einfügen
            continue

        #check if distances from all neighbors are already known
        if all(distances_known[neighbor] for neighbor in open_neighbor_sets[node]) and degree < degree_threshold:
            distances = [float('inf')] * len(adj_list)

            for u in range(len(adj_list)):
                if u == node:
                    distances[u] = 0
                elif u in open_neighbor_sets[node]:
                    distances[u] = 1
                else:
                    distances[u] = min(distance_matrix[neighbor][u] for neighbor in open_neighbor_sets[node]) + 1

        #compute distances from current vertex
        else:
            distances = single_source_shortest_paths(adj_list, node)

        distance_matrix[node] = distances
        distances_known[node] = True

        #check for vertices with the same (open or closed) neighborhood and copy their distances
        for other_node in get_same_neighborhood_nodes(node, open_neighbor_sets, closed_neighbor_sets):
            distance_matrix[other_node] = distances.copy() #copy distances
            distance_matrix[other_node][other_node] = 0
            if other_node in open_neighbor_sets[node]:
                distance_matrix[other_node][node] = 1 #direct neighbor
            else:
                distance_matrix[other_node][node] = 2 #no direct neighbor
            distances_known[other_node] = True

    return distance_matrix

print(improved_all_pairs_shortest_paths(test_adj_list_1))
