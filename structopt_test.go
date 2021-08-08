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

type Dummy struct {
	Help

	Flip bool `short:"f" help:"store true/false"`

	Age    uint  `short:"a" name:"âge" help:"The utf-8 field"`
	Amount int64 `short:"A" help:"the sign integer"`
	Base   int8  `short:"b" help:"check base" option:"trunc"`

	Price   float32 `help:"the sign float number"`
	Unicode string  `short:"多" name:"ユニコード" help:"the UTF-8 unicode option"`

	// pre-define type
	*os.File       `option:"flag" help:"open file, default is Read-Only"`
	*time.Time     `option:"flag" help:"the timestamp of RFC-3339 format"`
	*time.Duration `option:"flag" help:"the human-readable time duration"`
	*os.FileMode   `option:"flag" help:"oct-based file permission"`

	IFace   *net.Interface `option:"flag" help:"network interface"`
	CIDR    *net.IPNet     `option:"flag" help:"network address with mask, CIDR"`
	*net.IP `option:"flag" help:"the IPv4/IPv6 address"`

	ArgStr *string `help:"The string argument"`
}

func Example() {
	dummy := Dummy{}
	parser := MustNew(&dummy)
	parser.Name = "foo"

	parser.WriteUsage(os.Stdout, nil)
	// Output:
	// usage: foo [OPTION] ARGSTR
	//
	// options:
	//           -h --help              show this message
	//           -f --flip              store true/false
	//      -a UINT --âge UINT          The utf-8 field
	//       -A INT --amount INT        the sign integer
	//       -b INT --base INT          check base
	//              --price RAT         the sign float number
	//      -多 STR --ユニコード STR    the UTF-8 unicode option
	//              --file FILE         open file, default is Read-Only
	//              --time TIME         the timestamp of RFC-3339 format
	//              --duration SPAN     the human-readable time duration
	//              --filemode FMODE    oct-based file permission
	//              --iface IFACE       network interface
	//              --cidr CIDR         network address with mask, CIDR
	//              --ip IP             the IPv4/IPv6 address
	//
	// arguments:
	//     ARGSTR (STR)       The string argument
}
