package main

import (
	"github.com/cmj0121/structotp"
)

type Example struct {
	Name string `short:"n" help:"Enter your name"`
	Age  int    `short:"年" name:"âge" help:"The utf-8 field`
}

func main() {
	example := Example{}
	parser := structopt.MustNew(&example)
	parser.Run()
}
