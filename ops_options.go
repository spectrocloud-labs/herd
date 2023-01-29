package zeroinit

import "context"

type OpOption func(string, *opState, *Graph) error

var FatalOp OpOption = func(key string, os *opState, g *Graph) error {
	os.fatal = true
	return nil
}

var Background OpOption = func(key string, os *opState, g *Graph) error {
	os.background = true
	return nil
}

var WeakDeps OpOption = func(key string, os *opState, g *Graph) error {
	os.weak = true
	return nil
}

func WithDeps(deps ...string) OpOption {
	return func(key string, os *opState, g *Graph) error {
		for _, d := range deps {
			if err := g.Graph.DependOn(key, d); err != nil {
				return err
			}
		}
		return nil
	}
}

func WithCallback(fn func(context.Context) error) OpOption {
	return func(s string, os *opState, g *Graph) error {
		os.fn = fn
		return nil
	}
}
