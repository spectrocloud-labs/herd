package zeroinit

type GraphOption func(g *Graph)

var EnableInit GraphOption = func(g *Graph) {
	g.init = true
}
