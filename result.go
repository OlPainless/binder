package binder

import (
	"github.com/yuin/gopher-lua"
)

// Result provides return values from a function call
type Result struct {
	state   *lua.LState
	nValues int
	closed  bool
}

// Values returns number of return values from function call
func (r *Result) Values() int {
	return r.nValues
}

// Get returns the nth return value, starting from 1
func (r *Result) Get(num int) *Argument {
	return &Argument{
		state:  r.state,
		number: num,
	}
}

// Close closes the results and pops the stack
func (r *Result) Close() {
	if !r.closed {
		r.state.Pop(r.nValues)
		r.closed = true
	}
}
