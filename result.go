package binder

import (
	"github.com/yuin/gopher-lua"
)

type Result struct {
	state   *lua.LState
	nValues int
	closed  bool
}

func (r *Result) Values() int {
	return r.nValues
}

func (r *Result) Get(num int) *Argument {
	return &Argument{
		state:  r.state,
		number: num,
	}
}

func (r *Result) Close() {
	if !r.closed {
		r.state.Pop(r.nValues)
	}
}
