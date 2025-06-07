package graphlib

type VertexSet map[int]Naught

// Size returns the number of vertices in the set.
func (v *VertexSet) Size() int {
	if v == nil {
		return 0
	}
	return len(*v)
}

// HasVertex checks if the vertex with the given ID exists in the adjacency list.
func (v *VertexSet) HasVertex(vert Vertex) bool {
	if vert.Id < 0 {
		return false
	}
	_, ok := (*v)[vert.Id]
	return ok
}

func (v *VertexSet) CopyAndRemoveVertex(vert Vertex) VertexSet {
	vNew := v.Copy()
	vNew.RemoveVertex(vert)
	return vNew
}

func (v *VertexSet) Copy() VertexSet {
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
