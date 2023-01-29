package zeroinit

import (
	"context"
	"sync"
)

type opCallback func(context.Context) error

type opState struct {
	sync.Mutex
	fn         opCallback
	err        error
	fatal      bool
	background bool
	weak       bool
}

func (o *opState) toGraphEntry(name string) GraphEntry {
	return GraphEntry{
		WithCallback: o.fn != nil,
		Callback:     o.fn,
		Error:        o.err,
		Background:   o.background,
		WeakDeps:     o.weak,
		Fatal:        o.fatal,
		Name:         name,
	}
}
