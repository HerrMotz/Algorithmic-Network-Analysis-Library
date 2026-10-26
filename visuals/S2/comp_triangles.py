import matplotlib.pyplot as plt
import numpy as np
from pathlib import Path

HERE = Path(__file__).resolve().parent  # files are read from and written next to this script

# Raw benchmark data: (iterations, ns_per_op)
data = {
    'Erdos': {
        'NodeIterator': (1552, 725081),
        'CNI':          (854, 1379271),
    },
    'Dimacs': {
        'NodeIterator': (7024, 149437),
        'CNI':          (3596, 316147),
    }
}

# Prepare datasets and methods
datasets = list(data.keys())
methods = ['NodeIterator', 'CNI']

# Compute total run times in seconds
times = {method: [] for method in methods}
for ds in datasets:
    for method in methods:
        iters, ns_op = data[ds][method]
        total_sec = iters * ns_op / 1e9
        times[method].append(total_sec)

# Plot setup
x = np.arange(len(datasets))
width = 0.35

fig, ax = plt.subplots(figsize=(8, 6))
bars1 = ax.bar(x - width/2, times['NodeIterator'], width, label='Node Iterator')
bars2 = ax.bar(x + width/2, times['CNI'], width, label='Chiba-Nishizeki (CNI)')

# Labels and title
ax.set_xticks(x)
ax.set_xticklabels(datasets)
ax.set_xlabel('Dataset')
ax.set_ylabel('Run Time (s)')
ax.set_title('Triangle Finding: Node Iterator vs. CNI')
ax.legend(loc='lower right')
ax.grid(axis='y', linestyle='--', alpha=0.5)

# Annotate bar values
for bars in (bars1, bars2):
    for bar in bars:
        height = bar.get_height()
        ax.annotate(f'{height:.6f}',
                    xy=(bar.get_x() + bar.get_width() / 2, height),
                    xytext=(0, 5), textcoords='offset points',
                    ha='center', va='bottom')

# Finalize
plt.tight_layout()
plt.savefig(HERE / 'triangles_comparison.png')
plt.show()
