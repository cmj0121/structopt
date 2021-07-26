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
	Flip bool `short:"f" help:"store true/false"`
}

func Example() {
	dummy := Dummy{}
	parser := MustNew(&dummy)
	parser.WriteUsage(os.Stdout, nil)
	// Output:
	// usage: dummy
	//
	// options:
	//     -f  --flip        store true/false
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
	//     -f  --flip        store true/false
}
