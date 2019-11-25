package binder

import (
	"github.com/yuin/gopher-lua"
)

// Caller is a structure for creating calls to Lua global functions
type Caller struct {
	startStack int
	context    *Context
}

// newCaller creates a new Caller.
// fn is the function name
func newCaller(fn string, context *Context) Caller {
	startStack := context.state.GetTop()

	lfn := context.state.GetGlobal(fn)
	context.state.Push(lfn)
	context.increase()

	return Caller{
		startStack: startStack,
		context:    context,
	}
}

// Args returns a Push for arguments
func (c *Caller) Args() *Push {
	return &Push{
		context: c.context,
	}
}

// Execute executes the function with pushed parameters and returns a Result
func (c *Caller) Execute() (Result, error) {
	err := c.context.state.PCall(c.context.pushed-1, lua.MultRet, nil)
	if err != nil {
		return Result{}, err
	}

	numReturns := c.context.state.GetTop() - c.startStack
	return Result{
		state:   c.context.state,
		nValues: numReturns,
	}, nil
}
