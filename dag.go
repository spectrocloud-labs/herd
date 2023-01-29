package zeroinit

import (
	"context"
	"sync"

	"github.com/kendru/darwin/go/depgraph"
)

// Graph represents a directed graph.
type Graph struct {
	*depgraph.Graph
	ops map[string]func(context.Context)
}

// NewGraph creates a new instance of a Graph.
func NewGraph() *Graph {
	return &Graph{Graph: depgraph.New(), ops: make(map[string]func(context.Context))}
}

func (g *Graph) AddOp(name string, fn func(context.Context), deps ...string) error {
	g.ops[name] = fn
	for _, d := range deps {
		if err := g.Graph.DependOn(name, d); err != nil {
			return err
		}
	}
	return nil
}

//func (g *Graph) Analyze()

func (g *Graph) Run(ctx context.Context) error {
	for _, layer := range g.TopoSortedLayers() {
		select {
		case <-ctx.Done():
		default:
			parallelFuncs := []func(context.Context){}
			for _, r := range layer {
				parallelFuncs = append(parallelFuncs, g.ops[r])
			}
			var wg sync.WaitGroup
			for i := range parallelFuncs {
				fn := parallelFuncs[i]
				wg.Add(1)
				go func(ctx context.Context) {
					defer wg.Done()
					fn(ctx)
				}(ctx)

			}
			wg.Wait()
			// wait group
		}
	}
	return nil
}
