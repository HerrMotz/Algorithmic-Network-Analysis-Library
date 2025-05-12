package graphlib

import "os"

// exports adjacency matrices and adjacency lists to the plain-edge list and METIS format

func WriteAdjMatDenseToEdgeList(m *AdjMatDense, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)

	// Write every arc to the edge list file.

	return nil
}
