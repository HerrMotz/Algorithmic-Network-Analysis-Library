import osmnx as ox
import networkx as nx
import math

#this works, but there is no real sense in analyzing unweighted road networks (the original networks were weighted).
def save_graph_metis(G, filename):
    mapping = {node: idx + 1 for idx, node in enumerate(G.nodes())}
    G = nx.relabel_nodes(G, mapping)

    #remove isolated nodes
    G.remove_nodes_from(list(nx.isolates(G)))

    with open(filename, "w") as f:
        f.write(f"{G.number_of_nodes()} {G.number_of_edges()}\n")
        for node in sorted(G.nodes()):
            neighbors = sorted([neighbor for neighbor in G.neighbors(node)])
            f.write(" ".join(str(neighbor) for neighbor in neighbors) + "\n")

#better version, unfold edges with length > 1
def save_graph_metis_unfolded(G, filename, segment_length=50):
    # remove isolated nodes
    G.remove_nodes_from(list(nx.isolates(G)))

    # relabel nodes with contiguous integers starting from 1
    original_nodes = list(G.nodes())
    mapping = {node: idx + 1 for idx, node in enumerate(original_nodes)}
    next_node_id = len(original_nodes) + 1

    unfolded_edges = []

    for u, v, data in G.edges(data=True):
        weight = data.get('length', 1)  # use 'length' as weight, fallback to 1
        steps = max(1, math.ceil(weight / segment_length))  # how many edges to unfold

        if steps == 1:
            unfolded_edges.append((mapping[u], mapping[v]))
        else:
            prev = mapping[u]
            for _ in range(steps - 1):
                curr = next_node_id
                next_node_id += 1
                unfolded_edges.append((prev, curr))
                prev = curr
            unfolded_edges.append((prev, mapping[v]))

    # build new graph from unfolded edges
    unfolded_G = nx.Graph()
    for u, v in unfolded_edges:
        unfolded_G.add_edge(u, v)

    # save in METIS format
    with open(filename, "w") as f:
        f.write(f"{unfolded_G.number_of_nodes()} {unfolded_G.number_of_edges()}\n")
        for node in range(1, unfolded_G.number_of_nodes() + 1):
            neighbors = sorted(unfolded_G.neighbors(node))
            f.write(" ".join(str(neighbor) for neighbor in neighbors) + "\n")


#define cities
cities = ["Jena, Germany", "Erfurt, Germany", "Apolda, Germany", "Weimar, Germany"]

for city in cities:
    G = ox.graph_from_place(city, network_type="drive")
    G = G.to_undirected()
    city_name = city.split(",")[0].lower()
    save_graph_metis_unfolded(G, f"{city_name}_road_unfolded_50.metis")
    print(f"Saved {city_name}_road_unfolded_50.metis")

#It's meters...
#Apolda: shortest 2.73, longest 2032.54



