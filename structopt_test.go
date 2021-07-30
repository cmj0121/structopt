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
	Flip    bool    `short:"f" help:"store true/false"`
	Age     uint    `short:"a" help:"field with type hint"`
	Price   float32 `help:"the sign float number"`
	Unicode string  `short:"多" name:"ユニコード" help:"the UTF-8 unicode option"`

	// pre-define type
	*os.File    `help:"open file, default is Read-Only"`
	time.Time   `help:"the timestamp of RFC-3339 format"`
	os.FileMode `help:"oct-based file permission"`

	IFace  net.Interface `help:"network interface"`
	CIDR   net.IPNet     `help:"network address with mask, CIDR"`
	net.IP `help:"the IPv4/IPv6 address"`
}

func Example() {
	dummy := Dummy{}
	parser := MustNew(&dummy)
	parser.WriteUsage(os.Stdout, nil)
	// Output:
	// usage: dummy [OPTION]
	//
	// options:
	//     -f       --flip              store true/false
	//     -a  UINT --age UINT          field with type hint
	//              --price RAT         the sign float number
	//     -多 STR  --ユニコード STR    the UTF-8 unicode option
	//              --file FILE         open file, default is Read-Only
	//              --time TIME         the timestamp of RFC-3339 format
	//              --filemode FMODE    oct-based file permission
	//              --iface IFACE       network interface
	//              --cidr CIDR         network address with mask, CIDR
	//              --ip IP             the IPv4/IPv6 address
}

func ExampleT() {
	dummy := Dummy{}
	parser := MustNew(&dummy)
	parser.Name = "foo"

	parser.WriteUsage(os.Stdout, nil)
	// Output:
	// usage: foo [OPTION]
	//
	// options:
	//     -f       --flip              store true/false
	//     -a  UINT --age UINT          field with type hint
	//              --price RAT         the sign float number
	//     -多 STR  --ユニコード STR    the UTF-8 unicode option
	//              --file FILE         open file, default is Read-Only
	//              --time TIME         the timestamp of RFC-3339 format
	//              --filemode FMODE    oct-based file permission
	//              --iface IFACE       network interface
	//              --cidr CIDR         network address with mask, CIDR
	//              --ip IP             the IPv4/IPv6 address
}
