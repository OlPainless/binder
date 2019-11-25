package binder

import (
	"errors"
	"testing"

	"github.com/yuin/gopher-lua"
)

func TestLua_Func(t *testing.T) {
	b := New(Options{
		SkipOpenLibs: true,
	})
	b.Func("negate", func(c *Context) error {
		t := c.Top()
		if t != 1 {
			return errors.New("need an argument")
		}

		c.Push().Bool(!c.Arg(1).Bool())
		return nil
	})

	l := NewLoader()
	l.Func("sum", func(c *Context) error {
		t := c.Top()
		if t < 2 {
			return errors.New("need at least 2 arguments")
		}

		var sum float64
		for i := 1; i <= t; i++ {
			sum += c.Arg(i).Number()
		}

		c.Push().Number(sum)
		return nil
	})

	l.Func("gettype", func(c *Context) error {
		t := c.Top()
		if t != 1 {
			return errors.New("need an argument")
		}

		switch c.Arg(1).Any().(type) {
		case lua.LNumber:
			c.Push().String("number")
		case lua.LString:
			c.Push().String("string")
		case lua.LBool:
			c.Push().String("bool")
		default:
			c.Push().Bool(false)
		}

		return nil
	})

	l.Func("raiser", func(c *Context) error {
		return errors.New("For honnor!")
	})

	b.Load(l)

	if _, err := b.DoString(`
		assert(negate(true) == false, 'wrong negation')
		assert(negate(false) == true, 'wrong negation')

		assert(gettype(123) == 'number', '123 is not number')
		assert(gettype("string") == 'string', '"string" is not string')
		assert(gettype(true) == 'bool', 'true is not bool')

		assert(sum(1, 2) == 3, '1 + 2 != 3')
		assert(sum(5, 7) == 12, '5 + 7 != 12')
		assert(sum(100, 200) == 300, '100 + 200 != 300')
	`); err != nil {
		t.Error("Error execute function", err)
	}

	if _, err := b.DoString("raiser()"); err == nil {
		t.Error("Must return error", err)
	}
}

func TestLua_Module(t *testing.T) {
	b := New()

	m := b.Module("reverse")

	m.String("_STRING", "foobar")
	m.Number("_NUMBER", 123)
	m.Bool("_BOOL", true)

	m.Func("string", func(c *Context) error {
		if c.Top() == 0 {
			return errors.New("need arguments")
		}

		s := c.Arg(1).String()

		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}

		c.Push().String(string(runes))
		return nil
	})

	if _, err := b.DoFile("lua_test.lua"); err != nil {
		t.Error("Error execute module", err)
	}
}

type Person struct {
	Name string
}

func TestLua_Table(t *testing.T) {
	b := New()

	tbl := b.Table("person")
	tbl.Static("new", func(c *Context) error {
		if c.Top() == 0 {
			return errors.New("need arguments")
		}
		n := c.Arg(1).String()

		c.Push().Data(&Person{n}, "person")
		return nil
	})

	tbl.Dynamic("name", func(c *Context) error {
		p, ok := c.Arg(1).Data().(*Person)
		if !ok {
			return errors.New("person expected")
		}

		if c.Top() == 1 {
			c.Push().String(p.Name)
		} else {
			p.Name = c.Arg(2).String()
		}

		return nil
	})

	if _, err := b.DoString(`
		local p = person.new('Steeve')

		assert(p:name() == 'Steeve', 'Steve is not Steve')

		p:name('Alice')
		assert(p:name() == 'Alice', 'Alice is not Alice')
	`); err != nil {
		t.Error("Error execute module", err)
	}
}

func TestLua_Caller(t *testing.T) {
	b := New()

	b.Func("test", func(c *Context) error {
		if c.Top() != 2 {
			return errors.New("expected two arguments")
		}

		a1 := c.Arg(1).String()
		if a1 != "Hello" {
			return errors.New("expected 'Hello' argument")
		}

		a2 := c.Arg(2).Number()
		if a2 != 42 {
			return errors.New("expected 42 argument")
		}

		c.Push().String(a1 + " World!")
		c.Push().Number(100)
		return nil
	})

	call := b.Call("test")
	call.Args().String("Hello")
	call.Args().Number(42)
	result, err := call.Execute()

	if err != nil {
		t.Error("Error execute call", err)
	}

	if result.Values() != 2 {
		t.Error("Expected 2 return values")
	}
	if r1 := result.Get(1).String(); r1 != "Hello World!" {
		t.Error("Expected 'Hello World!' as first return value. Was", r1)
	}
	if r2 := result.Get(2).Number(); r2 != 100 {
		t.Error("Expected 100 as second return value. Was", r2)
	}

	result.Close()
}

func TestLua_DoString(t *testing.T) {
	b := New()

	result, err := b.DoString(`
		local a = "Hello"
		local b = "World!"
		return a .. " " .. b
	`)
	if err != nil {
		t.Error("Error executing DoString", err)
	}
	if result.Values() != 1 {
		t.Error("Expected 1 return value")
	}
	if r1 := result.Get(1).String(); r1 != "Hello World!" {
		t.Error("Expected 'Hello World!' as first return value. Was", r1)
	}
	result.Close()
}
