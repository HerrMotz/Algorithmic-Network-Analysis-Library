package graphlib

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// -----------------------------------------------------------------------------
// Loads plain edge lists, METIS or Matrix Market files into an AdjList.
// -----------------------------------------------------------------------------

// GraphFormat is the file format of a graph, see ReadGraphAdjListAs.
type GraphFormat int

const (
	// FormatAuto chooses the format from the file extension (see ReadGraphAdjList).
	FormatAuto GraphFormat = iota
	// FormatEdgeList is a plain edge list with one "u v" pair per line.
	FormatEdgeList
	// FormatMETIS is the METIS graph format: a header "n m [fmt [ncon]]" and
	// one adjacency line per vertex.
	FormatMETIS
	// FormatMatrixMarket is the Matrix Market coordinate format.
	FormatMatrixMarket
)

// ParseGraphFormat returns the GraphFormat called "auto", "edgelist", "metis"
// or "mtx", e.g. for a command-line flag.
func ParseGraphFormat(name string) (GraphFormat, error) {
	switch strings.ToLower(name) {
	case "", "auto":
		return FormatAuto, nil
	case "edgelist":
		return FormatEdgeList, nil
	case "metis":
		return FormatMETIS, nil
	case "mtx":
		return FormatMatrixMarket, nil
	}
	return FormatAuto, fmt.Errorf("unknown graph format %q (want auto, edgelist, metis or mtx)", name)
}

// ReadGraphAdjList opens *path* and returns the graph as an *AdjList*. The
// format is chosen by the file extension:
//
//   - .metis, .mgraph: METIS, first line “n m …”, next n lines adjacency lists
//   - .mtx: Matrix Market, “%%MatrixMarket matrix coordinate …” banner, size line, “i j” entries
//   - anything else: plain edge list, one “u v” per line
//
// Use ReadGraphAdjListAs to choose the format explicitly, e.g. for METIS files
// named *.graph. Ids are converted to 0-based ids: METIS and Matrix Market ids
// are 1-based (METIS files that use the id 0 are read as 0-based), and edge
// lists are renumbered so that their smallest id becomes 0.
// If *directed* is true, only the arc u→v is stored; otherwise both directions are recorded.
func ReadGraphAdjList(path string, directed bool) (*AdjList, error) {
	return ReadGraphAdjListAs(path, FormatAuto, directed)
}

// ReadGraphAdjListAs is ReadGraphAdjList with an explicitly chosen file format.
func ReadGraphAdjListAs(path string, format GraphFormat, directed bool) (*AdjList, error) {
	if format == FormatAuto {
		switch strings.ToLower(filepath.Ext(path)) {
		case ".metis", ".mgraph":
			format = FormatMETIS
		case ".mtx":
			format = FormatMatrixMarket
		default:
			format = FormatEdgeList
		}
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	switch format {
	case FormatMETIS:
		return readMetisToAdjList(r, directed)
	case FormatMatrixMarket:
		return readMatrixMarketToAdjList(r, directed)
	case FormatEdgeList:
		return readEdgeListToAdjList(r, directed)
	}
	return nil, fmt.Errorf("unknown graph format %d", format)
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
	size := n - subtract + 1
	adjList, err := InitializeAdjList(size)
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
	size := n - subtract + 1
	adjList, err := InitializeAdjList(size)
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
	minID := math.MaxInt

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
// Matrix Market format
// -----------------------------------------------------------------------------

// readMatrixMarketToAdjList reads a Matrix Market file in coordinate format, as
// distributed by the SuiteSparse Matrix Collection and Network Repository:
//
//	%%MatrixMarket matrix coordinate <field> <symmetry>
//	% comment lines
//	<rows> <cols> <entries>
//	<i> <j> [value]          (one line per entry, 1-based indices)
//
// Entry values are ignored. Symmetric (and skew-symmetric or Hermitian)
// matrices store only one triangle, so their entries are added in both
// directions, even if *directed* is true.
func readMatrixMarketToAdjList(r *bufio.Reader, directed bool) (*AdjList, error) {
	scanner := bufio.NewScanner(r)

	// ----- banner -----
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return nil, errors.New("Matrix Market: empty file")
	}
	banner := strings.Fields(strings.ToLower(scanner.Text()))
	if len(banner) < 5 || banner[0] != "%%matrixmarket" || banner[1] != "matrix" {
		return nil, fmt.Errorf("Matrix Market: invalid banner %q", scanner.Text())
	}
	if banner[2] != "coordinate" {
		return nil, fmt.Errorf("Matrix Market: only the coordinate format is supported, not %q", banner[2])
	}
	symmetric := banner[4] != "general"

	// ----- size line -----
	n, entries := 0, 0
	sizeFound := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "%") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			return nil, fmt.Errorf("Matrix Market: size line must contain rows, columns and entries, got %q", line)
		}
		rows, err1 := strconv.Atoi(fields[0])
		cols, err2 := strconv.Atoi(fields[1])
		nnz, err3 := strconv.Atoi(fields[2])
		if err1 != nil || err2 != nil || err3 != nil {
			return nil, fmt.Errorf("Matrix Market: bad size line %q", line)
		}
		if rows != cols {
			return nil, fmt.Errorf("Matrix Market: an adjacency matrix must be square, got %dx%d", rows, cols)
		}
		n, entries = rows, nnz
		sizeFound = true
		break
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if !sizeFound {
		return nil, errors.New("Matrix Market: missing size line")
	}

	adj, err := InitializeAdjList(n)
	if err != nil {
		return nil, err
	}

	// ----- entries -----
	read := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "%") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, fmt.Errorf("Matrix Market: expected at least two indices, got %q", line)
		}
		i, err1 := strconv.Atoi(fields[0])
		j, err2 := strconv.Atoi(fields[1])
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("Matrix Market: bad indices in %q", line)
		}
		if i < 1 || i > n || j < 1 || j > n {
			return nil, fmt.Errorf("Matrix Market: entry %q is outside the %dx%d matrix", line, n, n)
		}
		u, v := Vertex{i - 1}, Vertex{j - 1}

		if directed {
			err = adj.AddArc(Arc{From: u, To: v})
			if symmetric && (err == nil || errors.Is(err, EdgeOrArcExistsError)) {
				err = adj.AddArc(Arc{From: v, To: u})
			}
		} else {
			err = adj.AddEdge(Edge{U: u, V: v})
		}
		if err != nil && !errors.Is(err, EdgeOrArcExistsError) { // ignores duplicate entries
			return nil, err
		}
		read++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if read != entries {
		return nil, fmt.Errorf("Matrix Market: size line announces %d entries, found %d", entries, read)
	}
	return adj, nil
}

// -----------------------------------------------------------------------------
// METIS format
// -----------------------------------------------------------------------------

// metisGraph is the adjacency structure of a METIS file without its weights.
type metisGraph struct {
	n          int
	neighbours [][]int // neighbours[v] holds the neighbour ids of vertex v as written in the file
	offset     int     // subtract from an id in the file to get the 0-based vertex
}

// parseMETIS reads a graph in METIS format: a header "n m [fmt [ncon]]"
// followed by exactly one line per vertex that lists its neighbours. Lines
// starting with % are comments, and an empty line is a vertex without
// neighbours.
//
// fmt tells which weights the vertex lines contain: its last digit stands for
// edge weights (one after every neighbour), its middle digit for ncon vertex
// weights and its first digit for a vertex size (both at the start of the
// line). Weights and sizes are skipped.
//
// METIS numbers vertices from 1. Files that use the id 0 are treated as
// 0-based.
func parseMETIS(r io.Reader) (*metisGraph, error) {
	scanner := bufio.NewScanner(r)
	// allow very long adjacency lines
	scanner.Buffer(make([]byte, 64*1024), 32*1024*1024)

	// ----- header -----
	var n int
	var hasSize, hasVertexWeights, hasEdgeWeights bool
	ncon := 1
	headerFound := false
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
		if n, err = strconv.Atoi(fields[0]); err != nil || n <= 0 {
			return nil, fmt.Errorf("METIS: bad n in header: %q", fields[0])
		}
		if len(fields) >= 3 {
			f := fields[2]
			if len(f) > 3 || strings.Trim(f, "01") != "" {
				return nil, fmt.Errorf("METIS: bad fmt in header: %q", f)
			}
			f = strings.Repeat("0", 3-len(f)) + f
			hasSize, hasVertexWeights, hasEdgeWeights = f[0] == '1', f[1] == '1', f[2] == '1'
		}
		if len(fields) >= 4 {
			if ncon, err = strconv.Atoi(fields[3]); err != nil || ncon < 1 {
				return nil, fmt.Errorf("METIS: bad ncon in header: %q", fields[3])
			}
		}
		headerFound = true
		break
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if !headerFound {
		return nil, errors.New("METIS: missing header")
	}

	// number of leading values (vertex size and weights) on each vertex line
	skip := 0
	if hasSize {
		skip++
	}
	if hasVertexWeights {
		skip += ncon
	}
	step := 1 // values per neighbour
	if hasEdgeWeights {
		step = 2
	}

	// ----- one line per vertex -----
	g := &metisGraph{n: n, neighbours: make([][]int, 0, n)}
	minID := math.MaxInt
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "%") {
			continue
		}
		v := len(g.neighbours)
		if v == n {
			if line != "" {
				return nil, fmt.Errorf("METIS: too many adjacency lines (>%d)", n)
			}
			continue // trailing empty lines
		}

		fields := strings.Fields(line)
		if len(fields) < skip {
			return nil, fmt.Errorf("METIS: line %d has fewer than %d vertex weights", v+1, skip)
		}
		fields = fields[skip:]
		if len(fields)%step != 0 {
			return nil, fmt.Errorf("METIS: line %d has a neighbour without edge weight", v+1)
		}

		nbrs := make([]int, 0, len(fields)/step)
		for i := 0; i < len(fields); i += step {
			w, err := strconv.Atoi(fields[i])
			if err != nil {
				return nil, fmt.Errorf("METIS: bad neighbour id in line %d: %v", v+1, err)
			}
			nbrs = append(nbrs, w)
			minID = min(minID, w)
		}
		g.neighbours = append(g.neighbours, nbrs)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(g.neighbours) != n {
		return nil, fmt.Errorf("METIS: expected %d adjacency lines, got %d", n, len(g.neighbours))
	}

	// ----- decide offset -----
	g.offset = 1
	if minID == 0 {
		g.offset = 0
	}
	for v, nbrs := range g.neighbours {
		for _, w := range nbrs {
			if w-g.offset < 0 || w-g.offset >= n {
				return nil, fmt.Errorf("METIS: neighbour id %d of vertex %d out of range after shift", w, v+1)
			}
		}
	}
	return g, nil
}

func readMetisToAdjList(r *bufio.Reader, directed bool) (*AdjList, error) {
	g, err := parseMETIS(r)
	if err != nil {
		return nil, err
	}

	// ----- build adjacency -----
	adj, err := InitializeAdjList(g.n)
	if err != nil {
		return nil, err
	}

	for v, nbrs := range g.neighbours {
		for _, wRaw := range nbrs {
			w := wRaw - g.offset

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

// ReadMetisToSparse opens the METIS file *path* and returns the graph as an
// AdjMatSparse with one row per vertex (see parseMETIS for the accepted
// format). If *directed* is false each (v,w) pair is mirrored so that the
// matrix is symmetric.
//
// The routine keeps only the (i,j) pairs in memory, so it will happily ingest
// very large graphs.
func ReadMetisToSparse(path string, directed bool) (AdjMatSparse, error) {
	f, err := os.Open(path)
	if err != nil {
		return AdjMatSparse{}, err
	}
	defer f.Close()

	g, err := parseMETIS(f)
	if err != nil {
		return AdjMatSparse{}, err
	}

	// build the final i-j lists
	pairs := 0
	for _, nbrs := range g.neighbours {
		pairs += len(nbrs)
	}
	if !directed {
		pairs *= 2
	}
	s := make([]uint32, 0, pairs)
	d := make([]uint32, 0, pairs)

	for v, nbrs := range g.neighbours {
		src := uint32(v)
		for _, wRaw := range nbrs {
			w := uint32(wRaw - g.offset)

			s = append(s, src)
			d = append(d, w)

			if !directed && w != src { // mirroring (skip self-loops)
				s = append(s, w)
				d = append(d, src)
			}
		}
	}

	// sort + dedupe, then compress to CSR with one row per vertex, so that
	// isolated vertices at the end keep their (empty) rows
	if err := SortIJ(&s, &d); err != nil {
		return AdjMatSparse{}, err
	}
	ia, ja := compress(s, d, uint64(g.n))
	return AdjMatSparse{ia, ja}, nil
}

func readAdjListToSparse(r *bufio.Reader, directed bool) (*AdjMatSparse, error) {
	var s, d []uint32
	scanner := bufio.NewScanner(r)

	var minID uint32 = math.MaxUint32

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
