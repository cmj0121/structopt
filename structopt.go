package structopt

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/cmj0121/logger"
)

// The struct parser as the argument options.
type StructOpt struct {
	// Name of the command-line, default is the name of struct
	Name string

	// The raw value of the input struct, should be the pointer of the value.
	reflect.Value

	// The inner log sub-system, used for trace and warning log.
	*logger.Log

	options       []*Option
	named_options map[string]*Option
}

// Generate the parse by input struct, or return error message.
func New(in interface{}) (opt *StructOpt, err error) {
	value := reflect.ValueOf(in)

	if value.Kind() != reflect.Ptr || value.Elem().Kind() != reflect.Struct || !value.IsValid() {
		err = fmt.Errorf("should pass the *Struct: %T", in)
		return
	}

	opt = &StructOpt{
		Name:  strings.ToLower(value.Elem().Type().Name()),
		Value: value,
		Log:   logger.New(PROJ_NAME),

		options:       []*Option{},
		named_options: map[string]*Option{},
	}
	opt.Writer(os.Stderr)
	// opt.Level(logger.TRACE)

	err = opt.prepare()
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

// Show the usage message in the STDERR
func (opt *StructOpt) Usage(err error) {
	// show the message on STDERR
	opt.WriteUsage(os.Stderr, err)
}

// Write the usage message
func (opt *StructOpt) WriteUsage(writer io.Writer, err error) {
	show_message := func(writer io.Writer, format string, args ...interface{}) {
		msg := fmt.Sprintf(format+"\n", args...)
		// exact write the message via Writer
		if _, err := writer.Write([]byte(msg)); err != nil {
			// cannot write, show as error message
			opt.Error("cannot write to Writer: %v", err)
		}
	}

	if err != nil {
		// show the error message
		show_message(writer, "error: %v", err)
	}

	show_message(writer, "usage: %v", opt.Name)
	if len(opt.options) > 0 {
		show_message(writer, "")
		show_message(writer, "options:")
		for _, option := range opt.options {
			show_message(writer, "%v", option)
		}
	}
}

// Run as default command-line parser, read from os.Args and show error and usage when parse error.
func (opt *StructOpt) Run() {
	if err := opt.Parse(os.Args[1:]...); err != nil {
		// show the error message
		opt.Usage(err)
		// and then exit the program
		os.Exit(1)
	}
}

// Parse the input argument and setup the value for secified fields, or return error.
func (opt *StructOpt) Parse(args ...string) (err error) {
	disable_short_option := false
	disable_option := false

	for idx, arg := range args {
		opt.Trace("parse #%v argument: %#v", idx, arg)

		switch {
		case arg == "":
			// empty argument, skip
		case !disable_short_option && arg == "-":
			// disable short option
			opt.Trace("disable short option")
		case !disable_option && arg == "--":
			// disable option
			disable_short_option = true
			disable_option = true
			opt.Trace("disable option")
		case !disable_option && arg[:2] == "--":
			// long option
			opt.Info("option: %#v", arg[2:])
		case !disable_short_option && arg[:1] == "-":
			// short option
			opt.Info("short option: %#v", arg[1:])
			// single short option
			// multi- short options
		default:
			// argument
			opt.Info("argument: %#v", arg)
		}
	}
	return
}

// Start parse the field of the struct, and raise error if not support field or wrong setting.
func (opt *StructOpt) prepare() (err error) {
	value := opt.Value.Elem()
	typ := value.Type()

	// iterate each field in the struct
	for idx := 0; idx < typ.NumField(); idx++ {
		field := typ.Field(idx)

		// check the field can set or not
		v := value.Field(idx)
		opt.Trace(
			"#%d field in %T: %-6v (%v) %v",
			idx, typ, field.Name, field.Type, v.CanSet(),
		)

		if !v.CanSet() {
			// ignore the field that cannot set
			opt.Debug("skip the cannot set field: %v", field.Name)
			continue
		}

		// process the field what we need
		opt.Info("process field: %-6v (%v) `%v`", field.Name, field.Type, field.Tag)

		var option *Option
		if option, err = NewOption(field, v, opt.Log); err != nil {
			// invalid option
			return
		}
		// append to the option-list
		opt.options = append(opt.options, option)
		// the named option
		name := option.Name()
		if _, ok := opt.named_options[name]; ok {
			err = fmt.Errorf("duplicated option name: %v", name)
			return
		}
		opt.named_options[name] = option
		// the short-name option, if exist
		if name, ok := option.Lookup(TAG_SHORT); ok && name != "" {
			if _, ok := opt.named_options[name]; ok {
				err = fmt.Errorf("duplicated option name: %v", name)
				return
			}
		}
	}
	return
}
