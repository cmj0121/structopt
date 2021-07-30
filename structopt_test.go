package structopt

import (
	"os"
	"testing"
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
}

func Example() {
	dummy := Dummy{}
	parser := MustNew(&dummy)
	parser.WriteUsage(os.Stdout, nil)
	// Output:
	// usage: dummy
	//
	// options:
	//     -f       --flip              store true/false
	//     -a  UINT --age UINT          field with type hint
	//              --price RAT         the sign float number
	//     -多 STR  --ユニコード STR    the UTF-8 unicode option
}

func ExampleT() {
	dummy := Dummy{}
	parser := MustNew(&dummy)
	parser.Name = "foo"

	parser.WriteUsage(os.Stdout, nil)
	// Output:
	// usage: foo
	//
	// options:
	//     -f       --flip              store true/false
	//     -a  UINT --age UINT          field with type hint
	//              --price RAT         the sign float number
	//     -多 STR  --ユニコード STR    the UTF-8 unicode option
}
