package zeroinit

import (
	"github.com/kendru/darwin/go/depgraph"
)

// Graph represents a directed graph.
type Graph struct {
	*depgraph.Graph
}

// NewGraph creates a new instance of a Graph.
func NewGraph() *Graph {
	return &Graph{Graph: depgraph.New()}
}
