package structopt

import (
	"net"
	"os"
	"testing"
	"time"
)

func TestInvalidInput(t *testing.T) {
	cases := []interface{}{
		nil,
		true,
		false,
		1,
		1.3,
		"test",
		'c',
		struct{}{},
	}

	for _, c := range cases {
		if _, err := New(c); err == nil {
			// expect failure
			t.Errorf("expect cannot parse %T", c)
		}
	}
}

type Sub struct {
	Help

	Rat float64 `help:"the rational or float number"`
}

type Foo struct {
	Help

	private bool `help:"should not exposed"`  // nolint

	Ignore1 bool `-` // nolint
	Ignore2 int  `option:"skip"`

	Name string `short:"n" help:"please type your name"`
	Age  uint   `short:"a" help:"please type your age"`

	Now  time.Time  `short:"t" help:"type the RFC-3389 time format"`
	CIDR *net.IPNet `option:"flag" help:"please type the valid CIDR"`

	*Sub `help:"the sub-command"`
}

func Example() {
	example := Foo{}
	parser := MustNew(&example)

	os.Stdout.WriteString(parser.Usage())
	// Output:
	// usage: foo [OPTION] [SUB]
	//
	// options:
	//           -h --help          show this message
	//       -n STR --name STR      please type your name
	//      -a UINT --age UINT      please type your age
	//      -t TIME --now TIME      type the RFC-3389 time format
	//              --cidr CIDR     please type the valid CIDR
	//
	// arguments:
	//     sub          the sub-command
}
