package zeroinit

type GraphOption func(string, *opState, *Graph) error

var FatalOp GraphOption = func(key string, os *opState, g *Graph) error {
	os.fatal = true
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
