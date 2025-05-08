package graphlib

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ReadGraph opens the file at a path and, depending on the extension or the
// file header, parses either a plain-edge list (one "u v" per line) or a
// METIS adjacency‑list graph.  Vertex ids become 0‑based inside Go.  The
// function returns the *CSRAdjMatrix ready for use.
func ReadGraph(path string, directed bool) (*CSRAdjMatrix, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// The first non ‑ comment line tells us whether this is a METIS file: METIS
	// starts with "n m" (two ints where m >= n‑1) while a plain-edge list
	// typically has more than two ints by the time we read the first two
	// lines.  We peek into the first line without consuming the reader.
	r := bufio.NewReader(f)
	header, err := r.Peek(1024) // enough for the first line
	if err != nil && err != io.EOF {
		return nil, err
	}
	firstLine := strings.TrimSpace(strings.SplitN(string(header), "\n", 2)[0])

	// Heuristic: if the file has exactly two or three ints on the first
	// line, treat it as METIS, otherwise edge list.  Users can force one
	// format by naming the file with .metis or .graph, which overrides the
	// heuristic.
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".metis", ".graph", ".gr":
		return readMetis(r, firstLine, directed)
	default:
		// No decisive extension – decide by header length.
		fields := strings.Fields(firstLine)
		if len(fields) <= 3 {
			return readMetis(r, firstLine, directed)
		}
		return readEdgeList(r, directed)
	}
}

// ---------------- internal helpers ----------------

// readEdgeList parses a "u v" per-line edge list.  Blank lines and lines
// beginning with '#' are ignored.  The loader accepts either 0‑ or 1‑based
// ids and converts to 0‑based automatically.
func readEdgeList(r *bufio.Reader, directed bool) (*CSRAdjMatrix, error) {
	scanner := bufio.NewScanner(r)
	var edges []Edge
	maxID := -1
	oneBased := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, fmt.Errorf("edge list: expected at least two ints per line, got %q", line)
		}

		u64, err1 := strconv.ParseInt(fields[0], 10, 0)
		v64, err2 := strconv.ParseInt(fields[1], 10, 0)
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("edge list: bad ints in %q", line)
		}
		u, v := int(u64), int(v64)

		// Detect whether ids are 1‑based (no zero ever seen but a '1').
		if u == 0 || v == 0 {
			oneBased = false
		}
		if !oneBased && (u == 1 || v == 1) && maxID < 1 {
			// still might be 1‑based – wait until we see a 0 to decide
			// conclusively; postpone conversion.
			oneBased = true
		}

		edges = append(edges, Edge{From: u, To: v})
		if u > maxID {
			maxID = u
		}
		if v > maxID {
			maxID = v
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Convert to 0‑based if necessary.
	if oneBased {
		for i := range edges {
			edges[i].From--
			edges[i].To--
		}
		maxID--
	}

	n := maxID + 1
	return InitializeCSRAdjMat(n, edges)
}

// readMetis parses the (unweighted) METIS graph format.  Comments start
// with '%'.  The first line is "n m [fmt [ncon]]".  Each of the next n
// lines lists the neighbours of vertex i+1 (1‑based).  If the graph is
//
//	undirected, the file contains every edge exactly once.
func readMetis(r *bufio.Reader, firstLine string, directed bool) (*CSRAdjMatrix, error) {
	scanner := bufio.NewScanner(io.MultiReader(strings.NewReader(firstLine+"\n"), r))

	// Read header – skip comment lines starting with '%'.
	var n, m int
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
		if m, err = strconv.Atoi(fields[1]); err != nil {
			return nil, fmt.Errorf("METIS: bad m in header: %v", err)
		}
		// Ignore fmt / ncon for now – we only support unweighted graphs.
		break
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if n <= 0 {
		return nil, errors.New("METIS: n must be positive")
	}

	edges := make([]Edge, 0, m)
	v := 0 // current 0‑based vertex while reading adjacency lists

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
				return nil, fmt.Errorf("METIS: bad neighbour id in line %d: %v", v+1, err)
			}
			w := int(w64) - 1 // convert to 0‑based
			if w < 0 || w >= n {
				return nil, fmt.Errorf("METIS: neighbour id %d out of range", w+1)
			}
			if directed || v <= w { // avoid duplicates for undirected files
				edges = append(edges, Edge{From: v, To: w})
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

	return InitializeCSRAdjMat(n, edges)
}
