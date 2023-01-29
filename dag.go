package zeroinit

import (
	"context"
	"sync"

	"github.com/kendru/darwin/go/depgraph"
)

type opCallback func(context.Context) error

// Graph represents a directed graph.
type Graph struct {
	*depgraph.Graph
	ops map[string]*opState
}

type opState struct {
	fn    opCallback
	err   error
	fatal bool
}

// NewGraph creates a new instance of a Graph.
func NewGraph() *Graph {
	return &Graph{Graph: depgraph.New(), ops: make(map[string]*opState)}
}

func (g *Graph) AddOp(name string, opts ...GraphOption) error {
	state := &opState{}

	for _, o := range opts {
		if err := o(name, state, g); err != nil {
			return err
		}
	}
	g.ops[name] = state
	return nil
}

type GraphEntry struct {
	WithCallback bool
	Callback     opCallback
	Error        error
	Fatal        bool
	Name         string
}

func (g *Graph) buildStateGraph() (graph [][]GraphEntry) {
	for _, layer := range g.TopoSortedLayers() {

		states := []GraphEntry{}

		for _, r := range layer {
			states = append(states, GraphEntry{
				WithCallback: g.ops[r].fn != nil,
				Callback:     g.ops[r].fn,
				Error:        g.ops[r].err,
				Fatal:        g.ops[r].fatal,
				Name:         r,
			})
		}

		graph = append(graph, states)
	}
	return
}

func (g *Graph) Analyze() (graph [][]GraphEntry) {
	return g.buildStateGraph()
}

func (g *Graph) Run(ctx context.Context) error {
	for _, layer := range g.buildStateGraph() {
		var wg sync.WaitGroup

		for _, r := range layer {
			if !r.WithCallback {
				continue
			}
			fn := r.Callback
			wg.Add(1)
			go func(ctx context.Context, g *Graph, key string) {
				defer wg.Done()
				g.ops[key].err = fn(ctx)
			}(ctx, g, r.Name)
		}

		wg.Wait()

		for _, s := range layer {
			if s.Fatal && g.ops[s.Name].err != nil {
				return g.ops[s.Name].err
			}
		}
	}
	return nil
}
