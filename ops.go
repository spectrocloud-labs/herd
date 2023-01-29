package zeroinit

import "context"

type opCallback func(context.Context) error

type opState struct {
	fn         opCallback
	err        error
	fatal      bool
	background bool
}

func (o opState) toGraphEntry(name string) GraphEntry {
	return GraphEntry{
		WithCallback: o.fn != nil,
		Callback:     o.fn,
		Error:        o.err,
		Background:   o.background,
		Fatal:        o.fatal,
		Name:         name,
	}
}
