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
	Flip bool `short:"f" help:"store true/false"`

	Age    uint  `short:"a" help:"field with type hint"`
	Amount int64 `short:"A" help:"the sign integer"`
	Base   int8  `short:"b" help:"check base" option:"trunc"`

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

func TestStructOpt(t *testing.T) {
	dummy := Dummy{}
	parse := MustNew(&dummy)

	if err := parse.Parse("-f"); err != nil || !dummy.Flip {
		t.Fatalf("expect flip is workable: %v (%v)", dummy.Flip, err)
	}

	if err := parse.Parse("-ff", "--flip"); err != nil || dummy.Flip {
		t.Fatalf("expect multi-flip is workable: %v (%v)", dummy.Flip, err)
	}

	if err := parse.Parse("--age", "12"); err != nil || dummy.Age != 12 {
		t.Fatalf("expect UINT is workable: %v (%v)", dummy.Age, err)
	}

	if err := parse.Parse("--age", "18446744073709551615"); err != nil || dummy.Age != 18446744073709551615 {
		t.Fatalf("expect UINT is workable: %v (%v)", dummy.Age, err)
	}

	if err := parse.Parse("-A", "12"); err != nil || dummy.Amount != 12 {
		t.Fatalf("expect INT is workable: %v (%v)", dummy.Amount, err)
	}

	if err := parse.Parse("-A", "-0"); err != nil || dummy.Amount != -0 {
		t.Fatalf("expect INT is workable: %v (%v)", dummy.Amount, err)
	}

	if err := parse.Parse("-A", "-123"); err != nil || dummy.Amount != -123 {
		t.Fatalf("expect INT is workable: %v (%v)", dummy.Amount, err)
	}

	if err := parse.Parse("-A", "9223372036854775807"); err != nil || dummy.Amount != 9223372036854775807 {
		t.Fatalf("expect INT is workable: %v (%v)", dummy.Amount, err)
	}

	if err := parse.Parse("-A", "-9223372036854775808"); err != nil || dummy.Amount != -9223372036854775808 {
		t.Fatalf("expect INT is workable: %v (%v)", dummy.Amount, err)
	}

	if err := parse.Parse("--base", "0x12"); err != nil || dummy.Base != 0x12 {
		t.Fatalf("expect INT is workable: %v (%v)", dummy.Base, err)
	}

	if err := parse.Parse("--base", "0b10"); err != nil || dummy.Base != 2 {
		t.Fatalf("expect INT is workable: %v (%v)", dummy.Base, err)
	}

	if err := parse.Parse("--base", "0o77"); err != nil || dummy.Base != 63 {
		t.Fatalf("expect INT is workable: %v (%v)", dummy.Base, err)
	}

	if err := parse.Parse("--base", "0x1234"); err != nil || dummy.Base != 0x34 {
		t.Fatalf("expect INT is workable: %v (%v)", dummy.Base, err)
	}

	if err := parse.Parse("--price", "1.23"); err != nil || dummy.Price != 1.23 {
		t.Fatalf("expect RAT is workable: %v (%v)", dummy.Price, err)
	}

	if err := parse.Parse("--price", "-5/4"); err != nil || dummy.Price != -1.25 {
		t.Fatalf("expect RAT is workable: %v (%v)", dummy.Price, err)
	}
}

func Example() {
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
	//     -A  INT  --amount INT        the sign integer
	//     -b  INT  --base INT          check base
	//              --price RAT         the sign float number
	//     -多 STR  --ユニコード STR    the UTF-8 unicode option
	//              --file FILE         open file, default is Read-Only
	//              --time TIME         the timestamp of RFC-3339 format
	//              --filemode FMODE    oct-based file permission
	//              --iface IFACE       network interface
	//              --cidr CIDR         network address with mask, CIDR
	//              --ip IP             the IPv4/IPv6 address
}
