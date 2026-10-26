package graphlib

import (
	"errors"
	"math"
)

type Arc struct {
	From Vertex `json:"from,omitempty"`
	To   Vertex `json:"to,omitempty"`
}
type Edge struct {
	U Vertex `json:"u,omitempty"`
	V Vertex `json:"v,omitempty"`
}

type Vertex struct {
	Id int
}

// INF is the distance of an unreachable vertex. It is math.MaxInt32, so that it
// fits into an int on 32-bit platforms, too.
const INF = math.MaxInt32

var EdgeOrArcExistsError = errors.New("edge or arc already exists")

// EdgeOrArcNotFoundError is returned when an edge or arc that should be removed
// does not exist.
var EdgeOrArcNotFoundError = errors.New("edge or arc does not exist")
