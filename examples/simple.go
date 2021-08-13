package main

import (
	"github.com/cmj0121/structopt"
)

type Example struct {
	structopt.Help
}

func main() {
	example := Example{}
	parser := structopt.MustNew(&example)
	parser.Run()
}
