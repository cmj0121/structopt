package structopt

import (
	"reflect"
)

// The option of flip
type FlipFlag struct {
	// The raw value of the input struct, should be the pointer of the value.
	reflect.Value

	// The field of the option in the struct
	reflect.StructTag

	// Name of the command-line, default is the name of struct.
	name string

	// The callback function, may nil
	Callback
}

func (option *FlipFlag) Name() (name string) {
	name = option.name
	return
}

func (option *FlipFlag) ShortName() (name string) {
	return
}
func (option *FlipFlag) String() (str string) {
	return
}
func (option *FlipFlag) Set(args ...string) (count int, err error) {
	return
}

// Show the option type
func (option *FlipFlag) Type() (typ Type) {
	return
}

// Show the type-hint
func (option *FlipFlag) TypeHint() (typ TypeHint) {
	return
}

// Set callback fn
func (option *FlipFlag) SetCallback(fn Callback) {
	// override the callback
	option.Callback = fn
}
