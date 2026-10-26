import re
import argparse
import os
import numpy as np
import matplotlib.pyplot as plt

def parse_time(time_str):
    """Convert time string with h, m, s, ms, µs to seconds."""
    if 'h' in time_str:
        h, m, s = re.match(r'(\d+)h(\d+)m([\d.]+)s', time_str).groups()
        return int(h) * 3600 + int(m) * 60 + float(s)
    elif 'ms' in time_str:
        return float(time_str.replace('ms', '')) / 1000
    elif 'µs' in time_str:
        return float(time_str.replace('µs', '')) / 1e6
    elif 's' in time_str:
        return float(time_str.replace('s', ''))
    else:
        raise ValueError(f"Unknown time format: {time_str}")

def truncate_time_str(tstr):
    """Truncate all decimal numbers in the time string to two decimal places."""
    return re.sub(r'(\d+\.\d+)', lambda m: f"{float(m.group(1)):.2f}", tstr)

def plot_benchmark(data, cap_seconds=3600, ylim_multiplier=3, output=None):
    variants = list(next(iter(data.values())).keys())
    datasets = list(data.keys())
    n_vars = len(variants)
    n_dsets = len(datasets)

    fig, ax = plt.subplots(figsize=(12, 8))
    ax.set_axisbelow(True)
    ax.grid(True, which='both', axis='y', linestyle='--', linewidth=0.5, alpha=0.7, zorder=0)

    x = np.arange(n_dsets)
    total_width = 0.8
    bar_width = total_width / n_vars
    offsets = np.linspace(-total_width/2 + bar_width/2, total_width/2 - bar_width/2, n_vars)

    for i, variant in enumerate(variants):
        entries = [data[d][variant] for d in datasets]
        display_values = [min(sec, cap_seconds) for sec, _ in entries]
        bars = ax.bar(x + offsets[i], display_values, width=bar_width, label=variant, zorder=3)

        for dset, bar, (_, orig_str) in zip(datasets, bars, entries):
            height = bar.get_height()
            annotation = truncate_time_str(orig_str)
            if 'h' in orig_str and dset.endswith('fb-pages-media.edges') and variant == 'SimpleIter':
                # Place this annotation to the right of the bar
                ax.annotate(
                    annotation,
                    xy=(bar.get_x() + bar.get_width(), height),
                    xytext=(5, 0),
                    textcoords='offset points',
                    ha='left',
                    va='center',
                    fontsize=8,
                    rotation=0,
                    zorder=4,
                )
            else:
                ax.text(
                    bar.get_x() + bar.get_width()/2,
                    height * 1.05,
                    annotation,
                    ha='center',
                    va='bottom',
                    fontsize=8,
                    rotation=90,
                    zorder=4,
                    clip_on=False
                )

    ax.set_xticks(x)
    ax.set_xticklabels([os.path.basename(d) for d in datasets], rotation=45, ha='right')
    ax.set_ylabel('Time (seconds, log scale)')
    ax.set_yscale('log')

    ymin, _ = ax.get_ylim()
    ax.set_ylim(ymin, cap_seconds * ylim_multiplier)

    ax.set_title('Benchmark: Bron–Kerbosch Variants Runtimes (Log Scale)')

    legend = ax.legend(loc='upper center', bbox_to_anchor=(0.5, 0.96), ncol=n_vars, frameon=True)
    legend.get_frame().set_facecolor('white')
    legend.get_frame().set_edgecolor('black')

    fig.subplots_adjust(top=0.90, bottom=0.15)
    plt.tight_layout()
    if output:
        plt.savefig(output, dpi=150)
        print("saved", output)
    else:
        plt.show()

def main():
    parser = argparse.ArgumentParser(description="Plot benchmark results for Bron–Kerbosch variants with log scale.")
    parser.add_argument('logfile', help="Path to the benchmark log file.")
    parser.add_argument('-o', '--output', help="Save the figure to this file instead of showing it.")
    args = parser.parse_args()

    all_data = {}
    with open(args.logfile, 'r', encoding='utf-8') as f:
        content = f.read().split('go run')
        for chunk in content[1:]:
            lines = chunk.strip().splitlines()
            graph_match = re.search(r'-graph="([^"]+)"', lines[0])
            if not graph_match:
                continue
            graph_path = graph_match.group(1)
            chunk_data = {}
            for line in lines:
                m = re.search(r'BronKerbosch(\w+).*Zeit: ([\dhms\.µ]+)', line)
                if m:
                    var, tstr = m.groups()
                    chunk_data[var] = (parse_time(tstr), tstr)
            if chunk_data:
                all_data[graph_path] = chunk_data

    plot_benchmark(all_data, output=args.output)

if __name__ == "__main__":
    main()
