package structopt

import (
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
	Name string
	Age  int
}

func ExampleStructOpt() {
	dummy := Dummy{}
	parse := MustNew(&dummy)
	parse.Parse("-n", "name", "--age", "21") // nolint
	// Output:
}
