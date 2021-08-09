package main

import (
	"fmt"

	"github.com/cmj0121/structopt"
)

type Sub struct {
	structopt.Help
}

type Example struct {
	structopt.Help

	Ignore bool `-`
	Skip   bool `option:"skip"`

	// used to process as flip
	Flip bool `short:"f" help:"flip the value"`
	// used to process as flag
	Name string `short:"n" help:"set as name"`
	// force set as flag
	Age *uint `short:"a" option:"flag" help:"force set as flag"`

	// treate as argument
	Argument *string

	*Sub `help:"sub-command"`
}

func main() {
	example := Example{}
	parser := structopt.MustNew(&example)
	parser.Run()

	fmt.Printf("%#v\n", example)
}
