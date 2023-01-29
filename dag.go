package zeroinit

import (
	"context"
	"sync"

	"github.com/kendru/darwin/go/depgraph"
)

// Graph represents a directed graph.
type Graph struct {
	*depgraph.Graph
	ops map[string]*opState
}

type opState struct {
	fn    func(context.Context) error
	err   error
	fatal bool
}

// NewGraph creates a new instance of a Graph.
func NewGraph() *Graph {
	return &Graph{Graph: depgraph.New(), ops: make(map[string]*opState)}
}

type GraphOption func(string, *opState, *Graph) error

func WithDeps(deps ...string) GraphOption {
	return func(key string, os *opState, g *Graph) error {
		for _, d := range deps {
			if err := g.Graph.DependOn(key, d); err != nil {
				return err
			}
		}
		return nil
	}
}

func (g *Graph) AddOp(name string, fn func(context.Context) error, opts ...GraphOption) error {
	state := &opState{fn: fn}

	for _, o := range opts {
		if err := o(name, state, g); err != nil {
			return err
		}
	}
	g.ops[name] = state

	return nil
}

//func (g *Graph) Analyze()

func (g *Graph) Run(ctx context.Context) error {
	for _, layer := range g.TopoSortedLayers() {
		select {
		case <-ctx.Done():
		default:
			states := map[string]*opState{}

			for _, r := range layer {
				states[r] = g.ops[r]
			}

			var wg sync.WaitGroup
			for r, s := range states {
				fn := s.fn
				wg.Add(1)
				go func(ctx context.Context, g *Graph, key string) {
					defer wg.Done()
					g.ops[key].err = fn(ctx)
				}(ctx, g, r)
			}
			wg.Wait()
		}
	}
	return nil
}
