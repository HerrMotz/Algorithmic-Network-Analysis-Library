# Algorithmic Network Analysis Library

[![CI](https://github.com/HerrMotz/Algorithmic-Network-Analysis-Library/actions/workflows/ci.yml/badge.svg)](https://github.com/HerrMotz/Algorithmic-Network-Analysis-Library/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/HerrMotz/Algorithmic-Network-Analysis-Library/graphlib.svg)](https://pkg.go.dev/github.com/HerrMotz/Algorithmic-Network-Analysis-Library/graphlib)

A Go library and command-line tools for analysing networks: shortest paths and
diameters, centrality measures, maximal cliques and 2-plexes, and clustering
heuristics. It was written for the lab course *Algorithmische
Netzwerkanalyse* (Prof. Christian Komusiewicz) at Friedrich Schiller University
Jena in the summer term 2025. The repository also contains the evaluations we
made for the course's worksheets.

## Using the library

```sh
go get github.com/HerrMotz/Algorithmic-Network-Analysis-Library/graphlib
```

The library needs Go 1.24 or newer and has no dependencies besides
[Gonum](https://www.gonum.org/).

```go
package main

import (
	"fmt"
	"log"

	"github.com/HerrMotz/Algorithmic-Network-Analysis-Library/graphlib"
)

func main() {
	g, err := graphlib.ReadGraphAdjList("exampleGraphs/soc-karate.mtx", false)
	if err != nil {
		log.Fatal(err)
	}

	d, err := g.Degeneracy()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d vertices, %d edges, degeneracy %d\n", g.Size, g.TotalDegree()/2, d)

	cliques := g.BronKerboschDegeneracy()
	fmt.Printf("%d maximal cliques\n", len(cliques))

	clusters, err := g.MergeClusteringGreedyBFS(graphlib.PartialModularityMetric, graphlib.PartialModularityCompare)
	if err != nil {
		log.Fatal(err)
	}
	q, err := graphlib.TotalModularityMetric(g, clusters)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d clusters, modularity %.3f\n", len(clusters), q)
}
```

Run from the root of this repository, this prints

```
34 vertices, 78 edges, degeneracy 4
36 maximal cliques
4 clusters, modularity 0.410
```

## What is in the library

| Area | Functions and types |
|---|---|
| Graph representations | `AdjList` (adjacency list of maps, used by most algorithms), `WeightedAdjList`, `AdjMatDense`, `AdjMatSparse` (compressed sparse rows), `InducedGraph` |
| Reading and writing | `ReadGraphAdjList` / `ReadGraphAdjListAs` for edge lists, METIS and Matrix Market files; `ReadMetisToSparse`; `WriteToEdgeList`, `WriteToMETIS` |
| Paths and connectivity | `SingleSourceShortestPath`, `NaiveAllPairsShortestPath`, `ImprovedAllPairsShortestPaths` (also multi-threaded), `NaiveDiameter`, `ImprovedDiameter`, `CutVertices`, `NeighbourhoodSubsets` |
| Structural parameters | `Density`, `MaxDegree`, `HIndex`, `Degeneracy`, `DegeneracyOrder`, `KCoreDecomp` |
| Triangles and clustering coefficients | `TrianglesNodeIterator`, `TrianglesCNI`, `LocalClustering`, `GlobalClustering` |
| Centrality | Degree, closeness, power closeness and exponential closeness centrality (`…CentralityAll`, `…CentralityOf`, `TopK…`); eigenvector centrality with `PowerIteration` and `PowerIterationRayleigh`; `KatzCentrality` |
| Cliques and 2-plexes | Bron–Kerbosch without pivot, with pivot and with degeneracy ordering (`BronKerbosch…`); 2-plex search (`TwoPlex…`) |
| Clustering | `ComponentClustering`, `MergeClusteringGreedyBFS`, `MergeClusteringMetricWithRandomPivot`, `HalveClustering`, `ICEClustering`; modularity and disagreement as metrics (`Partial…Metric`, `Total…Metric`) |
| Random graphs | `GenerateRandomGraph` |

The package documentation is on
[pkg.go.dev](https://pkg.go.dev/github.com/HerrMotz/Algorithmic-Network-Analysis-Library/graphlib).

## Command-line tools

Run them from the repository root, so that the paths to `exampleGraphs/` work.

```sh
# all clustering heuristics with modularity and disagreement; writes results/abi/*.clusters.txt
go run ./cmd/clusterbench -graph exampleGraphs/abi.edgelist -out results/abi

# timings of the Bron–Kerbosch variants and of the 2-plex search
go run ./cmd/clique2plex -graph exampleGraphs/soc-karate.mtx

# centrality rankings as CSV files
go run ./cmd/centralitybench -graph exampleGraphs/jena_road_unfolded_50.metis -out rankings

# vertices, edges, degeneracy and maximal cliques of all example graphs
go run ./cmd/cliqueanalysis -out graph_clique_analysis.csv
```

The tools that take `-graph` choose the input format by the file extension
(`.metis`/`.mgraph`: METIS, `.mtx`: Matrix Market, otherwise an edge list);
`-format` overrides it.

## Tests and benchmarks

```sh
go test ./...                                   # takes about 10 s and 1 GB of memory
go test -run '^$' -bench Triangles ./graphlib   # a selection of the benchmarks
```

Running all benchmarks takes a long time; select them with `-bench`.

## Evaluations

The `visuals/` directory holds the evaluations for the worksheets, with the
Python scripts that made the figures:

- `S2`: shortest paths and diameter, naive and improved algorithms, and Go
  compared with the Python prototype
- `S3`: centrality rankings of a road network and two co-authorship networks,
  and how strongly they agree (Kendall's τ)
- `S4`: Bron–Kerbosch variants and 2-plexes
- `S5`: the clustering heuristics; the raw clusterings are in `results/`

The scripts need the packages in `requirements.txt`
(`pip install -r requirements.txt`). `sh visuals/S5/regenerate.sh` recomputes
`results/` and the S5 figures. `pygraphlib/` is the first, Python version of the
library (`python -m pygraphlib.main`), and `benchmarks/` compares it with
networkx.

## Example graphs

`exampleGraphs/` contains the graphs used in the tests and evaluations: social
and collaboration networks, road networks from OpenStreetMap and the METIS
sample graphs. Their sources and licenses are listed in
[exampleGraphs/README.md](exampleGraphs/README.md).

## Authors

Daniel Motz, Niklas Becker-Klöser and Maximilian Stock

## License

The code is licensed under the [MIT License](LICENSE). The example graphs are
redistributed under their own licenses, see
[exampleGraphs/README.md](exampleGraphs/README.md).
