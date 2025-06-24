# mk_truth.py  –  robust für .mtx und .txt/.edges usw.
import sys, json, pathlib, networkx as nx

path = pathlib.Path(sys.argv[1])

if path.suffix.lower() == ".mtx":
    # ---- Matrix-Market --------------------------------------------------
    from scipy.io import mmread           # pip install scipy
    A = mmread(path).tocsr()              # sparse Adjazenzmatrix
    # Falls die Matrix nicht exakt symmetrisch ist, symmetrisieren
    if (A != A.T).nnz:
        A = A.maximum(A.T)
    G = nx.from_scipy_sparse_array(A)     # ab NX ≥ 3.0
else:
    # ---- Gewöhnliche Edge-List -----------------------------------------
    # * data=False -> Gewichts-Spalte ignorieren
    # * comments='%' damit MatrixMarket-Headerzeilen auch hier kein Problem wären
    G = nx.read_edgelist(path, nodetype=int, data=False, comments='%')

# 1-basige Indizes → 0-basiert umnummerieren (nur nötig, wenn größte Knoten-ID == |V|)
max_id = max(G.nodes)
if 0 not in G and 1 in G:
    mapping = {v: v - 1 for v in G.nodes}
    G = nx.relabel_nodes(G, mapping, copy=True)

cliques = [sorted(c) for c in nx.find_cliques(G)]
print(json.dumps(cliques))
