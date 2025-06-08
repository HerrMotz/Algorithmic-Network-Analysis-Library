import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
from scipy.stats import kendalltau
import networkx as nx
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import networkx as nx
from matplotlib import colormaps
from matplotlib import colors

def plot_kendall_heatmap(measure_csv_paths, output_path):
    """
    Compute pairwise Kendall's τ between all centrality measures and plot a heatmap with grid boxes
    and annotated values.

    Parameters:
    - measure_csv_paths: dict[str, str]
        Mapping from measure name to CSV file path. Each CSV must have a 'Vertex' column.
    - output_path: str
        Path to save the heatmap image.
    """
    # Load rankings
    measures = list(measure_csv_paths.keys())
    rankings = {}
    for m, path in measure_csv_paths.items():
        df = pd.read_csv(path)
        rankings[m] = df['Vertex'].astype(int).tolist()

    # All unique vertices
    all_vertices = sorted({v for rank in rankings.values() for v in rank})

    # Position maps
    pos = {m: {v: i for i, v in enumerate(rank_list)} for m, rank_list in rankings.items()}

    # Compute tau matrix
    m = len(measures)
    tau_matrix = np.zeros((m, m))
    for i, mi in enumerate(measures):
        for j, mj in enumerate(measures):
            x = [pos[mi][v] for v in all_vertices]
            y = [pos[mj][v] for v in all_vertices]
            tau_val, _ = kendalltau(x, y)
            tau_matrix[i, j] = tau_val

    # Plot
    fig, ax = plt.subplots()
    im = ax.imshow(tau_matrix, cmap='coolwarm', vmin=-1, vmax=1)
    
    # Major ticks
    ax.set_xticks(np.arange(m))
    ax.set_yticks(np.arange(m))
    ax.set_xticklabels(measures, rotation=45, ha='right')
    ax.set_yticklabels(measures)
    
    # Minor ticks for grid
    ax.set_xticks(np.arange(-.5, m, 1), minor=True)
    ax.set_yticks(np.arange(-.5, m, 1), minor=True)
    ax.grid(which='minor', color='black', linestyle='-', linewidth=1)
    ax.tick_params(which='minor', bottom=False, left=False)
    
    # Annotate each cell
    for i in range(m):
        for j in range(m):
            ax.text(j, i, f"{tau_matrix[i, j]:.2f}",
                    ha='center', va='center', color='white' if abs(tau_matrix[i, j]) > 0.5 else 'black')

    # Colorbar
    cbar = fig.colorbar(im, ax=ax, fraction=0.046, pad=0.04)
    cbar.set_label("Kendall's τ")

    plt.tight_layout()
    plt.savefig(output_path)
    plt.close()

def plot_rank_scatter(csv_path_x, csv_path_y, measure_x, measure_y, output_path):
    """
    Plot a scatter of ranks: x-axis = rank in measure_x, y-axis = rank in measure_y.

    Parameters:
    - csv_path_x: str
        Path to CSV for measure X (must have 'Vertex' column).
    - csv_path_y: str
        Path to CSV for measure Y (must have 'Vertex' column).
    - measure_x: str
        Name of measure X (for axis label).
    - measure_y: str
        Name of measure Y (for axis label).
    - output_path: str
        Path to save the scatter plot image.
    """
    # Load and assign ranks by position in file
    df_x = pd.read_csv(csv_path_x)
    df_y = pd.read_csv(csv_path_y)
    df_x['Rank_x'] = np.arange(1, len(df_x) + 1)
    df_y['Rank_y'] = np.arange(1, len(df_y) + 1)

    # Merge on Vertex
    merged = pd.merge(df_x[['Vertex', 'Rank_x']], df_y[['Vertex', 'Rank_y']],
                      on='Vertex')
    
    mx = max(merged['Rank_x'].max(), merged['Rank_y'].max())

    # Plot scatter
    fig, ax = plt.subplots()
    ax.scatter(merged['Rank_x'], merged['Rank_y'])
    ax.set_xlabel(f"{measure_x} Rank")
    ax.set_ylabel(f"{measure_y} Rank")
    ax.set_title(f"Rank Scatter: {measure_x} vs {measure_y}")
    ax.plot([1, mx], [1, mx], color='red', linestyle='--', linewidth=1)
    plt.tight_layout()
    plt.savefig(output_path)
    plt.close()

def plot_network_with_color_and_size(graph_path, csv_path, output_path, measure_name):
    """
    Draw the network with node sizes and colors proportional to a centrality measure.
    Size and color scale both reflect the node's centrality value.
    
    Parameters:
    - graph_path: str
        Path to an edge-list file (one edge per line: "u v").
    - csv_path: str
        Path to CSV with columns ['Vertex','Score'].
    - output_path: str
        Path to save the network plot image.
    - measure_name: str
        Name of the centrality measure (for title and colorbar label).
    """
    # Load graph
    G = nx.read_edgelist(graph_path, comments="%", nodetype=int)
    
    # Load scores
    df = pd.read_csv(csv_path)
    cent = {int(v): s for v, s in zip(df['Vertex'], df['Score'])}
    
    # Normalize scores to [0,1]
    scores = np.array(list(cent.values()))
    min_s, max_s = scores.min(), scores.max()
    if max_s > min_s:
        norm_scores = {v: (s - min_s) / (max_s - min_s) for v, s in cent.items()}
    else:
        norm_scores = {v: 0.5 for v in cent}  # uniform mid-value if all equal
    
    # Prepare visuals
    node_sizes = [20 + 600 * norm_scores.get(node, 0) for node in G.nodes()]
    node_colors = [norm_scores.get(node, 0) for node in G.nodes()]
    
    # Layout
    pos = nx.spring_layout(G, iterations=20, seed=42)  
    
    # Draw edges
    fig, ax = plt.subplots(figsize=(12, 8))
    nx.draw_networkx_edges(G, pos, ax=ax, alpha=0.5)
    
    # Draw nodes with scatter to allow colorbar
    cmap = colormaps.get_cmap('autumn_r')
    norm = colors.Normalize(vmin=0, vmax=1)
    sc = ax.scatter(
        [pos[n][0] for n in G.nodes()],
        [pos[n][1] for n in G.nodes()],
        s=node_sizes,
        c=node_colors,
        cmap=cmap,
        norm=norm,
        edgecolors='k'
    )
    # Draw labels
    nx.draw_networkx_labels(G, pos, ax=ax, font_size=2)
    
    # Colorbar
    cbar = plt.colorbar(sc, ax=ax, fraction=0.046, pad=0.04)
    cbar.set_label(f"{measure_name} (normalized)")
    
    ax.set_title(f"Network view ({measure_name})")
    ax.axis('off')
    plt.tight_layout()
    plt.savefig(output_path, dpi=600)
    plt.close()


# Example usage (commented out):
measure_files = {
     'Degree': 'centrality_ranking/degree_ranking.csv',
     'Closeness': 'centrality_ranking/closeness_ranking.csv',
     'PowerC': 'centrality_ranking/power_closeness_ranking.csv',
     'Organ Donor': 'centrality_ranking/exponential_closeness_ranking.csv',
     'Power Iteration': 'centrality_ranking/power_iteration_ranking.csv',
     #'Katz': 'centrality_ranking/katz_centrality_ranking.csv',

     # !! hier weitere Measures hinzufügen !! als Beispiel: das sieht in der Heatmap dann gut aus desto mehr desto besser
     #'Eigenvector': 'centrality_ranking/eigenvector_ranking.csv',
     #'MyMeasure': 'centrality_ranking/mymeasure_ranking.csv',
}
plot_kendall_heatmap(measure_files, 'kendall_heatmap_erdos.png')

# Scatter nur zwischen zwei measures möglich, denke da reicht es, wenn ich nur das für erdos mache (oder Straßennetz, wenn ihr das möchtet)
# plot_rank_scatter('centrality_ranking/degree_ranking.csv', 'centrality_ranking/closeness_ranking.csv', 'Degree', 'Closeness', 'degree_vs_closeness_scatter.png')

# genauso für Netzwerkanzeige 
#plot_network_with_color_and_size('../../exampleGraphs/jena_road_unfolded_50.edgelist', 'centrality_ranking/exponential_closeness_ranking.csv', 'network_ExpoC.png', 'Organ Donor Centrality')

