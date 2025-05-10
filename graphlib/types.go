package graphlib

import "errors"

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

var EdgeOrArcExistsError = errors.New("edge or arc already exists")
