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
// Loads plain edge lists or METIS files into an AdjList.
// -----------------------------------------------------------------------------

// ReadGraphAdjList opens *path* and returns the graph as an *AdjList*.
//
// • Plain edge‑list  -> one “u v” per line
// • METIS            -> first line “n m …”, next n lines adjacency lists
//
// The loader auto‑detects the format and converts any 1‑based ids to 0‑based.
// If *directed* is true, only the arc u→v is stored; otherwise both directions are recorded.
func ReadGraphAdjList(path string, directed bool) (*AdjList, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
		}
	}(f)

	r := bufio.NewReader(f)

	// Peek at the first non‑comment line so we can guess the format.

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".metis", ".mgraph":
		return readMetisToAdjList(r, directed)
	default:
		return readEdgeListToAdjList(r, directed)
	}
}

// buildAdjFromArcs turns a vertex count and arc slice into an *AdjList*.
// It ignores duplicate arcs.
func buildAdjFromArcs(n int, arcs []Arc, subtract int) (*AdjList, error) {
	adjList, err := InitializeAdjList(n)
	if err != nil {
		return nil, err
	}
	for _, arc := range arcs {
		if err := adjList.AddArc(Arc{Vertex{arc.From.Id - subtract}, Vertex{arc.To.Id - subtract}}); err != nil && !errors.Is(err, EdgeOrArcExistsError) { // ignores duplicate arcs
			return nil, err
		}
	}
	return adjList, nil
}

// buildAdjFromEdges turns a vertex count and arc slice into an *AdjList*.
// it ignores duplicate edges.
func buildAdjFromEdges(n int, edges []Edge, subtract int) (*AdjList, error) {
	adjList, err := InitializeAdjList(n)
	if err != nil {
		return nil, err
	}
	for _, edge := range edges {
		if err := adjList.AddEdge(Edge{Vertex{edge.U.Id - subtract}, Vertex{edge.V.Id - subtract}}); err != nil && !errors.Is(err, EdgeOrArcExistsError) { // ignored duplicate edges
			return nil, err
		}
	}
	return adjList, nil
}

// -----------------------------------------------------------------------------
// Plain “u v” per‑line edge list
// -----------------------------------------------------------------------------

func readEdgeListToAdjList(r *bufio.Reader, directed bool) (*AdjList, error) {
	scanner := bufio.NewScanner(r)
	var edges []Edge
	var arcs []Arc

	maxID := -1
	minID := INF

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "%") {
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
			arcs = append(arcs, Arc{From: Vertex{u}, To: Vertex{v}})
		} else {
			edges = append(edges, Edge{U: Vertex{u}, V: Vertex{v}})
		}
		maxID = max(maxID, u, v)
		minID = min(minID, u, v)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if directed {
		return buildAdjFromArcs(maxID, arcs, minID)
	} else {
		return buildAdjFromEdges(maxID, edges, minID)
	}
}

// -----------------------------------------------------------------------------
// METIS format
// -----------------------------------------------------------------------------
func readMetisToAdjList(r *bufio.Reader, directed bool) (*AdjList, error) {
	scanner := bufio.NewScanner(r)

	// ----- header -----
	var n, v int
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

	// ----- read all adjacency lists once -----
	neighs := make([][]int, 0, n)
	minID := int(^uint(0) >> 1) // +∞
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "%") {
			continue
		}

		if v >= n {
			return nil, fmt.Errorf("METIS: too many adjacency lines (>%d)", n)
		}

		raw := make([]int, 0)
		for _, tok := range strings.Fields(line) {
			w, err := strconv.Atoi(tok)
			if err != nil {
				return nil, fmt.Errorf("METIS: bad neighbour id in line %d: %v", v+1, err)
			}
			raw = append(raw, w)
			if w < minID {
				minID = w
			}
		}
		neighs = append(neighs, raw)
		v++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if v != n {
		return nil, fmt.Errorf("METIS: expected %d adjacency lines, got %d", n, v)
	}

	// ----- decide offset -----
	var offset int
	switch minID {
	case 0:
		offset = 0
	case 1:
		offset = 1
	default:
		return nil, fmt.Errorf("METIS: vertex ids must start with 0 or 1 (found %d)", minID)
	}

	// ----- build adjacency -----
	adj, err := InitializeAdjList(n)
	if err != nil {
		return nil, err
	}

	for v, nbrs := range neighs {
		for _, wRaw := range nbrs {
			w := wRaw - offset
			if w < 0 || w >= n {
				return nil, fmt.Errorf("METIS: neighbour id %d out of range after shift", wRaw)
			}

			if directed {
				if err := adj.AddArc(
					Arc{From: Vertex{v}, To: Vertex{w}},
				); err != nil && !errors.Is(err, EdgeOrArcExistsError) {
					return nil, err
				}
			} else {
				if err := adj.AddEdge(
					Edge{U: Vertex{v}, V: Vertex{w}},
				); err != nil && !errors.Is(err, EdgeOrArcExistsError) {
					return nil, err
				}
			}
		}
	}
	return adj, nil
}
