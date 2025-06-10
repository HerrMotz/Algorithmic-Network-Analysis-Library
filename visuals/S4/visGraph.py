import argparse
import networkx as nx
import matplotlib.pyplot as plt
from scipy.io import mmread
from scipy.sparse import coo_matrix
from collections import defaultdict

def main():
    parser = argparse.ArgumentParser(description="Plot graph and highlight maximum cliques.")
    parser.add_argument('graph', help="Path to edge list or .mtx file.")
    parser.add_argument('--output', help="Optional path to save the figure.", default=None)
    args = parser.parse_args()

    # Load graph (handles .mtx format)
    if args.graph.endswith('.mtx'):
        sparse_matrix = mmread(args.graph)
        if not isinstance(sparse_matrix, coo_matrix):
            sparse_matrix = sparse_matrix.tocoo()
        G = nx.Graph()
        for i, j, w in zip(sparse_matrix.row, sparse_matrix.col, sparse_matrix.data):
            if i != j:  # Avoid self-loops if not needed
                G.add_edge(i, j)
    else:
        G = nx.read_edgelist(args.graph, nodetype=int)

    # Find all cliques and select maximum ones
    cliques = list(nx.find_cliques(G))
    max_size = max(len(c) for c in cliques)
    max_cliques = [c for c in cliques if len(c) == max_size]

    # Track which nodes appear in multiple max cliques
    node_clique_count = defaultdict(int)
    for clique in max_cliques:
        for node in clique:
            node_clique_count[node] += 1
    overlapping_nodes = {node for node, count in node_clique_count.items() if count > 1}

    # Compute layout
    pos = nx.spring_layout(G, seed=42)
    plt.figure(figsize=(16, 16))

    # Draw all edges in black (non-clique edges)
    all_edges = set(G.edges())
    clique_edges = set()
    for clique in max_cliques:
        for i in range(len(clique)):
            for j in range(i + 1, len(clique)):
                if G.has_edge(clique[i], clique[j]):
                    clique_edges.add((clique[i], clique[j]))
    
    other_edges = all_edges - clique_edges
    nx.draw_networkx_edges(G, pos, edgelist=other_edges, edge_color='black', alpha=0.3, width=0.5)

    # Define a colormap for different cliques
    colors = plt.cm.tab10.colors  # 10 distinct colors

    # Draw each max clique with a different color
    for i, clique in enumerate(max_cliques):
        # Get edges within this clique
        edges = [(u, v) for u in clique for v in clique if u < v and G.has_edge(u, v)]
        
        # Draw edges in a unique color
        nx.draw_networkx_edges(
            G, pos,
            edgelist=edges,
            edge_color=colors[i % len(colors)],
            width=1.5,
            alpha=0.8
        )
        
        # Draw nodes in the same color (but skip overlapping nodes for now)
        non_overlapping_nodes = [node for node in clique if node not in overlapping_nodes]
        if non_overlapping_nodes:
            nx.draw_networkx_nodes(
                G, pos,
                nodelist=non_overlapping_nodes,
                node_size=100,
                node_color=[colors[i % len(colors)]],
                label=f'Max Clique {i+1} (Size {max_size})'
            )

    # Draw overlapping nodes in green
    if overlapping_nodes:
        nx.draw_networkx_nodes(
            G, pos,
            nodelist=list(overlapping_nodes),
            node_size=120,  # Slightly larger to emphasize
            node_color='red',
            label=f'Overlapping Nodes ({len(overlapping_nodes)})'
        )

    # Draw remaining nodes in black
    all_nodes = set(G.nodes())
    clique_nodes = set().union(*max_cliques)
    other_nodes = all_nodes - clique_nodes
    nx.draw_networkx_nodes(G, pos, nodelist=other_nodes, node_size=50, node_color='black')

    # Final touches
    plt.title(f'Dolphin with {len(max_cliques)} Maximum Cliques (Size {max_size})')
    plt.legend(loc='upper right')
    plt.axis('off')
    plt.tight_layout()

    if args.output:
        plt.savefig(args.output, dpi=300)
        print(f"Saved figure to {args.output}")
    else:
        plt.show()

if __name__ == "__main__":
    main()