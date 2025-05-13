package graphlib

import (
	"os"
	"sort"
	"strconv"
)

// exports adjacency matrices and adjacency lists to the plain-edge list and METIS format

// WriteToEdgeList Currently only supports undirected, unweighted graphs
func (l *AdjList) WriteToEdgeList(path string) error {
	var buf []byte

	for from, nbrs := range l.List {
		keys := make([]int, 0, len(nbrs))
		for k := range nbrs {
			keys = append(keys, k)
		}
		sort.Ints(keys)

		for _, to := range keys {
			// “from to\n”
			buf = strconv.AppendInt(buf, int64(from), 10)
			buf = append(buf, ' ')
			buf = strconv.AppendInt(buf, int64(to), 10)
			buf = append(buf, '\n')
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)

	_, err = f.Write(buf)
	if err != nil {
		return err
	}

	return nil
}

func (l *AdjList) WriteAdjListToMetis(path string) error {

}
