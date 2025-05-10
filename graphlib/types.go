package graphlib

type Arc struct {
	From int `json:"from,omitempty"`
	To   int `json:"to,omitempty"`
}
type Edge struct {
	U int `json:"u,omitempty"`
	V int `json:"v,omitempty"`
}
