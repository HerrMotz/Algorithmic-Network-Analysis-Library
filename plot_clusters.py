#!/usr/bin/env python3
import argparse
import os
from collections import defaultdict

import matplotlib.pyplot as plt
import networkx as nx
from matplotlib.patches import Patch
from scipy.io import mmread
from scipy.sparse import coo_matrix


# ---------- Einlesen (unverändert) -----------------------------------
def read_graph(path):
    if path.endswith(".mtx"):
        sparse = mmread(path)
        if not isinstance(sparse, coo_matrix):
            sparse = sparse.tocoo()
        g = nx.Graph()
        for u, v, _ in zip(sparse.row, sparse.col, sparse.data):
            if u != v:
                g.add_edge(int(u), int(v))
        return g
    return nx.read_edgelist(path, nodetype=int)


def read_clusters(path):
    node2cid = {}
    with open(path, encoding="utf8") as fh:
        for cid, line in enumerate(fh):
            for tok in line.split():
                node2cid[int(tok)] = cid
    return node2cid


def build_weighted_graph(g, cluster_of, intra_w, inter_w):
    for u, v in g.edges():
        same = cluster_of[u] == cluster_of[v]
        g[u][v]["weight"] = intra_w if same else inter_w


# ---------- Hauptprogramm --------------------------------------------
def main():
    p = argparse.ArgumentParser(
        description="Visualise graph clustering (tight intra-layout + legend).")
    p.add_argument("graph")
    p.add_argument("clusters")
    p.add_argument("-o", "--output")
    p.add_argument("--seed", type=int, default=42)
    p.add_argument("--intra-w", type=float, default=3.0)
    p.add_argument("--inter-w", type=float, default=0.05)
    args = p.parse_args()

    g = read_graph(args.graph)
    assign = read_clusters(args.clusters)

    # Isolierte Knoten abfangen
    if len(assign) < g.number_of_nodes():
        maxcid = max(assign.values(), default=-1)
        for v in g.nodes():
            assign.setdefault(v, maxcid := maxcid + 1)

    build_weighted_graph(g, assign, args.intra_w, args.inter_w)
    pos = nx.spring_layout(g, seed=args.seed, weight="weight")

    # Farbpalette
    nc = max(assign.values()) + 1
    palette = plt.cm.get_cmap("tab20", nc)
    node_colors = [palette(assign[v]) for v in g.nodes()]

    # Plot
    fig = plt.figure(figsize=(10, 10))
    ax = plt.gca()
    nx.draw_networkx_edges(g, pos, alpha=0.25, width=0.4, ax=ax)
    nx.draw_networkx_nodes(g, pos,
                           node_color=node_colors,
                           node_size=60,
                           edgecolors="black",
                           linewidths=0.3,
                           ax=ax)

    # Legende
    size_per_cluster = defaultdict(int)
    for cid in assign.values():
        size_per_cluster[cid] += 1
    handles = [
        Patch(facecolor=palette(cid), edgecolor="black",
              label=f"Cluster {cid+1} (|C|={size_per_cluster[cid]})")
        for cid in range(nc)
    ]
    ax.legend(handles=handles, fontsize="small",
              loc="upper right", frameon=True)

    # ----------------------------------- Titelzeilen ------------------
    # 1) Algorithmusname
    algo_name = os.path.splitext(os.path.basename(args.clusters))[0]
    fig.suptitle(algo_name, fontsize=16, y=0.97)

    # 2) Stats (kleinere Schrift)
    avg_csize = sum(size_per_cluster.values()) / nc
    stats = (f"|V|={g.number_of_nodes()}   "
             f"|E|={g.number_of_edges()}   "
             f"k={nc}   "
             f"⌀|C|={avg_csize:.1f}")
    ax.set_title(stats, fontsize=10, pad=4)

    # Layout & Save/Show
    ax.axis("off")
    plt.tight_layout(rect=[0, 0, 1, 0.96])  # Platz für suptitle
    if args.output:
        plt.savefig(args.output, dpi=300)
        print("saved", args.output)
    else:
        plt.show()


if __name__ == "__main__":
    main()
