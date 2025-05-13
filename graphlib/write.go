package graphlib

import (
	"os"
	"sort"
	"strconv"
)

// exports adjacency matrices and adjacency lists to the plain-edge list and METIS format

// WriteToEdgeList Currently only supports undirected, unweighted graphs
func (l *AdjList) WriteToEdgeList(path string, heuristicSize int) (int, error) {
	if heuristicSize == 0 {
		heuristicSize = l.Size * 10
	}
	var buf = make([]byte, 0, heuristicSize) // heuristic could be average node degree

	for from, nbrs := range l.List {
		keys := make([]int, 0, len(nbrs))

		// this is only necessary for sorting
		for k := range nbrs {
			keys = append(keys, k)
		}
		// sorting
		sort.Ints(keys)

		for _, to := range keys {
			// “from to\n”
			buf = appendNum(buf, from+1)
			buf = append(buf, ' ')
			buf = appendNum(buf, to+1) // assumes 1-based external representation
			buf = append(buf, '\n')
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return 0, err
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)

	written, err2 := f.Write(buf)
	if err2 != nil {
		return 0, err2
	}

	return written, nil
}

func (l *AdjList) WriteToMETIS(path string, heuristicSize int) (int, error) {
	if heuristicSize == 0 {
		heuristicSize = l.Size * 10 * 3 // e.g. |vertices| × average node degree × bytes per neighbour
	}

	var buf = make([]byte, 0, heuristicSize) // heuristic could be average node degree

	// add the meta line: <number of vertices>SPACE<num edges>
	numVertices := l.Size
	totalDegree := l.TotalDegreeParallel(4)

	buf = appendNum(buf, numVertices)
	buf = append(buf, ' ')
	// total degree
	buf = appendNum(buf, totalDegree)
	buf = append(buf, '\n')

	for _, nbrs := range l.List {
		keys := make([]int, 0, len(nbrs))
		for k := range nbrs {
			keys = append(keys, k)
		}
		sort.Ints(keys)

		for _, to := range keys {
			buf = append(buf, ' ')
			buf = appendNum(buf, to+1)
		}
		buf = append(buf, '\n')
	}

	f, err := os.Create(path)
	if err != nil {
		return 0, err
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)

	written, err2 := f.Write(buf)
	if err2 != nil {
		return 0, err2
	}

	return written, nil
}

// appendNum is a helper for writing to a buffer
func appendNum(buf []byte, num int) []byte {
	return strconv.AppendUint(buf, uint64(num), 10)
}
