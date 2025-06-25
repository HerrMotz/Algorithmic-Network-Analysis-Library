#!/usr/bin/env python3
"""
parse_stats_to_csv.py  (Runtime-frei)
Wandelt stats.txt in zwei CSVs um:
  • modularity.csv   (…_M)
  • disagreement.csv (…_D)
"""

import argparse
import csv
import re
from pathlib import Path

TIME_RE  = re.compile(r"t=(\d+(?:\.\d+)?)(µs|ms|s|m\d+(?:\.\d+)?s)")
GRAPH_RE = re.compile(r"-graph\s+([^\s]+)")

def parse_log(path: Path):
    rows = []
    graph = None
    with path.open() as fh:
        for ln in fh:
            if ln.startswith("go run"):
                gm = GRAPH_RE.search(ln)
                if gm:
                    graph = Path(gm.group(1)).stem
                continue

            if " k=" in ln and " Q=" in ln:
                parts = ln.split()
                combo              = parts[0]                   # mergeBFS_M
                heuristic, metric   = combo.rsplit("_", 1)
                k_val               = int(parts[1].split("=")[1])
                score_val           = float(parts[2].split("=")[1])

                rows.append(dict(
                    Graph=graph,
                    Heuristic=heuristic,
                    Metric=metric,
                    k=k_val,
                    Score=score_val
                ))
    return rows

def write_csv(rows, metric_tag, filename):
    subset = [r for r in rows if r["Metric"] == metric_tag]
    subset.sort(key=lambda r: (r["Graph"], r["Heuristic"]))
    with open(filename, "w", newline="") as fh:
        w = csv.DictWriter(fh, fieldnames=["Graph", "Heuristic", "k", "Score"])
        w.writeheader()
        for r in subset:
            w.writerow({k: r[k] for k in w.fieldnames})

def main():
    ap = argparse.ArgumentParser(description="stats.txt → two CSVs (M & D)")
    ap.add_argument("stats", help="Pfad zur Log-Datei stats.txt")
    args = ap.parse_args()

    rows = parse_log(Path(args.stats))
    write_csv(rows, "M", "modularity.csv")
    write_csv(rows, "D", "disagreement.csv")
    print("✓ erzeugt modularity.csv und disagreement.csv")

if __name__ == "__main__":
    main()
