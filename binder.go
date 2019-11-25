package binder

import (
	"io/ioutil"

	"github.com/yuin/gopher-lua"
)

// Handler is binder function handler
type Handler func(*Context) error

// Binder is a binder... that's all
type Binder struct {
	*Loader
	state   *lua.LState
	loaders []*Loader
}

// Close closes the state and frees resources
func (b *Binder) Close() {
	b.state.Close()
}

// Load apply Loader
func (b *Binder) Load(loader *Loader) {
	b.loaders = append(b.loaders, loader)
}

// DoString runs lua script string
func (b *Binder) DoString(s string) (Result, error) {
	b.load()
	return b.do(func() error {
		return b.state.DoString(s)
	}, func(problem int) *source {
		return newSource(s, problem)
	})
}

// DoFile runs lua script file
func (b *Binder) DoFile(f string) (Result, error) {
	b.load()
	return b.do(func() error {
		return b.state.DoFile(f)
	}, func(problem int) *source {
		s, _ := ioutil.ReadFile(f)
		return newSource(string(s), problem)
	})
}

// Call creates a call to a Lua global function
func (b *Binder) Call(fn string) Caller {
	b.load()
	return newCaller(fn, &Context{
		state: b.state,
	})
}

// do captures return values after execution, or returns improved errors if needed
func (b *Binder) do(fn func() error, h errSourceHandler) (Result, error) {
	base := b.state.GetTop()
	if err := fn(); err != nil {
		return Result{}, newError(err, h)
	}

	numReturns := b.state.GetTop() - base
	return Result{
		state:   b.state,
		nValues: numReturns,
	}, nil
}

// source returns lua source script
func (b *Binder) source() string {
	return b.state.String()
}

func (b *Binder) load() {
	loaders := append([]*Loader{b.Loader}, b.loaders...)

	for _, l := range loaders {
		l.load(b.state)
	}
}

// New returns new binder instance
func New(opts ...Options) *Binder {
	options := []lua.Options{}

	if len(opts) > 0 {
		o := opts[0]

		options = append(options, lua.Options{
			CallStackSize:       o.CallStackSize,
			RegistrySize:        o.RegistrySize,
			SkipOpenLibs:        o.SkipOpenLibs,
			IncludeGoStackTrace: o.IncludeGoStackTrace,
		})
	}

	s := lua.NewState(options...)

	if len(opts) > 0 && opts[0].SkipOpenLibs {
		type lib struct {
			Name string
			Func lua.LGFunction
		}

		libs := []lib{
			{lua.LoadLibName, lua.OpenPackage},
			{lua.BaseLibName, lua.OpenBase},
			{lua.TabLibName, lua.OpenTable},
		}

		for _, l := range libs {
			s.Push(s.NewFunction(l.Func))
			s.Push(lua.LString(l.Name))
			s.Call(1, 0)
		}
	}

	b := &Binder{
		state:   s,
		loaders: []*Loader{},
	}
	b.Loader = NewLoader()

	return b
}

// NewLoader returns new loader
func NewLoader() *Loader {
	return &Loader{
		funcs:   map[string]Handler{},
		modules: []*Module{},
		tables:  []*Table{},
	}
}

func exports(funcs map[string]Handler) map[string]lua.LGFunction {
	e := make(map[string]lua.LGFunction, len(funcs))

	for name, handler := range funcs {
		e[name] = handle(handler)
	}

	return e
}

func handle(handler Handler) lua.LGFunction {
	return func(state *lua.LState) int {
		c := &Context{
			state: state,
		}

		err := handler(c)
		if err != nil {
			c.error(err.Error())
			return 0
		}

		return c.pushed
	}
}
