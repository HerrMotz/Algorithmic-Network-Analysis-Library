import pandas as pd
import matplotlib.pyplot as plt

# Load CSV file (adjust path if needed)
df = pd.read_csv('benchmark_metrics.csv')

# Algorithms in desired order
algos = ['NaiveAllPairs', 'ImprovedAllPairs', 'NaiveDiameter', 'ImprovedDiameter']

# Convert times from nanoseconds to seconds
for algo in algos:
    df[f'{algo}_time_s'] = df[f'{algo}_ns'] / 1e9

# Convert memory from bytes to megabytes
for algo in algos:
    df[f'{algo}_mem_mb'] = df[f'{algo}_bytes'] / (1024**2)

# --- Plot 1: Run Time vs n ---
plt.figure(figsize=(10, 6))
for algo in algos:
    plt.plot(df['n'], df[f'{algo}_time_s'], marker='o', label=algo)
plt.xlabel('n')
plt.ylabel('Run Time (s)')
plt.title('Run Time vs. n')
plt.legend()
plt.grid(True, linestyle='--', alpha=0.5)
plt.tight_layout()
plt.savefig('runtime_vs_n.png')
plt.close()  # Schließt die Figure, damit die nächste sauber startet

# --- Plot 2: Memory Usage vs n ---
plt.figure(figsize=(10, 6))
for algo in algos:
    plt.plot(df['n'], df[f'{algo}_mem_mb'], marker='s', label=algo)
plt.xlabel('n')
plt.ylabel('Memory Usage (MB)')
plt.title('Memory Usage vs. n')
plt.legend()
plt.grid(True, linestyle='--', alpha=0.5)
plt.tight_layout()
plt.savefig('memory_usage_vs_n.png')
plt.close()  # Schließt auch diese Figure

print("Saved two files: 'runtime_vs_n.png' and 'memory_usage_vs_n.png'")
