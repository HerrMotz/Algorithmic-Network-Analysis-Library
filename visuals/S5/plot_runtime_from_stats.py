#!/usr/bin/env python3
"""
plot_runtime_single_axis.py
Liest stats.txt (Log) und zeichnet *einen* Balken-Chart:
x-Achse = Graphen, Balken = Heuristik+Metrik-Kombination.
"""

import argparse
import csv
import re
from pathlib import Path
from collections import defaultdict

import matplotlib.pyplot as plt
import numpy as np


# ------------------------------------------------------------------- Parsing
TIME_RE = re.compile(r"t=(?P<num>[0-9.]+)(?P<unit>µs|ms|s|m[0-9.]+s)")
GRAPH_RE = re.compile(r"-graph\s+([^\s]+)")

def ms(token: str) -> float:
    m = TIME_RE.search(token)
    num, unit = float(m["num"]), m["unit"]
    if unit.endswith("µs"): return num / 1_000
    if unit.endswith("ms"): return num
    if unit == "s":        return num * 1_000
    if unit.startswith("m"):  # "m<min>s"
        minutes = float(unit[1:-1])
        return minutes * 60_000 + num * 1_000
    raise ValueError(unit)

def pretty(ms_val: float) -> str:
    if ms_val < 1:      return f"{ms_val*1e3:.0f} µs"
    if ms_val < 1_000:  return f"{ms_val:.1f} ms"
    if ms_val < 60_000: return f"{ms_val/1_000:.2f} s"
    m, s = divmod(ms_val/1_000, 60)
    return f"{int(m)} m {s:.0f} s"

def canon(path: str) -> str:
    return Path(path).stem

def parse(path: Path):
    graphs, heuristics = [], set()
    times = defaultdict(dict)  # times[graph][combo] = ms
    curr = None
    with path.open() as fh:
        for ln in fh:
            if ln.startswith("go run"):
                m = GRAPH_RE.search(ln)
                if m:
                    curr = canon(m.group(1)); graphs.append(curr)
                continue
            if " t=" not in ln: continue
            name, *_, tok = ln.split()
            base, metric = name.rsplit("_", 1)   # mergeBFS, M|D
            heuristics.add(base)
            combo = f"{base}_{metric}"
            times[curr][combo] = ms(tok)
    return graphs, sorted(heuristics), times


# ------------------------------------------------------------------- Farben
# Pastell (M)  +  dunkler Ton (D)  je Heuristik
palette = {
    "mergeBFS": ("#F4D6C3", "#E2A76F"),
    "mergeRandom": ("#C8E6D4", "#7AB59A"),
    "halve": ("#F8D7DA", "#EA8D95"),
    "ice": ("#DCD7F3", "#9B92D6"),
}


# ------------------------------------------------------------------- Plot
def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("stats", help="stats.txt")
    ap.add_argument("-o", "--output")
    ap.add_argument("-s", "--summary", help="also write the runtime summary as CSV to this file")
    args = ap.parse_args()

    graphs, heuristics, times = parse(Path(args.stats))
    combos = [f"{h}_{m}" for h in heuristics for m in ("M", "D")]

    # ------------------------------------------------------------------- Runtime-Summary
    print("\n=== Runtime-Summary (ms) ===")
    total_mod, total_dis = 0.0, 0.0
    summary = []  # (graph, M, D, Δ)

    for g in graphs:
        sum_mod = sum(v for c, v in times[g].items() if c.endswith("_M"))
        sum_dis = sum(v for c, v in times[g].items() if c.endswith("_D"))
        total_mod += sum_mod
        total_dis += sum_dis
        diff = sum_mod - sum_dis
        print(f"{g:25s}:  M = {sum_mod:10.2f}   D = {sum_dis:10.2f}   Δ = {diff:+.2f}")
        summary.append((g, sum_mod, sum_dis, diff))

    overall_diff = total_mod - total_dis
    print("-" * 60)
    print(f"Total{'':21s}  M = {total_mod:10.2f}   D = {total_dis:10.2f}   "
        f"Δ = {overall_diff:+.2f} ms")
    print("============================\n")
    summary.append(("Total", total_mod, total_dis, overall_diff))

    if args.summary:
        with open(args.summary, "w", newline="") as fh:
            w = csv.writer(fh, lineterminator="\n")
            w.writerow(["Graph", "Modularity (ms)", "Disagreement (ms)", "Δ (ms)"])
            for g, m, d, delta in summary:
                w.writerow([g, f"{m:.2f}", f"{d:.2f}", f"{delta:.2f}"])
        print("saved", args.summary)

    n_graphs = len(graphs)
    n_combos = len(combos)
    bar_w = 0.8 / n_combos
    group_w  = len(combos) * bar_w                 
    x_pos = np.arange(n_graphs)

    fig, ax = plt.subplots(figsize=(10, 5))
    ax.set_yscale("log")
    ax.set_ylabel("Runtime [ms] (log)")
    ax.set_xticks(x_pos)
    ax.set_xticklabels(graphs, fontsize=11)
    ax.set_title("Runtime for each heuristic-metric combination", pad=10)
    ax.grid(True, axis="y", which="both", linestyle="--", alpha=0.4)

    legend_handles = []

    for ci, combo in enumerate(combos):
        base, metric = combo.split("_")
        col = palette[base][0 if metric == "M" else 1]

        # Bars nun symmetrisch um x_pos angeordnet
        xs = x_pos - group_w/2 + bar_w/2 + ci*bar_w
        ys = [times[g].get(combo, np.nan) for g in graphs]
        bars = ax.bar(xs, ys, bar_w, color=col, edgecolor="black")

        # Beschriftung: <1 ms ⇒ oberhalb, sonst im Balken
        for bar, y in zip(bars, ys):
            if not np.isfinite(y):
                continue
            center = bar.get_x() + bar.get_width()/2
            label  = pretty(y)
            if y < 1.0:                  # Mikro­sekunden-Balken
                ax.text(center, y*1.5, label, ha="center", va="bottom",
                        fontsize=7, rotation=90)
            else:
                ax.text(center, y*0.8, label, ha="center", va="top",
                        fontsize=7, rotation=90)
        # Legende nur einmal pro combo
        legend_handles.append(
            plt.Rectangle((0,0),1,1, facecolor=col, edgecolor="black",
                          label=f"{base} {metric}")
        )

    ax.legend(handles=legend_handles, fontsize="small", ncol=4, framealpha=0.9)
    plt.tight_layout()
    if args.output:
        plt.savefig(args.output, dpi=300)
        print("saved", args.output)
    else:
        plt.show()

if __name__ == "__main__":
    main()
