#!/usr/bin/env python3
"""
Analyse von Graphdateien – Ermittlung von Knotenzahl, Kantenzahl, Dichte,
sowie realer und theoretischer Anzahl maximaler Cliquen.

Neu:
    • Berechnung der Degeneracy d(G) mittels NetworkX-k-Kern-Zerlegung
    • Worst-Case-Bound  n · 3^((d(G)+1)/3)
    • Ausgabe der Verhältnisse zur neuen Schranke
"""

import csv
import sys
from pathlib import Path
import networkx as nx

# Sehr große Ganzzahlen zulassen
sys.set_int_max_str_digits(100_000)

# ----------------------------
# Manuell aus Go-Ausgabe übernommene Werte
# ----------------------------
GRAPH_DATA = {
    "soc-karate.mtx":             {"clique_count": 36,       "max_clique_size": 5},
    "soc-dolphins.mtx":           {"clique_count": 84,       "max_clique_size": 5},
    "abi.edgelist":               {"clique_count": 30,       "max_clique_size": 20},
    "out.dimacs10-netscie.graph": {"clique_count": 613,      "max_clique_size": 20},
    "out.pajek-erdos.graph":      {"clique_count": 9_210,    "max_clique_size": 8},
    "fb-pages-public-figure.edges": {"clique_count": 150_772, "max_clique_size": 21},
    "fb-pages-media.edges":         {"clique_count": 126_862, "max_clique_size": 31},
}

# ----------------------------
# Hilfsfunktionen
# ----------------------------
def moon_moser_bound(n: int) -> float:
    """Klassische Moon-Moser-Schranke 3^(n/3)."""
    return 3 ** (n // 3)

def adjusted_omega_bound(n: int, omega: int) -> float:
    """
    Schranke (n - ω + 1) · 3^((ω-1)/3) für Graphen mit bekanntem Max-Clique-Size ω.
    """
    if omega <= 0:
        return moon_moser_bound(n)
    return (n - omega + 1) * 3 ** ((omega - 1) // 3)

def degeneracy_bound(n: int, degeneracy: int) -> float:
    """
    Neue Worst-Case-Schranke n · 3^((d(G)+1)/3), vgl. Eppstein et al. 2010.
    """
    return n * 3 ** ((degeneracy + 1) // 3)

# ----------------------------
# Kernanalyse pro Graph
# ----------------------------
def analyze_graph(graph_path: str) -> dict | None:
    try:
        graph_name = Path(graph_path).name
        print(f"Processing {graph_name} …")

        # ------------- Einlesen des Graphen -------------
        if graph_path.endswith('.mtx'):
            # Matrix-Market (Koordinatenformat) manuell parsen
            G = nx.Graph()
            with open(graph_path, 'r') as f:
                for line in f:
                    if line.startswith('%') or not line.strip():
                        continue
                    u, v, *_ = line.split()
                    G.add_edge(int(u), int(v))
        else:
            # ------------------------------------------------------------
            # Fehlertolerantes Einlesen beliebiger Edge-List-Formate
            # ------------------------------------------------------------
            G = nx.Graph()

            # Datei öffnen und prüfen, ob .graph-Headerzeile vorhanden ist
            with open(graph_path, "r", encoding="utf-8", errors="ignore") as f:
                first_line = f.readline().strip()
                has_header = graph_path.endswith(".graph") and first_line.isdigit()
                if not has_header:
                    f.seek(0)  # kein Header → wieder an den Dateianfang

                for line in f:
                    line = line.strip()
                    if not line or line.startswith("#"):
                        continue          # Kommentare / Leerzeilen überspringen

                    parts = line.split()
                    if len(parts) < 2:
                        continue          # unvollständige Zeile ignorieren

                    u_token, v_token = parts[:2]   # nur die ersten beiden Spalten

                    # Knoten möglichst als int, sonst als str
                    def cast(token: str):
                        try:
                            return int(token)
                        except Exception:
                            return token

                    u = cast(u_token)
                    v = cast(v_token)
                    G.add_edge(u, v)


        n = G.number_of_nodes()
        m = G.number_of_edges()
        density = (2 * m) / (n * (n - 1)) if n > 1 else 0.0

        # ------------- Degeneracy d(G) -------------
        try:
            core_numbers = nx.core_number(G)
            degeneracy = max(core_numbers.values())
        except nx.NetworkXError:
            degeneracy = 0  # tritt nur bei leerem Graph auf

        # ------------- Manuelle Go-Daten -------------
        manual = GRAPH_DATA.get(graph_name, {})
        clique_count     = manual.get("clique_count",     0)
        max_clique_size  = manual.get("max_clique_size",  0)

        # ------------- Schranken -------------
        mm_bound        = moon_moser_bound(n)
        omega_bound     = adjusted_omega_bound(n, max_clique_size)
        degen_bound     = degeneracy_bound(n, degeneracy)

        return {
            "graph_name":               graph_name,
            "nodes":                    n,
            "edges":                    m,
            "density":                  f"{density:.6f}",
            "degeneracy":               degeneracy,
            "max_clique_size":          max_clique_size,
            "clique_count":             clique_count,
            "moon_moser_bound":         f"{mm_bound:.3e}",
            "adjusted_omega_bound":     f"{omega_bound:.3e}",
            "degeneracy_bound":         f"{degen_bound:.3e}",
            "rel_to_moon_moser":        f"{clique_count / mm_bound:.3e}"      if mm_bound     else "n/a",
            "rel_to_adjusted_omega":    f"{clique_count / omega_bound:.3e}"   if omega_bound  else "n/a",
            "rel_to_degeneracy_bound":  f"{clique_count / degen_bound:.3e}"   if degen_bound  else "n/a",
        }

    except Exception as err:
        print(f"Error processing {graph_path}: {err}")
        return None

# ----------------------------
# Hauptprogramm
# ----------------------------
def main() -> None:
    graph_files = [
        "../../exampleGraphs/soc-karate.mtx",
        "../../exampleGraphs/soc-dolphins.mtx",
        "../../exampleGraphs/abi.edgelist",
        "../../exampleGraphs/out.dimacs10-netscie.graph",
        "../../exampleGraphs/out.pajek-erdos.graph",
        "../../exampleGraphs/fb-pages-public-figure.edges",
        "../../exampleGraphs/fb-pages-media.edges",
    ]

    results: list[dict] = []
    for file in graph_files:
        stats = analyze_graph(file)
        if stats:
            results.append(stats)

    # ------------- CSV schreiben -------------
    out_file = "graph_clique_analysis.csv"
    with open(out_file, "w", newline='') as csvfile:
        fieldnames = [
            "graph_name", "nodes", "edges", "density",
            "degeneracy", "max_clique_size", "clique_count",
            "moon_moser_bound", "adjusted_omega_bound", "degeneracy_bound",
            "rel_to_moon_moser", "rel_to_adjusted_omega", "rel_to_degeneracy_bound",
        ]
        writer = csv.DictWriter(csvfile, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(results)

    print(f"\nAnalyse abgeschlossen – Ergebnisse in {out_file} gespeichert.")

# ----------------------------
if __name__ == "__main__":
    main()
