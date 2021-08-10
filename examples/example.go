package main

import (
	"fmt"
	"net"
	"time"

	"github.com/cmj0121/structopt"
)

type Sub struct {
	structopt.Help

	Flip *bool `short:"f" option:"flag" help:"flip the value"`
	Name *string `short:"n" option:"flag" help:"set as name"`
	Age *uint `short:"a" option:"flag" help:"force set as flag"`
}

type Example struct {
	structopt.Help

	Ignore bool `-`
	Skip   bool `option:"skip"`

	Flip bool `short:"f" help:"flip the value"`
	Name string `short:"n" help:"set as name"`
	Age uint `short:"a" help:"force set as flag"`

	Now  time.Time  `short:"t" help:"type the RFC-3389 time format"`
	CIDR *net.IPNet `option:"flag" help:"please type the valid CIDR"`

	// treate as argument
	Argument *string `name:"arg" help:"required argument"`

	*Sub `help:"sub-command"`
}

func main() {
	example := Example{}
	parser := structopt.MustNew(&example)
	parser.Run()

	fmt.Printf("%#v\n", example)
}
