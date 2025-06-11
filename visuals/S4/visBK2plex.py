#!/usr/bin/env python3
import re
import argparse
import numpy as np
import matplotlib.pyplot as plt


def parse_time(time_str):
    """Convert time strings like '41.209µs', '1.23ms', '2s', '1m2.5s', '1h2m3s' to seconds."""
    if 'h' in time_str:
        h, m, s = re.match(r"(\d+)h(\d+)m([\d.]+)s", time_str).groups()
        return int(h)*3600 + int(m)*60 + float(s)
    if 'm' in time_str and 's' in time_str and 'ms' not in time_str:
        m, s = re.match(r"(\d+)m([\d.]+)s", time_str).groups()
        return int(m)*60 + float(s)
    if 'ms' in time_str:
        return float(time_str.replace('ms',''))/1000
    if 'µs' in time_str:
        return float(time_str.replace('µs',''))/1e6
    if time_str.endswith('s'):
        return float(time_str[:-1])
    raise ValueError(f"Unbekanntes Zeitformat: {time_str}")


def truncate_time_str(tstr):
    """Truncate decimals to two places for annotations."""
    return re.sub(r"(\d+\.\d+)", lambda m: f"{float(m.group(1)):.2f}", tstr)


def main():
    parser = argparse.ArgumentParser(description="Plot benchmark results as a bar chart.")
    parser.add_argument('logfile', help="Pfad zur Logdatei mit Benchmarks.")
    args = parser.parse_args()

    # Einlesen
    with open(args.logfile, 'r', encoding='utf-8') as f:
        lines = f.read().splitlines()

    dataset = lines[0].strip()
    data = []  # list of (variant, seconds, original)
    for line in lines[1:]:
        if re.match(r'^=====', line):
            continue
        m = re.match(r'^(\w+)\s+⇒.*Zeit:\s*([\d\.]+(?:µs|ms|s|m|h))', line)
        if m:
            var, tstr = m.groups()
            sec_val = parse_time(tstr)
            data.append((var, sec_val, tstr))

    variants = [d[0] for d in data]
    times = [d[1] for d in data]
    orig = [d[2] for d in data]

    # Plot
    fig, ax = plt.subplots(figsize=(10,6))
    ax.set_yscale('log')
    ax.bar(variants, times, zorder=3)
    ax.grid(True, axis='y', linestyle='--', linewidth=0.5, zorder=0)
    ax.set_ylabel('time (s, log scale)')
    ax.set_title(f'Runtime comparison on {dataset} for 2plex and Bron-Kerbosch variants')

    # X-Achsen-Beschriftungen
    ax.set_xticks(range(len(variants)))
    ax.set_xticklabels(variants, rotation=45, ha='right')

    # Annotieren mit horizontalen Zeit-Strings
    for bar, ann in zip(ax.patches, orig):
        height = bar.get_height()
        ax.text(
            bar.get_x() + bar.get_width()/2,
            height * 1.05,
            truncate_time_str(ann),
            ha='center', va='bottom', fontsize=8, rotation=0, zorder=4
        )

    plt.tight_layout()
    plt.show()

if __name__ == '__main__':
    main()
