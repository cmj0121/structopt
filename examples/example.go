package main

import (
	"fmt"

	"github.com/cmj0121/structotp"
)

type Example struct {
	Flip bool   `short:"f" help:"flip the variable"`
	Name string `short:"n" help:"Enter your name"`
	Age  int    `short:"年" name:"âge" help:"The utf-8 field"`

	Price float64 `short:"F" help:"the float or rational number format"`
}

func main() {
	example := Example{}
	parser := structopt.MustNew(&example)
	parser.Run()

	fmt.Printf("#%v\n", example)
}
