package structopt

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

// The struct parser as the argument options.
type StructOpt struct {
	// The raw value of the input struct, should be the pointer of the value.
	reflect.Value

	// Name of the command-line, default is the name of struct.
	name string

	// callback function when set
	Callback

	// The properties of the Option used in StructOpt.
	named_options map[string]Option
}

// Must generate the parse, or raise panic when failure.
func MustNew(in interface{}) (opt *StructOpt) {
	var err error

	opt, err = New(in)
	if err != nil {
		// raise the panic
		panic(err)
	}
	return
}

// Generate the parse by input struct, or return error message.
func New(in interface{}) (opt *StructOpt, err error) {
	value := reflect.ValueOf(in)

	log.Trace("StructOpt.New(%T)", in)
	switch {
	case value.Kind() != reflect.Ptr:
		err = fmt.Errorf("should pass the *Struct: %T", in)
		return
	case value.Elem().Kind() != reflect.Struct:
		err = fmt.Errorf("should pass the *Struct: %T", in)
		return
	case !value.IsValid():
		err = fmt.Errorf("should pass the *Struct: %T (invalid)", in)
		return
	}

	opt = &StructOpt{
		Value: value,

		name:          strings.ToLower(value.Elem().Type().Name()),
		named_options: map[string]Option{},
	}
	return
}

// Syntax-sugar for show help message
func (opt *StructOpt) Help(option Option) {
	os.Stderr.WriteString(opt.String())
	os.Exit(0)
}

// Run as default command-line parser, read from os.Args and show error and usage when parse error.
func (opt *StructOpt) Run() {
	if _, err := opt.Set(os.Args[1:]...); err != nil {
		// show the error message
		os.Stderr.WriteString(fmt.Sprintf("error: %v\n%v", err, opt.String()))
		// and then exit the program
		os.Exit(1)
	}
}
