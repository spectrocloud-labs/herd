package zeroinit

import (
	"context"
	"fmt"
	"sync"

	"github.com/kendru/darwin/go/depgraph"
)

// Graph represents a directed graph.
type Graph struct {
	*depgraph.Graph
	ops map[string]*opState
}

// GraphEntry is the external representation of
// the operation to execute (opState)
type GraphEntry struct {
	WithCallback    bool
	Background      bool
	Callback        opCallback
	Error           error
	Fatal, WeakDeps bool
	Name            string
}

// NewGraph creates a new instance of a Graph.
func NewGraph() *Graph {
	return &Graph{Graph: depgraph.New(), ops: make(map[string]*opState)}
}

func (g *Graph) AddOp(name string, opts ...GraphOption) error {
	state := &opState{Mutex: sync.Mutex{}}

	for _, o := range opts {
		if err := o(name, state, g); err != nil {
			return err
		}
	}
	g.ops[name] = state
	return nil
}

func (g *Graph) State(name string) GraphEntry {
	g.ops[name].Lock()
	defer g.ops[name].Unlock()
	return g.ops[name].toGraphEntry(name)
}

func (g *Graph) buildStateGraph() (graph [][]GraphEntry) {
	for _, layer := range g.TopoSortedLayers() {
		states := []GraphEntry{}

		for _, r := range layer {
			g.ops[r].Lock()
			states = append(states, g.ops[r].toGraphEntry(r))
			g.ops[r].Unlock()
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

	LAYER:
		for _, r := range layer {
			if !r.WithCallback {
				continue
			}
			fn := r.Callback

			if !r.WeakDeps {
				nodeSet := g.Graph.Dependencies(r.Name)
				for k := range nodeSet {
					g.ops[r.Name].Lock()
					g.ops[k].Lock()

					if g.ops[k].err != nil {
						g.ops[r.Name].err = fmt.Errorf("'%s' deps %s failed", r.Name, k)
						g.ops[r.Name].Unlock()
						g.ops[k].Unlock()

						continue LAYER
					}
					g.ops[r.Name].Unlock()
					g.ops[k].Unlock()

				}
			}

			if !r.Background {
				wg.Add(1)
			}
			go func(ctx context.Context, g *Graph, key string) {
				err := fn(ctx)
				g.ops[key].Lock()
				g.ops[key].err = err
				if !g.ops[key].background {
					wg.Done()
				}
				g.ops[key].Unlock()
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
