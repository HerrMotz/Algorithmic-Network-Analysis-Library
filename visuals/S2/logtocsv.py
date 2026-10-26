import argparse
import csv
import re
from pathlib import Path

# Input: the output of `go test -bench ... -benchmem ./graphlib` (the log itself is not part of the repository)
parser = argparse.ArgumentParser(description="Convert a Go benchmark log into benchmark_metrics2.csv.")
parser.add_argument('logfile', help="benchmark log written by go test -bench")
parser.add_argument('-o', '--output', default=Path(__file__).resolve().parent / 'benchmark_metrics2.csv',
                    help="CSV file to write (default: benchmark_metrics2.csv next to this script)")
args = parser.parse_args()
input_file = args.logfile
output_file = args.output

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
    writer = csv.writer(csvfile, lineterminator='\n')
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
