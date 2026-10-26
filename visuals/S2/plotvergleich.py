import matplotlib.pyplot as plt
from pathlib import Path

HERE = Path(__file__).resolve().parent  # files are read from and written next to this script

# Zeiten in Sekunden
labels = ['Read', 'SSSP', 'All Pairs']
python_times = [0.9096, 0.15, 308.6996]
go_times     = [1.122036125, 0.080614042, 26.069707291]

# Figure mit 3 Subplots erstellen
fig, axes = plt.subplots(1, 3, figsize=(15, 5))

# Für jeden Task einen eigenen Balkenplot und Werteannotationen
for ax, label, py, go in zip(axes, labels, python_times, go_times):
    bars = ax.bar(['Python', 'Go'], [py, go])
    ax.set_title(label)
    ax.set_ylabel('Time (s)')
    ax.grid(True, linestyle='--', alpha=0.5)

    # 10–15% zusätzlichen Abstand oben durch Y-Limits
    max_height = max(py, go)
    ax.set_ylim(0, max_height * 1.15)

    # Werte über die Balken schreiben
    for bar in bars:
        height = bar.get_height()
        ax.annotate(
            f'{height:.4f}',
            xy=(bar.get_x() + bar.get_width() / 2, height),
            xytext=(0, 5),            # 5 Punkte über dem Balken
            textcoords="offset points",
            ha='center', va='bottom'
        )

# Gesamt-Titel und etwas mehr Platz oberhalb des Subplot-Bereichs
fig.suptitle('Comparison of Python and Go Implementation (n=2000, p=0.5)', fontsize=16)
fig.subplots_adjust(top=0.85)  # mehr Abstand für Titel und Annotations

# Diagramm speichern
plt.savefig(HERE / 'comparison.png')
