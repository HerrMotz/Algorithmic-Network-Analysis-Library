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

func ReadGraphSparseMatrix(path string, directed bool) (*AdjMatSparse, error) {
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

	return readAdjListToSparse(r, directed)
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

// ReadMetisToSparse opens *path*, auto-detects whether the file is 0- or 1-indexed
// and returns the graph as an AdjMatSparse. If *directed* is false each
// (v,w) pair is mirrored so that the matrix is symmetric.
//
// The routine keeps only the (i,j) pairs in memory, so it will happily ingest
// very large graphs.
func ReadMetisToSparse(path string, directed bool) (AdjMatSparse, error) {
	f, err := os.Open(path)
	if err != nil {
		return AdjMatSparse{}, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// allow very long adjacency lines
	scanner.Buffer(make([]byte, 64*1024), 32*1024*1024)

	// header
	var n int // number of vertices
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "%") {
			continue
		}
		fs := strings.Fields(line)
		if len(fs) < 2 {
			return AdjMatSparse{}, errors.New("METIS: header must contain at least n and m")
		}
		if n, err = strconv.Atoi(fs[0]); err != nil || n <= 0 {
			return AdjMatSparse{}, fmt.Errorf("METIS: bad n in header: %v", fs[0])
		}
		break
	}
	if err := scanner.Err(); err != nil {
		return AdjMatSparse{}, err
	}

	// collect raw pairs  (row = current line index, col = token)
	type pair struct {
		src uint32
		dst uint32
	}
	rawPairs := make([]pair, 0)
	var minID uint32 // necessary, because the compiler automagically gives minID type int32 for whatever reason
	minID = INF
	maxID := uint32(0)

	v := 0 // 0-based row index we will keep
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "%") {
			continue
		}
		if v >= n {
			return AdjMatSparse{}, fmt.Errorf("METIS: too many adjacency lines (>%d)", n)
		}

		for _, tok := range strings.Fields(line) {
			w64, err := strconv.ParseInt(tok, 10, 0)
			if err != nil {
				return AdjMatSparse{}, fmt.Errorf("METIS: bad neighbour id in line %d: %v", v+1, err)
			}
			w := uint32(w64)

			rawPairs = append(rawPairs, pair{src: uint32(v), dst: w})
			if w < minID {
				minID = w
			}
			if w > maxID {
				maxID = w
			}
		}
		v++
	}
	if err := scanner.Err(); err != nil {
		return AdjMatSparse{}, err
	}
	if v != n {
		return AdjMatSparse{}, fmt.Errorf("METIS: expected %d adjacency lines, got %d", n, v)
	}

	// decide index offset (0- vs 1-based)
	var offset uint32
	switch minID {
	case 0:
		offset = 0
	case 1:
		offset = 1
	default:
		return AdjMatSparse{}, fmt.Errorf("METIS: vertex ids must start with 0 or 1 (found %d)", minID)
	}
	if maxID-offset != uint32(n-1) {
		return AdjMatSparse{}, fmt.Errorf("METIS: header says n=%d, but max id after shift is %d", n, maxID-offset)
	}

	// build the final i-j lists
	s := make([]uint32, 0, len(rawPairs)*func() int {
		if directed {
			return 1
		}
		return 2
	}())
	d := make([]uint32, 0, cap(s))

	for _, p := range rawPairs {
		w := p.dst - offset
		if w >= uint32(n) {
			return AdjMatSparse{}, fmt.Errorf("METIS: neighbour id %d out of range after shift", p.dst)
		}

		s = append(s, p.src)
		d = append(d, w)

		if !directed && w != p.src { // mirroring (skip self-loops)
			s = append(s, w)
			d = append(d, p.src)
		}
	}

	// sort + dedupe, then compress to CSR
	if err := SortIJ(&s, &d); err != nil {
		return AdjMatSparse{}, err
	}
	mat, err := NewFromSortedIJ(s, d)
	if err != nil {
		return AdjMatSparse{}, err
	}
	return mat, nil
}

func readAdjListToSparse(r *bufio.Reader, directed bool) (*AdjMatSparse, error) {
	var s, d []uint32
	scanner := bufio.NewScanner(r)

	var minID uint32

	minID = INF

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "%") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			return &AdjMatSparse{}, fmt.Errorf("invalid adjacency list line (need at least 1 neighbor): %q", line)
		}

		src64, err := strconv.ParseUint(fields[0], 10, 32)
		if err != nil {
			return &AdjMatSparse{}, fmt.Errorf("invalid vertex id: %q", fields[0])
		}
		src := uint32(src64)
		if src < minID {
			minID = src
		}

		for _, field := range fields[1:] {
			dst64, err := strconv.ParseUint(field, 10, 32)
			if err != nil {
				return &AdjMatSparse{}, fmt.Errorf("invalid neighbor id: %q", field)
			}
			dst := uint32(dst64)
			if dst < minID {
				minID = dst
			}

			s = append(s, src)
			d = append(d, dst)
			if !directed {
				s = append(s, dst)
				d = append(d, src)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return &AdjMatSparse{}, err
	}

	// Normalize to zero indexing if file was one-indexed
	if minID == 1 {
		for i := range s {
			s[i]--
			d[i]--
		}
	}

	if err := SortIJ(&s, &d); err != nil {
		return &AdjMatSparse{}, err
	}

	m, e := NewFromSortedIJ(s, d)
	return &m, e
}

// readWeightedEdgeListToAdjList reads a text edge list where every non‑comment
// line has three fields «u v w».  Vertex IDs may start with 0 or 1 – the function
// auto‑detects the offset to store vertices as 0‑based. Blank lines or lines
// beginning with # or % are treated as comments.
//
// When *directed==false* the edge is stored symmetrically via *AddEdge*;
// otherwise via *AddArc*.
func readWeightedEdgeListToAdjList(r *bufio.Reader, directed bool) (*WeightedAdjList, error) {
	type triple struct {
		u, v int
		w    Weight
	}

	scanner := bufio.NewScanner(r)

	triples := make([]triple, 0)
	minID := int(^uint(0) >> 1) // +∞
	maxID := -1

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "%") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			return nil, fmt.Errorf("weighted edge list: expected at least three fields 'u v w', got %q", line)
		}

		u, err1 := strconv.Atoi(fields[0])
		v, err2 := strconv.Atoi(fields[1])
		w64, err3 := strconv.ParseFloat(fields[2], 64)
		if err1 != nil || err2 != nil || err3 != nil {
			return nil, fmt.Errorf("weighted edge list: bad numbers in %q", line)
		}

		triples = append(triples, triple{u: u, v: v, w: Weight(w64)})

		minID = min(minID, u, v)
		maxID = max(maxID, u, v)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// ---------- decide offset (0‑ vs 1‑based) ----------
	var offset int
	switch minID {
	case 0:
		offset = 0
	case 1:
		offset = 1
	default:
		return nil, fmt.Errorf("weighted edge list: vertex ids must start with 0 or 1 (found %d)", minID)
	}

	n := maxID - offset + 1

	g, err := InitializeWeightedAdjList(n)
	if err != nil {
		return nil, err
	}

	for _, t := range triples {
		u := t.u - offset
		v := t.v - offset
		if directed {
			if err := g.AddArc(Arc{From: Vertex{u}, To: Vertex{v}}, t.w); err != nil && !errors.Is(err, EdgeOrArcExistsError) {
				return nil, err
			}
		} else {
			if err := g.AddEdge(Edge{U: Vertex{u}, V: Vertex{v}}, t.w); err != nil && !errors.Is(err, EdgeOrArcExistsError) {
				return nil, err
			}
		}
	}

	return g, nil
}
