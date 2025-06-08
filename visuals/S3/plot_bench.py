import re
import pandas as pd
import matplotlib.pyplot as plt

# Map raw metric names from out.txt to display names
metric_map = {
    "DegreeCentralityAll": "DegreeCentralityAll",
    "ClosenessCentralityAll": "ClosenessCentralityAll",
    "PowerClosenessCentralityAll": "PowerClosenessCentralityAll",
    "ExponentialClosenessCentralityAll": "OrganDonorCentrality",
    "PowerIteration": "PowerIteration (10k iterations)",
}

# Container for parsed records
records = []

# Regex to parse lines like: n= 100 DegreeCentralityAll: 30.708µs
pattern = re.compile(r"n=\s*(\d+)\s+(\w+):\s*([\d\.]+)([µm]?s)")

with open("out.txt", "r", encoding="utf-8") as f:
    for line in f:
        m = pattern.match(line.strip())
        if not m:
            continue
        n_str, raw_metric, val_str, unit = m.groups()
        if raw_metric not in metric_map:
            continue

        n = int(n_str)
        val = float(val_str)
        # convert to seconds
        if unit == "s":
            time_sec = val
        elif unit == "ms":
            time_sec = val / 1e3
        elif unit == "µs":
            time_sec = val / 1e6
        else:
            raise ValueError(f"Unknown time unit: {unit}")

        records.append({
            "n": n,
            "metric": metric_map[raw_metric],
            "time_s": time_sec,
        })

# build DataFrame
df = pd.DataFrame(records)

# pivot for plotting
pivot_df = df.pivot(index="n", columns="metric", values="time_s")

# plot
plt.figure(figsize=(10, 6))
for metric in pivot_df.columns:
    plt.plot(
        pivot_df.index,
        pivot_df[metric],
        marker='o',            # add circle markers
        linestyle='-',         # solid lines
        markersize=4,          # size of the markers
        label=metric
    )

plt.xlabel("Number of Nodes (n)")
plt.ylabel("Time (seconds)")
plt.title("Timing of Centrality Measures")
plt.legend()
plt.grid(True)
plt.tight_layout()
plt.show()
