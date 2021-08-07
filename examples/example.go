package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/cmj0121/logger"
	"github.com/cmj0121/structopt"
)

// The example command-line
type Example struct {
	structopt.Help

	Skip        bool `-`
	Skip2       bool ` - `
	*logger.Log `-`

	Version bool `short:"v" callback:"ver" help:"show version info"`

	Flip bool   `short:"f" help:"flip the variable"`
	Name string `short:"n" help:"Enter your name"`
	Age  int    `short:"年" name:"âge" help:"The utf-8 field"`

	Price float64 `short:"F" help:"the float or rational number format"`

	*os.File     `option:"flag" help:"open file, default is Read-Only"`
	*os.FileMode `option:"flag" help:"oct-based file permission"`

	*time.Time     `option:"flag" help:"the timestamp of RFC-3339 format"`
	*time.Duration `option:"flag" help:"the human-readable time span"`

	*net.Interface `option:"flag" help:"network interface"`
	*net.IPNet     `option:"flag" help:"network address with mask, CIDR"`
	net.IP         `option:"flag" help:"the IPv4/IPv6 address"`
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
	case example.Time != nil:
		fmt.Printf("time: %v\n", example.Time)
	case example.Duration != nil:
		fmt.Printf("duration: %v\n", example.Duration)
	case example.Interface != nil:
		fmt.Printf("IFace: %v\n", example.Interface)
	case example.IPNet != nil:
		fmt.Printf("IPNet: %v\n", example.IPNet)
	case example.IP != nil:
		fmt.Printf("IP: %v\n", example.IP)
	default:
		fmt.Printf("%#v\n", example)
	}
}
