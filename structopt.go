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

// Syntax-sugar for show help message
func (opt *StructOpt) Help(option *Option) (err error) {
	opt.Usage(nil)
	os.Exit(0)
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

	usage_bar := fmt.Sprintf("usage: %v", opt.Name)
	if len(opt.options) > 0 {
		// add the [OPTION]
		usage_bar = fmt.Sprintf("%v [OPTION]", usage_bar)
	}

	show_message(writer, usage_bar)
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

	idx := 0

	// the inner functin, which may increate the argument and auto-increate
	// index, or return error
	set_args := func(arg, key string) (err error) {
		// Set the value of the specified arg in the args, and rethrn the next index or error
		option, ok := opt.named_options[key]
		if !ok {
			// cannot found the option by name
			opt.Warn("cannot find %#v in %#v", key, opt.Name)
			err = fmt.Errorf("%#v not defined in %#v", arg, opt.Name)
			return
		}

		switch option.Type() {
		case Ignore:
		case Flip:
			if err = option.Set(""); err != nil {
				err = fmt.Errorf("cannot set %v: %v", option.Name(), err)
				return
			}
		case Flag:
			if idx += 1; idx >= len(args) {
				err = fmt.Errorf("%v should pass %v", option.Name(), option.TypeHint())
				return
			}
			if err = option.Set(args[idx]); err != nil {
				err = fmt.Errorf("cannot set %v: %v", option.Name(), err)
				return
			}
		default:
			err = fmt.Errorf("not implemented to set %v", option.Type())
			return
		}

		return
	}

	for idx < len(args) {
		arg := args[idx]
		opt.Trace("parse #%v argument: %#v", idx, arg)

		switch {
		case len(arg) == 0:
			// empty argument, skip
		case !disable_short_option && arg == "-":
			// disable short option
			opt.Debug("#%v argument %#v: disable short option", idx, arg)
		case !disable_option && arg == "--":
			// disable option
			disable_short_option = true
			disable_option = true
			opt.Debug("#%v argument %#v: disable option", idx, arg)
		case len(arg) > 1 && !disable_option && arg[:2] == "--":
			// long option
			opt.Info("#%v argument %#v: option: %#v", idx, arg, arg[2:])
			if err = set_args(arg, arg[2:]); err != nil {
				// cannot set args
				return
			}
		case !disable_short_option && arg[:1] == "-":
			// short option
			opt.Trace("#%v argument %#v: short option: %#v", idx, arg, arg[1:])
			switch len([]rune(arg[1:])) {
			case 1:
				// single short option
				opt.Info("#%v argument %#v: single short option: %#v", idx, arg, arg[1:])
				if err = set_args(arg, arg[1:]); err != nil {
					// cannot set args
					return
				}
			default:
				// multi- short options
				for short_opt_idx, short_opt := range arg[1:] {
					opt.Info("#%v argument %#v: #%v short option: %#v", idx, arg, short_opt_idx, string(short_opt))
					if err = set_args(arg, string(short_opt)); err != nil {
						// cannot set args
						return
					}
				}
			}
		default:
			// argument
			opt.Info("#%v argument %#v", idx, arg)
		}

		idx++
	}

	return
}

// Start parse the field of the struct, and raise error if not support field or wrong setting.
func (opt *StructOpt) prepare() (err error) {
	base_value := opt.Value.Elem()

	// iterate each field in the struct
	for idx := 0; idx < base_value.Type().NumField(); idx++ {
		field := base_value.Type().Field(idx)
		value := base_value.Field(idx)

		opt.Trace(
			"#%d field in %v: %-6v (%-8v canset: %v)",
			idx, base_value.Type(), field.Name, field.Type, value.CanSet(),
		)

		if !value.CanSet() {
			// ignore the field that cannot set
			opt.Debug("skip the cannot set field: %v", field.Name)
			continue
		}

		// process the field what we need
		opt.Trace("process field: %-6v (%v) `%v`", field.Name, field.Type, field.Tag)
		var option *Option

		switch {
		case field.Type.Kind() == reflect.Struct && field.Anonymous:
			for f_idx := 0; f_idx < field.Type.NumField(); f_idx++ {
				sub_field := field.Type.Field(f_idx)
				sub_value := value.Field(f_idx)

				opt.Trace("#%d sub-field in %v", f_idx, field.Name)
				if !sub_value.CanSet() {
					// cannot set the value, skip
					continue
				}

				if option, err = opt.add_option(sub_value, sub_field); err != nil {
					opt.Warn("set %v as option: %v", sub_field.Name, err)
					return
				} else if err = opt.add_callback(base_value, option); err != nil {
					opt.Warn("set %v as option: %v", sub_field.Name, err)
					return
				}
			}
		default:
			if option, err = opt.add_option(value, field); err != nil {
				opt.Warn("set %v as option: %v", field.Name, err)
				return
			} else if err = opt.add_callback(base_value, option); err != nil {
				opt.Warn("set %v as option: %v", field.Name, err)
				return
			}
		}
	}
	return
}

// add the option to the StructOpt
func (opt *StructOpt) add_option(value reflect.Value, sfield reflect.StructField) (option *Option, err error) {
	// setup the  option
	option = &Option{
		Log:       opt.Log,
		Value:     value,
		StructTag: sfield.Tag,

		name:        strings.ToLower(sfield.Name),
		option_type: Ignore,
		type_hint:   TYPEHINT_NONE,
		options:     map[string]struct{}{},
	}

	if strings.TrimSpace(string(sfield.Tag)) == TAG_IGNORE {
		// parse the field but skip
		opt.Info("skip the option: %v", option.Name())
		return
	}

	if err = option.Prepare(); err != nil {
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
	opt.Trace("set %#v as option", name)
	// the short-name option, if exist
	if name, ok := option.Lookup(TAG_SHORT); ok && name != "" {
		opt.Trace("set %#v as option", name)
		if _, ok := opt.named_options[name]; ok {
			err = fmt.Errorf("duplicated option name: %v", name)
			return
		}
		opt.named_options[name] = option
	}

	return
}

func (opt *StructOpt) add_callback(base_value reflect.Value, option *Option) (err error) {
	if fn_name, ok := option.Lookup(TAG_CALLBACK); ok {
		// always convert as the Title format
		fn_name = strings.Title(fn_name)
		opt.Trace("try add callback: %v", fn_name)

		// search the local callback
		if fn_value := base_value.MethodByName(fn_name); fn_value.IsValid() && !fn_value.IsZero() {
			opt.Debug("found possible local callback: %T", fn_value.Interface())
			// NOTE - using (func(*Option) error) instead of (Callback)
			if fn, ok := fn_value.Interface().(func(*Option) error); ok {
				// found the callback
				option.Callback = fn
				return
			}
		}

		// search the global callback
		if fn_value := reflect.ValueOf(opt).MethodByName(fn_name); fn_value.IsValid() && !fn_value.IsZero() {
			opt.Debug("found possible global callback: %T", fn_value.Interface())
			// NOTE - using (func(*Option) error) instead of (Callback)
			if fn, ok := fn_value.Interface().(func(*Option) error); ok {
				// found the callback
				option.Callback = fn
				return
			}
		}

		err = fmt.Errorf("cannot find the callback: %v", fn_name)
	}

	return
}
