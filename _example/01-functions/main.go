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
