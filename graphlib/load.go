// graphlib/load.go – loads plain edge lists or METIS files into an AdjList.

package graphlib

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// -----------------------------------------------------------------------------
// Public API
// -----------------------------------------------------------------------------

// ReadGraph opens *path* and returns the graph as an *AdjList*.
//
// • Plain edge‑list  -> one “u v” per line
// • METIS            -> first line “n m …”, next n lines adjacency lists
//
// The loader auto‑detects the format (file‑name extension overrides the
// heuristic) and converts any 1‑based ids to 0‑based.  If *directed* is true
// only the arc u→v is stored; otherwise both directions are recorded.
func ReadGraph(path string, directed bool) (*AdjList, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := bufio.NewReader(f)

	// Peek at the first non‑comment line so we can guess the format.

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".metis", ".mgraph":
		return readMetis(r, directed)
	default:
		return readEdgeList(r, directed)
	}
}

// buildAdj turns a vertex count and arc slice into an *AdjList*.
func buildAdjFromArcs(n int, arcs []Arc) (*AdjList, error) {
	adjList, err := InitializeAdjList(n)
	if err != nil {
		return nil, err
	}
	for _, arc := range arcs {
		adjList.AddArc(arc)
	}
	return adjList, nil
}

// buildAdj turns a vertex count and arc slice into an *AdjList*.
func buildAdjFromEdges(n int, edges []Edge) (*AdjList, error) {
	adjList, err := InitializeAdjList(n)
	if err != nil {
		return nil, err
	}
	for _, edge := range edges {
		adjList.AddEdge(edge)
	}
	return adjList, nil
}

// -----------------------------------------------------------------------------
// Plain “u v” per‑line edge list
// -----------------------------------------------------------------------------

func readEdgeList(r *bufio.Reader, directed bool) (*AdjList, error) {
	scanner := bufio.NewScanner(r)
	var edges []Edge
	var arcs []Arc

	maxID := -1

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, fmt.Errorf("edge list: expected at least two ints, got %q", line)
		}

		u64, err1 := strconv.ParseInt(fields[0], 10, 0)
		v64, err2 := strconv.ParseInt(fields[1], 10, 0)
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("edge list: bad ints in %q", line)
		}
		u, v := int(u64), int(v64)

		if directed {
			arcs = append(arcs, Arc{From: u - 1, To: v - 1}) // assumes the list is 1-indexed
		} else {
			edges = append(edges, Edge{U: u - 1, V: v - 1}) // assumes the list is 1-indexed
		}
		maxID = max(maxID, u, v)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if directed {
		return buildAdjFromArcs(maxID, arcs)
	} else {
		return buildAdjFromEdges(maxID, edges)
	}
}

// -----------------------------------------------------------------------------
// METIS format
// -----------------------------------------------------------------------------

func readMetis(r *bufio.Reader, directed bool) (*AdjList, error) {
	// scanner reads firstLine + the rest of the file.
	scanner := bufio.NewScanner(r)

	// -------- header --------
	var n int
	var v int
	var adj *AdjList

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "%") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, errors.New("METIS: header must contain at least n and m")
		}
		var err error
		if n, err = strconv.Atoi(fields[0]); err != nil {
			return nil, fmt.Errorf("METIS: bad n in header: %v", err)
		}
		break
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if n <= 0 {
		return nil, errors.New("METIS: n must be positive")
	}

	adj, err := InitializeAdjList(n)
	if err != nil {
		return nil, err
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "%") {
			continue
		}

		if v >= n {
			return nil, fmt.Errorf("METIS: too many adjacency lines (>%d)", n)
		}

		for _, token := range strings.Fields(line) {
			w64, err := strconv.ParseInt(token, 10, 0)
			if err != nil {
				return nil, fmt.Errorf("METIS: bad neighbour id in line %d: %v", v+1, err) // assumes the list is 1-indexed
			}
			w := int(w64) - 1   // assumes METIS is 1-indexed
			if w < 0 || w > n { // assumes the list is 1-indexed
				return nil, fmt.Errorf("METIS: neighbour id %d out of range", w+1) // assumes the list is 1-indexed
			}

			if directed {
				adj.AddArc(Arc{From: v, To: w})
			} else {
				adj.AddEdge(Edge{U: v, V: w})
			}
		}
		v++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if v != n {
		return nil, fmt.Errorf("METIS: expected %d adjacency lines, got %d", n, v)
	}

	return adj, nil
}
