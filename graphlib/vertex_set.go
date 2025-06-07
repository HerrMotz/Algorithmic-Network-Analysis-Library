package graphlib

type VertexSet map[int]Naught

// HasVertex checks if the vertex with the given ID exists in the adjacency list.
func (v *VertexSet) HasVertex(vert Vertex) bool {
	if vert.Id < 0 {
		return false
	}
	_, ok := (*v)[vert.Id]
	return ok
}

func CopyVertexSet(v *VertexSet) VertexSet {
	v1 := make(VertexSet, len(*v))
	for k, v := range *v {
		v1[k] = v
	}
	return v1
}

func (v *VertexSet) RemoveVertex(vert Vertex) {
	if !v.HasVertex(vert) {
		return
	}
	delete(*v, vert.Id)
}
