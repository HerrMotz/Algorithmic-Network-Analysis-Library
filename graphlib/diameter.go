package graphlib

import (
	"math"
	"slices"
)

func (l *AdjList) NaiveDiameter() int {
	diameter := 0

	for i := 0; i < l.Size; i++ {
		distances := l.SingleSourceShortestPath(&Vertex{i})
		_n := slices.Max(distances)
		if _n == math.MaxUint32 { // if there is more than one connected component
			return math.MaxUint32
		}
		diameter = max(diameter, _n)

	}

	return diameter
}
