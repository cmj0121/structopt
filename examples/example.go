package main

import (
	"fmt"
	"os"

	"github.com/cmj0121/structopt"
)

type Example struct {
	structopt.Help

	Version bool `short:"v" callback:"ver" help:"show version info"`

	Flip bool   `short:"f" help:"flip the variable"`
	Name string `short:"n" help:"Enter your name"`
	Age  int    `short:"年" name:"âge" help:"The utf-8 field"`

	Price float64 `short:"F" help:"the float or rational number format"`

	*os.File     `help:"open file, default is Read-Only"`
	*os.FileMode `help:"oct-based file permission"`
}

func (example Example) Ver(option *structopt.Option) (err error) {
	fmt.Println("v0.0.0")
	os.Exit(0)
	return
}

func main() {
	example := Example{}
	parser := structopt.MustNew(&example)
	parser.Run()

	switch {
	case example.File != nil:
		fmt.Printf("open Read-Only file: %v\n", example.File.Name())
	case example.FileMode != nil:
		fmt.Printf("file mode: %v\n", example.FileMode)
	default:
		fmt.Printf("%#v\n", example)
	}
}
