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
	if n := option.StructTag.Get(TAG_NAME); n != "" {
		// override the option name
		name = n
	}
	return
}

func (option *FlipFlag) ShortName() (name string) {
	name = option.StructTag.Get(TAG_SHORT)
	return
}
func (option *FlipFlag) String() (str string) {
	return
}
func (option *FlipFlag) Set(args ...string) (count int, err error) {
	if option.Callback != nil {
		// call the callback
		log.Trace("execute callback %v", option.Callback)
		option.Callback(option)
	}
	return
}

// Show the option type
func (option *FlipFlag) Type() (typ Type) {
	field_type := option.Value.Type()

	for field_type.Kind() == reflect.Ptr {
		// try to find the under struct
		field_type = field_type.Elem()
	}

	switch field_type.Kind() {
	case reflect.Bool:
		typ = Flip
	default:
		typ = Flag
	}
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
