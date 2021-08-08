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

	*Sub `help:"sub-command"`
}

func main() {
	example := Example{}
	parser := structopt.MustNew(&example)
	parser.Run()

	fmt.Printf("%#v\n", example)
}
