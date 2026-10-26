import matplotlib.pyplot as plt
from pathlib import Path

HERE = Path(__file__).resolve().parent  # files are read from and written next to this script

# Data for out.pajek-erdos.graph
iters_erdos = {
    'NaiveAllPairs': 1,
    'ImprovedAllPairs': 2,
    'NaiveDiameter': 1,
    'ImprovedDiameter': 3,
}
ns_per_op_erdos = {
    'NaiveAllPairs': 2713055625,
    'ImprovedAllPairs':  939819042,
    'NaiveDiameter':    2826053958,
    'ImprovedDiameter': 3057517444,
}

# Data for out.dimacs10-netscie.graph
iters_dimacs = {
    'NaiveAllPairs':       121,
    'ImprovedAllPairs':     48,
    'NaiveDiameter':    529064,
    'ImprovedDiameter':      715,
}
ns_per_op_dimacs = {
    'NaiveAllPairs':   9898584,
    'ImprovedAllPairs':24044772,
    'NaiveDiameter':      2287,
    'ImprovedDiameter':  1654066,
}

# Helper to compute total time in seconds
def compute_times(iters, ns_per_op):
    return {algo: (iters[algo] * ns_per_op[algo]) / 1e9
            for algo in iters}

times_erdos = compute_times(iters_erdos, ns_per_op_erdos)
times_dimacs = compute_times(iters_dimacs, ns_per_op_dimacs)

# Common order of algorithms
algos = ['NaiveAllPairs', 'ImprovedAllPairs', 'NaiveDiameter', 'ImprovedDiameter']

# --- Plot 1: ERDOS dataset ---
plt.figure(figsize=(8, 5))
plt.bar(algos, [times_erdos[a] for a in algos])
plt.xlabel('Algorithm')
plt.ylabel('Run Time (s)')
plt.title('Run Time Comparison on out.pajek-erdos.graph')
plt.grid(axis='y', linestyle='--', alpha=0.5)
# annotate values
for i, algo in enumerate(algos):
    y = times_erdos[algo]
    plt.text(i, y * 1.01, f'{y:.2f}', ha='center', va='bottom')
plt.tight_layout()
plt.savefig(HERE / 'erdos_runtime_comparison.png')
plt.close()

# --- Plot 2: DIMACS dataset ---
plt.figure(figsize=(8, 5))
plt.bar(algos, [times_dimacs[a] for a in algos])
plt.xlabel('Algorithm')
plt.ylabel('Run Time (s)')
plt.title('Run Time Comparison on out.dimacs10-netscie.graph')
plt.grid(axis='y', linestyle='--', alpha=0.5)
# annotate values
for i, algo in enumerate(algos):
    y = times_dimacs[algo]
    plt.text(i, y * 1.01, f'{y:.2f}', ha='center', va='bottom')
plt.tight_layout()
plt.savefig(HERE / 'dimacs_runtime_comparison.png')
plt.close()
