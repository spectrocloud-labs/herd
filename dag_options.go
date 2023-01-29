package zeroinit

import "context"

type GraphOption func(string, *opState, *Graph) error

var FatalOp GraphOption = func(key string, os *opState, g *Graph) error {
	os.fatal = true
	return nil
}

var Background GraphOption = func(key string, os *opState, g *Graph) error {
	os.background = true
	return nil
}

var WeakDeps GraphOption = func(key string, os *opState, g *Graph) error {
	os.weak = true
	return nil
}

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

func WithCallback(fn func(context.Context) error) GraphOption {
	return func(s string, os *opState, g *Graph) error {
		os.fn = fn
		return nil
	}
}
