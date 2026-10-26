#!/bin/sh
# Regenerates results/ and the figures and tables in visuals/S5 from scratch.
# Run it from the repository root (needs Go and the packages in requirements.txt):
#
#   sh visuals/S5/regenerate.sh
#
# Runtimes in stats.txt depend on the machine; everything else is reproducible.
set -e

stats=visuals/S5/stats.txt
: > "$stats"
for run in abi:exampleGraphs/abi.edgelist \
           dimacs10:exampleGraphs/out.dimacs10-netscie.graph \
           erdos:exampleGraphs/out.pajek-erdos.graph; do
    name=${run%%:*}
    graph=${run#*:}
    cmd="go run ./cmd/clusterbench -graph $graph -out results/$name/"
    echo "$cmd" >> "$stats"
    $cmd >> "$stats"
    echo >> "$stats"
done

python visuals/S5/stats_to_csv.py "$stats"
python visuals/S5/plot_runtime_from_stats.py "$stats" \
    -o visuals/S5/plot_runtimes.png -s visuals/S5/runtime_summary.csv

# cluster drawings for the abi graph
for heuristic in mergeBFS halve mergeRandom ice; do
    for metric in M D; do
        python plot_clusters.py exampleGraphs/abi.edgelist \
            "results/abi/${heuristic}_${metric}.clusters.txt" \
            -o "visuals/S5/${heuristic}_${metric}.png"
    done
done
