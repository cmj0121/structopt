package structopt

import (
	"fmt"
	"reflect"

	"github.com/cmj0121/logger"
)

// The struct parser as the argument options.
type StructOpt struct {
	// The raw value of the input struct, should be the pointer of the value.
	reflect.Value

	// The inner log sub-system, used for trace and warning log.
	*logger.Log
}

// Generate the parse by input struct, or return error message.
func New(in interface{}) (opt *StructOpt, err error) {
	value := reflect.ValueOf(in)

	if value.Kind() != reflect.Ptr || value.Elem().Kind() != reflect.Struct || !value.IsValid() {
		err = fmt.Errorf("should pass the *Struct: %T", in)
		return
	}

	opt = &StructOpt{
		Value: value,
		Log:   logger.New(value.Elem().Type().Name()),
	}
	return
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
