package graphlib

import "slices"

func (l *AdjList) NaiveDiameter() int {
	diameter := 0

	for i := 0; i < l.Size; i++ {
		distances := l.SingleSourceShortestPath(&Vertex{i})
		diameter = max(diameter, slices.Max(distances))
	}

	return diameter
}
