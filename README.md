# Binder
High level go to Lua binder. Write less, do more.

**This is a fork of [github.com/alexeyco/binder](https://github.com/alexeyco/binder) - it has some improvements, but also backwards-incompatible changes. If you need stability, please use the original library**

Package binder allows to easily bind to Lua. Based on [gopher-lua](https://github.com/yuin/gopher-lua).

*Write less, do more!*

1. [Changes from original](#changes-from-original)
1. [Killer-feature](#killer-feature)
1. [Installation](#installation)
1. [Examples](#examples)
    1. [Functions](#functions)
    1. [Modules](#modules)
    1. [Tables](#tables)
    1. [Calling Lua functions from Go](#calling-lua-functions-from-go)
    1. [Options](#options)
    1. [Killer-featured errors](#killer-featured-errors)
1. [License](#license)

## Changes from original

- Fixes crashes with newer versions of _gopher-lua_: Now binder must be explicitly closed after use
- _Call_ functionality to call global Lua functions from Go
- _DoString_ and _DoFile_ now return a _Results_ struct in addition to error, so that return values from evaluated scripts can be retrieved (and cleaned from the stack)


## Killer-feature

You can display detailed information about the error and get something like this:

![Error](https://raw.githubusercontent.com/alexeyco/binder/master/Error.png)

See [_example/04-highlight-errors](https://github.com/olpainless/binder/tree/master/_example/04-highlight-errors). And [read more](#killer-featured-errors) about it.

## Installation
```
$ go get -u github.com/olpainless/binder
```
To run unit tests:
```
$ cd $GOPATH/src/github.com/olpainless/binder
$ go test -cover
```
To see why you need to bind go to lua (need few minutes):
```
$ cd $GOPATH/src/github.com/olpainless/binder
$ go test -bench=.
```

## Examples

### Functions
```go
package main

import (
	"errors"
	"log"

	"github.com/olpainless/binder"
)

func main() {
	b := binder.New(binder.Options{
		SkipOpenLibs: true,
	})
	defer b.Close()

	b.Func("log", func(c *binder.Context) error {
		t := c.Top()
		if t == 0 {
			return errors.New("need arguments")
		}

		l := []interface{}{}

		for i := 1; i <= t; i++ {
			l = append(l, c.Arg(i).Any())
		}

		log.Println(l...)
		return nil
	})

	if _, err := b.DoString(`
		log('This', 'is', 'Lua')
	`); err != nil {
		log.Fatalln(err)
	}
}
```

### Modules
```go
package main

import (
	"errors"
	"log"

	"github.com/olpainless/binder"
)

func main() {
	b := binder.New()

	m := b.Module("reverse")
	m.Func("string", func(c *binder.Context) error {
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

	if _, err := b.DoString(`
		local r = require('reverse')

		print(r.string('ABCDEFGHIJKLMNOPQRSTUFVWXYZ'))
	`); err != nil {
		log.Fatalln(err)
	}
}

```
### Tables
```go
package main

import (
	"errors"
	"log"

	"github.com/olpainless/binder"
)

type Person struct {
	Name string
}

func main() {
	b := binder.New()

	t := b.Table("person")
	t.Static("new", func(c *binder.Context) error {
		if c.Top() == 0 {
			return errors.New("need arguments")
		}
		n := c.Arg(1).String()

		c.Push().Data(&Person{n}, "person")
		return nil
	})

	t.Dynamic("name", func(c *binder.Context) error {
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
		print(p:name())

		p:name('Alice')
		print(p:name())
	`); err != nil {
		log.Fatalln(err)
	}
}
```

### Calling Lua functions from Go
```go
package main

import (
	"errors"
	"log"

	"github.com/olpainless/binder"
)

func main() {
	b := binder.New()

	b.Func("hello", func(c *binder.Context) error {
		if c.Top() == 0 {
			return errors.New("need arguments")
		}
		arg := c.Arg(1).String()

		c.Push().String("Hello " + arg)
		return nil
	})

	caller := b.Call("hello")
	caller.Args().String("World")
	result, err := caller.Execute()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(result.Get(1).String())
	result.Close()
}
``` 

### Options
```go
// Options binder options object
type Options struct {
	// CallStackSize is call stack size
	CallStackSize int
	// RegistrySize is data stack size
	RegistrySize int
	// SkipOpenLibs controls whether or not libraries are opened by default
	SkipOpenLibs bool
	// IncludeGoStackTrace tells whether a Go stacktrace should be included in a Lua stacktrace when panics occur.
	IncludeGoStackTrace bool
}
```
Read [more](https://github.com/yuin/gopher-lua#miscellaneous-luanewstate-options).

For example:
```go
b := binder.New(binder.Options{
	SkipOpenLibs: true,
})
```

### Killer-featured errors

```go
package main

import (
	"errors"
	"log"
	"os"

	"github.com/olpainless/binder"
)

type Person struct {
	Name string
}

func main() {
	b := binder.New()
	
	// ...

	if err := b.DoString(`-- some string`); err != nil {
		switch err.(type) {
		case *binder.Error:
			e := err.(*binder.Error)
			e.Print()

			os.Exit(0)
			break
		default:
			log.Fatalln(err)
		}
	}
}
```

Note: if `SkipOpenLibs` is `true`, not all open libs will be skipped in contrast to the basic logic of 
[gopher-lua](https://github.com/yuin/gopher-lua). If you set `SkipOpenLibs` to `true`, the following 
basic libraries will be loaded: all [basic functions](https://www.lua.org/manual/5.1/manual.html#5.1), 
[table](https://www.lua.org/manual/5.1/manual.html#5.5) and [package](https://www.lua.org/manual/5.1/manual.html#5.3).

## License
```
MIT License

Copyright (c) 2017 Alexey Popov

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
