import re
import csv

# Pfade anpassen
input_file = 'benchmark.log'
output_file = 'benchmark_metrics.csv'

# Regex zum Parsen: n, Algorithmus, Iterationen, ns/op, B/op, allocs/op
pattern = re.compile(
    r'n=(\d+)/([^-/]+)-\d+\s+(\d+)\s+(\d+)\s+ns/op\s+(\d+)\s+B/op\s+(\d+)\s+allocs/op'
)

# Daten zwischenspeichern
records = {}

with open(input_file, 'r') as f:
    for line in f:
        m = pattern.search(line)
        if not m:
            continue
        n = int(m.group(1))
        algo = m.group(2)
        iters = int(m.group(3))
        ns_op = int(m.group(4))
        bytes_op = int(m.group(5))
        allocs_op = int(m.group(6))
        
        # Gesamte Metriken berechnen
        total_ns     = iters * ns_op
        total_bytes  = iters * bytes_op
        total_allocs = iters * allocs_op
        
        # Nach n und Algorithmus speichern
        records.setdefault(n, {})[algo] = (total_ns, total_bytes, total_allocs)

# Reihenfolge der Algorithmen und Spaltenüberschriften
algos   = ['NaiveDiameter', 'ImprovedDiameter', 'NaiveAllPairs', 'ImprovedAllPairs']
columns = ['n'] + [f'{algo}_{metric}' 
                   for algo in algos 
                   for metric in ('ns', 'bytes', 'allocs')]

# CSV schreiben
with open(output_file, 'w', newline='') as csvfile:
    writer = csv.writer(csvfile)
    writer.writerow(columns)
    for n in sorted(records):
        row = [n]
        for algo in algos:
            if algo in records[n]:
                row.extend(records[n][algo])
            else:
                row.extend([None, None, None])
        writer.writerow(row)

print(f"CSV-Datei '{output_file}' wurde erzeugt.")
