package graphlib

import "slices"

func (l *AdjList) naiveDiameter() int {
	diameter := 0

	for i := 0; i < l.Size; i++ {
		distances := l.singleSourceShortestPath(&Vertex{i})
		diameter = max(diameter, slices.Max(distances))
	}

	return diameter
}
