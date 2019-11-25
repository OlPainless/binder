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
