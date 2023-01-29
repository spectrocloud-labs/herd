package zeroinit

type GraphOption func(g *Graph) error

var EnableInit GraphOption = func(g *Graph) error {
	g.init = true
	return nil
}
