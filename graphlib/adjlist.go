package graphlib

import (
	"errors"
	"sort"
)

// ---------------------------
//  Adjacency List (slices)
// ---------------------------

type Naught struct{}

type AdjList struct {
	Size int              `json:"size,omitempty"`
	List []map[int]Naught `json:"list,omitempty"` // List[u] are neighbours of u
}

// InitializeAdjList creates an adjacency list with *size* isolated vertices.
func InitializeAdjList(size int) (*AdjList, error) {
	if size <= 0 {
		return nil, errors.New("size must be > 0")
	}
	l := &AdjList{
		Size: size,
		List: make([]map[int]Naught, size),
	}

	// Initialise the adjacency list for each vertex
	for i := 0; i < size; i++ {
		// Todo: If we have knowledge of the average vertex degree, we could preallocate the map with a size.
		l.List[i] = make(map[int]Naught, 4)
	}

	return l, nil
}

// AddArc adds an arc (u,v) to the graph.
func (l *AdjList) AddArc(arc Arc) error {
	if l.Size == 0 {
		return errors.New("graph is empty")
	}

	if !l.HasVertex(arc.From) || !l.HasVertex(arc.To) {
		return errors.New("vertex does not exist")
	}

	if _, ok := l.List[arc.From.Id][arc.To.Id]; ok {
		return EdgeOrArcExistsError
	}

	l.List[arc.From.Id][arc.To.Id] = Naught{}

	return nil
}

func (l *AdjList) AddEdge(edge Edge) error {
	if l.HasArc(Arc{edge.U, edge.V}) || l.HasArc(Arc{edge.V, edge.U}) {
		return EdgeOrArcExistsError
	}

	err := l.AddArc(Arc{edge.U, edge.V})
	if err != nil {
		return err
	}

	err = l.AddArc(Arc{edge.V, edge.U})
	if errors.Is(err, EdgeOrArcExistsError) && edge.V == edge.U {
		// Ignore self-loops
		return nil
	}

	return err
}

func (l *AdjList) HasEdge(edge Edge) bool {
	return l.HasArc(Arc{From: edge.U, To: edge.V}) || l.HasArc(Arc{From: edge.V, To: edge.U})
}

func (l *AdjList) HasArc(arc Arc) bool {
	_, ok := l.List[arc.From.Id][arc.To.Id]
	return ok
}

func (l *AdjList) AddVertex() {
	l.Size++
	l.List = append(l.List, make(map[int]Naught))
}

func (l *AdjList) HasVertex(v Vertex) bool {
	return v.Id >= 0 && v.Id < l.Size
}

func (l *AdjList) RemoveVertex(v Vertex) error {
	if !l.HasVertex(v) {
		return errors.New("vertex does not exist")
	}

	l.List[v.Id] = nil
	l.Size--

	for i := 0; i < l.Size; i++ {
		delete(l.List[i], v.Id)
	}

	return nil
}

func (l *AdjList) RemoveArc() error {
	if l.Size == 0 {
		return errors.New("graph is empty")
	}

	for i := 0; i < l.Size; i++ {
		for j := 0; j < l.Size; j++ {
			if l.HasArc(Arc{From: Vertex{i}, To: Vertex{j}}) {
				delete(l.List[i], j)
			}
		}
	}

	return nil
}

func (l *AdjList) RemoveEdge(edge Edge) error {
	if l.Size == 0 {
		return errors.New("graph is empty")
	}

	if !l.HasVertex(edge.U) || !l.HasVertex(edge.V) {
		return errors.New("vertex does not exist")
	}

	if l.HasEdge(edge) {
		delete(l.List[edge.U.Id], edge.V.Id)
		delete(l.List[edge.V.Id], edge.U.Id)
	} else {
		return EdgeOrArcExistsError
	}

	return nil
}

func (l *AdjList) Neighbours(u Vertex) (map[int]Naught, error) {
	if !l.HasVertex(u) {
		return nil, errors.New("vertex does not exist")
	}
	return l.List[u.Id], nil
}

func (l *AdjList) Degree(u Vertex) (int, error) {
	n, err := l.Neighbours(u)
	return len(n), err
}

// Computation of Parameters ÜS2T2

// Density |A| / |V| * (|V| - 1), |A| = arcCount number of arcs, |V| = number of vertices
func (l *AdjList) Density() (float64, error) {
	if l.Size == 0 {
		return 0, errors.New("graph is empty")
	}
	if l.Size == 1 {
		return 0, nil
	}

	// count arcs
	var arcCount int
	for _, neighbours := range l.List {
		arcCount += len(neighbours)
	}

	// max number of arcs = |V| * (|V| - 1)
	maxArcs := l.Size * (l.Size - 1)
	return float64(arcCount) / float64(maxArcs), nil
}

// MaxDegree max(|N(u)|) für alle u
func (l *AdjList) MaxDegree() (int, error) {
	if l.Size == 0 {
		return 0, errors.New("graph is empty")
	}

	// get degrees of all vertices and find the maximum
	_max := 0
	for i := 0; i < l.Size; i++ {
		deg := len(l.List[i])
		if deg > _max {
			_max = deg
		}
	}
	return _max, nil
}

// HIndex computes the h-index of the graph.
func (l *AdjList) HIndex() (int, error) {
	if l.Size == 0 {
		return 0, errors.New("graph is empty")
	}

	// get degrees of all vertices
	degrees := make([]int, 0, l.Size)
	for i := 0; i < l.Size; i++ {
		degrees = append(degrees, len(l.List[i]))
	}

	// sort degrees in descending order
	sort.Sort(sort.Reverse(sort.IntSlice(degrees)))

	// find h-index
	h := 0
	for i, d := range degrees {
		if d >= i+1 {
			h = i + 1
		} else {
			break
		}
	}
	return h, nil
}

// DegeneracyOrder implements the pseudocode from the lecture
func (l *AdjList) DegeneracyOrder() ([]Vertex, error) {
	n := l.Size
	if n == 0 {
		return nil, errors.New("graph is empty")
	}

	// σ ← empty list
	sigma := make([]Vertex, 0, n)

	// D ← array of length n, each an empty list
	D := make([][]int, n)
	for i := 0; i < n; i++ {
		D[i] = []int{}
	}

	// degree array
	deg := make([]int, n)

	// foreach v ∈ V(G)
	for v := 0; v < n; v++ {
		d := len(l.List[v])
		deg[v] = d
		D[d] = append(D[d], v)
	}

	// mark vertices already in σ
	inSigma := make([]bool, n)

	// repeat n times
	for i := 0; i < n; i++ {
		// find smallest j with D[j] non-empty
		j := 0
		for j < n && len(D[j]) == 0 {
			j++
		}
		if j >= n {
			break
		}

		// remove v ← D[j].remove()
		idx := len(D[j]) - 1
		v := D[j][idx]
		D[j] = D[j][:idx]

		// σ.append(v)
		sigma = append(sigma, Vertex{v})
		inSigma[v] = true

		// foreach u ∈ N(v)
		for u := range l.List[v] {
			if inSigma[u] {
				continue
			}
			du := deg[u]

			// remove u from D[du]
			for k, x := range D[du] {
				if x == u {
					D[du] = append(D[du][:k], D[du][k+1:]...)
					break
				}
			}

			// du ← du - 1
			deg[u] = du - 1

			// D[du-1].append(u)
			D[du-1] = append(D[du-1], u)
		}
	}
	return sigma, nil
}

// KCoreDecomp implements the pseudocode from the lecture
func (l *AdjList) KCoreDecomp() ([][]Vertex, [][]Vertex, error) {
	n := l.Size
	if n == 0 {
		return nil, nil, errors.New("graph is empty")
	}

	// 1) σ ← Degen-Order(G)
	sigma, err := l.DegeneracyOrder()
	if err != nil {
		return nil, nil, err
	}

	// build rank[v] = position of v in σ
	rank := make([]int, n)
	for i, v := range sigma {
		rank[v.Id] = i
	}

	// 2) D ← D^→_σ(G): orientate each edge u—w to ensure that rank[u] < rank[w]
	D := make([][]Vertex, n)
	for u := 0; u < n; u++ {
		for w := range l.List[u] {
			if rank[u] < rank[w] {
				D[u] = append(D[u], Vertex{w})
			}
		}
	}

	// 3) L[0..n-1] ← array of empty lists
	L := make([][]Vertex, n)
	k := 0

	// 6–9) for i starting at 0 until n-1
	for _, v := range sigma {
		dPlus := len(D[v.Id]) // is deg⁺_D(v)
		if dPlus > k {
			k = dPlus
		}
		L[k] = append(L[k], v)
	}

	// 10–12) fill empty slots in L
	for i := 0; i < n; i++ {
		if len(L[i]) == 0 {
			j := i + 1
			for j < n && len(L[j]) == 0 {
				j++
			}
			if j < n {
				L[i] = L[j]
			}
		}
	}
	return L, D, nil
}

// LocalClusteringCoefficient
// (from lecture:) C(v) = |{{u,w} ⊆ N(v) : {u,w} ∈ E}| / (deg(v) über 2).
func (l *AdjList) LocalClusteringCoefficient(v Vertex) (float64, error) {
	if !l.HasVertex(v) {
		return 0, errors.New("vertex does not exist")
	}

	// neighbours of v as slices
	neighMap, _ := l.Neighbours(v)
	k := len(neighMap)
	if k < 2 {
		return 0, nil
	}

	neighs := make([]Vertex, 0, k)
	for uID := range neighMap {
		neighs = append(neighs, Vertex{uID})
	}

	// count edges between neighbours
	var connected int
	for i := 0; i < k; i++ {
		for j := i + 1; j < k; j++ {
			u, w := neighs[i], neighs[j]
			if l.HasEdge(Edge{U: u, V: w}) {
				connected++
			}
		}
	}

	possible := float64(k*(k-1)) / 2.0
	return float64(connected) / possible, nil
}

// GlobalClusteringCoefficient using triangles (lecture)
// c(G) = 6·#triangles / # length-2 paths,
// where length-2 paths = ∑ₖ deg(k)·(deg(k)−1).
func (l *AdjList) GlobalClusteringCoefficient() (float64, error) {
	// 1) count all undirected triangles
	tris, err := l.TrianglesCNI()
	if err != nil {
		return 0, err
	}
	numTris := len(tris)

	// 2) count all 2-paths: ∑ₖ deg(k)*(deg(k)-1)
	var numPaths int
	for v := 0; v < l.Size; v++ {
		d := len(l.List[v])
		numPaths += d * (d - 1) // /2 bei ungerichteten?
	}

	if numPaths == 0 {
		return 0, nil
	}
	return 6 * float64(numTris) / float64(numPaths), nil
}

// Triangles ÜS2T3

// TrianglesNodeIterator finds all triangles in the graph using a node iterator.
func (l *AdjList) TrianglesNodeIterator() ([][3]Vertex, error) {
	n := l.Size
	if n < 3 {
		return nil, nil
	}

	// 1) count degrees
	deg := make([]int, n)
	for u := 0; u < n; u++ {
		deg[u] = len(l.List[u])
	}

	// 2) build forward adjacency list
	fwd := make([][]int, n)
	for u := 0; u < n; u++ {
		for v := range l.List[u] {
			if deg[u] < deg[v] || (deg[u] == deg[v] && u < v) {
				fwd[u] = append(fwd[u], v)
			}
		}
		sort.Ints(fwd[u])
	}

	// 3) merge
	var tris [][3]Vertex
	for u := 0; u < n; u++ {
		for _, v := range fwd[u] {
			a, b := fwd[u], fwd[v]
			i, j := 0, 0
			for i < len(a) && j < len(b) {
				switch {
				case a[i] < b[j]:
					i++
				case a[i] > b[j]:
					j++
				default:
					w := a[i]
					tris = append(tris, [3]Vertex{
						{Id: u}, {Id: v}, {Id: w},
					})
					i++
					j++
				}
			}
		}
	}

	return tris, nil
}

// TrianglesCNI lists all triangles via Chiba&Nishizeki.
// returns triples (u,v,w) with r[u]<r[v]<r[w].
func (l *AdjList) TrianglesCNI() ([][3]Vertex, error) {
	n := l.Size
	if n < 3 {
		return nil, nil
	}

	// 1) get Degeneracy-Ordering
	ordering, err := l.DegeneracyOrder()
	if err != nil {
		return nil, err
	}

	// get rank r[v] = pos of v in the ordering
	rank := make([]int, n)
	for i, v := range ordering {
		rank[v.Id] = i
	}
	out := make([][]int, n)
	for u := 0; u < n; u++ {
		for v := range l.List[u] {
			if rank[u] < rank[v] {
				out[u] = append(out[u], v)
			}
		}
	}

	// 3) mark them and find triples
	tris := make([][3]Vertex, 0)
	mark := make([]bool, n)

	// for u = v1, . . . , vn−2 do {
	for u := 0; u < n; u++ {
		// for each v ∈ N (u) do mark v;
		for _, w := range out[u] {
			mark[w] = true
		}

		// for each v ∈ N (u) do {
		for _, v := range out[u] {
			// for each w ∈ N (v) do {
			for _, w := range out[v] {
				// if w is marked then {
				if mark[w] {
					// output triangle {u, v, w};
					tris = append(tris, [3]Vertex{
						{Id: u}, {Id: v}, {Id: w},
					})
				}
			}
		}

		// unmark v;
		for _, w := range out[u] {
			mark[w] = false
		}
	}

	return tris, nil
}
