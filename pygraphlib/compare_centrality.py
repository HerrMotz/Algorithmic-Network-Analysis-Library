import json
from pathlib import Path

import networkx as nx

def main():
    path = Path(__file__).resolve().parent.parent / "exampleGraphs" / "graph_for_test.graph"

    # Graph einlesen
    G = nx.Graph()
    with open(path) as f:
        for line in f:
            if line.strip() and not line.startswith("#"):
                u, v = line.split()
                G.add_edge(u, v)

    n = G.number_of_nodes()

    raw_deg = dict(G.degree())

    # beide Varianten der Closeness:
    closeness_raw  = {}
    closeness_norm = {}
    for v in G.nodes():
        # Kürzeste Distanzen von v zu allen erreichbaren Knoten
        lengths = nx.single_source_shortest_path_length(G, v)
        s = sum(lengths.values())
        if s > 0:
            closeness_raw[v]  = 1.0 / s
            # (|R(v)|−1) = len(lengths)−1; bei zusammenhängendem Graphen = n−1
            closeness_norm[v] = float(len(lengths) - 1) / s
        else:
            closeness_raw[v]  = 0.0
            closeness_norm[v] = 0.0

    # Top-5 sortieren
    raw_top5 = sorted(raw_deg.items(), key=lambda x: -x[1])[:5]
    cr_top5  = sorted(closeness_raw.items(),  key=lambda x: -x[1])[:5]
    cn_top5  = sorted(closeness_norm.items(), key=lambda x: -x[1])[:5]

    output = {
        "raw_degree_top5":         [{"v": v, "val": val} for v, val in raw_top5],
        "closeness_raw_top5":      [{"v": v, "val": val} for v, val in cr_top5],
        "closeness_norm_top5":     [{"v": v, "val": val} for v, val in cn_top5],
    }

    print(json.dumps(output, indent=2))

if __name__ == "__main__":
    main()
