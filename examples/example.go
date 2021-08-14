package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/cmj0121/logger"
	"github.com/cmj0121/structopt"
)

type Sub struct {
	structopt.Help

	Flip *bool   `short:"f" option:"flag" help:"flip the value"`
	Name *string `short:"n" option:"flag" help:"set as name"`
	Age  *uint   `short:"a" option:"flag" help:"force set as flag"`
}

type Example struct {
	structopt.Help

	*logger.Log `-`
	LogLevel    string `name:"log" choice:"warn info debug trac3" help:"set log level"`

	Version bool `short:"v" help:"show version info" callback:"ver"`

	private bool `help:"should not exposed"`

	Ignore bool `-`
	Skip   bool `option:"skip"`

	Flip bool   `short:"f" help:"flip the value"`
	Name string `short:"n" help:"set as name" default:"john"`
	Age  uint   `short:"a" help:"force set as flag"`

	Now  time.Time  `short:"t" help:"type the RFC-3389 time format"`
	CIDR *net.IPNet `option:"flag" help:"please type the valid CIDR"`

	// treate as argument
	Arg1 *int    `help:"required argument"`
	Arg2 *string `help:"required argument"`

	*Sub `help:"sub-command"`
}

func (example Example) Ver(option structopt.Option) {
	fmt.Println("example: v0.0.0")
	os.Exit(0)
}

func main() {
	example := Example{
		Name: "default",
		Age:  5566,
	}
	parser := structopt.MustNew(&example)
	parser.Run()

	encode, err := json.MarshalIndent(example, "", "    ")
	if err != nil {
		fmt.Printf("%#v\n", example)
	} else {
		fmt.Printf("%v\n", string(encode))
	}

	encode, err = json.MarshalIndent(example.Sub, "", "    ")
	if err != nil {
		fmt.Printf("%#v\n", example.Sub)
	} else {
		fmt.Printf("%v\n", string(encode))
	}
}
