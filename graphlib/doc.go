// Package graphlib implements graph representations and algorithms for network
// analysis: shortest paths and the diameter, structural parameters such as the
// degeneracy, triangles and clustering coefficients, centrality measures,
// maximal cliques and 2-plexes, and clustering heuristics.
//
// Most algorithms work on an [AdjList]. It stores the neighbours of every
// vertex in a map ([VertexSet]), so testing for an edge takes O(1), and the
// neighbours of a vertex can be iterated without copying them, which suits
// breadth-first searches. Vertices are numbered from 0 to n-1.
//
// The eigenvector-based centralities work on [AdjMatSparse], a sparse adjacency
// matrix in compressed sparse row format; [AdjList.ToAdjMatSparse] converts an
// AdjList.
//
// [ReadGraphAdjList] reads plain edge lists, METIS files and Matrix Market
// files.
package graphlib
