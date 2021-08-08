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

	options       []Option
	arg_cmd_opts  []Option
	named_options map[string]Option
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

		named_options: map[string]Option{},
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
func (opt *StructOpt) Help(option Option) (err error) {
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

	for _, arg := range opt.arg_cmd_opts {
		// add the ARGUMENT / SUB-COMMAND
		usage_bar = fmt.Sprintf("%v %v", usage_bar, strings.ToUpper(arg.Name()))
	}

	show_message(writer, usage_bar)
	if len(opt.options) > 0 {
		show_message(writer, "")
		show_message(writer, "options:")
		for _, option := range opt.options {
			show_message(writer, "%v", option)
		}
	}

	if len(opt.arg_cmd_opts) > 0 {
		show_message(writer, "")
		show_message(writer, "arguments:")

		for _, arg := range opt.arg_cmd_opts {
			show_message(writer, "%v", arg)
		}
	}
}

// Run as default command-line parser, read from os.Args and show error and usage when parse error.
func (opt *StructOpt) Run() {
	if err := opt.Set(os.Args[1:]...); err != nil {
		// show the error message
		opt.Usage(err)
		// and then exit the program
		os.Exit(1)
	}
}

// Set the input argument and setup the value for secified fields, or return error.
func (opt *StructOpt) Set(args ...string) (err error) {
	disable_short_option := false
	disable_option := false

	idx, arg_idx := 0, 0

	// the inner functin, which may increate the argument and auto-increate
	// index, or return error
	set_args := func(key string, args ...string) (count int, err error) {
		// Set the value of the specified arg in the args, and rethrn the next index or error
		option, ok := opt.named_options[key]
		if !ok {
			// cannot found the option by name
			opt.Warn("cannot find %#v in %#v", key, opt.Name)
			err = fmt.Errorf("%#v not defined in %#v", key, opt.Name)
			return
		}

		switch option.Type() {
		case Ignore:
		case Flip, Flag:
			if count, err = option.Set(args...); err != nil {
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
		var count int

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
			opt.Info("#%v argument %#v", idx, arg)
			if count, err = set_args(arg[2:], args[idx+1:]...); err != nil {
				// cannot set args
				return
			}

			idx += count
		case !disable_short_option && arg[:1] == "-":
			// short option
			opt.Trace("#%v argument %#v", idx, arg)
			switch len([]rune(arg[1:])) {
			case 1:
				// single short option
				opt.Info("#%v argument %#v: single short option", idx, arg)
				if count, err = set_args(arg[1:], args[idx+1:]...); err != nil {
					// cannot set args
					return
				}
				idx += count
			default:
				// multi- short options
				for short_opt_idx, short_opt := range arg[1:] {
					opt.Info("#%v argument %#v: #%v short option: %#v", idx, arg, short_opt_idx, string(short_opt))
					if _, err = set_args(string(short_opt)); err != nil {
						// cannot set args
						return
					}
				}
			}
		default:
			// argument
			if arg_idx >= len(opt.arg_cmd_opts) {
				// too-many argument
				err = fmt.Errorf("too many argument: %v", arg)
				return
			}

			if count, err = opt.arg_cmd_opts[arg_idx].Set(args[idx:]...); err != nil {
				// cannot set args
				return
			}
			opt.Info("#%v argument %#v", idx, arg)
			arg_idx++
			idx += count
		}

		idx++
	}

	return
}

// Start parse the field of the struct, and raise error if not support field or wrong setting.
func (opt *StructOpt) prepare() (err error) {
	based := opt.Value.Elem()

	for option := range opt.GetOption() {
		if err = option.Prepare(); err != nil {
			opt.Warn("set cannot as option: %v", option.Name(), err)
			return
		}

		switch option.Type() {
		case Flip, Flag:
			// append to the option-list
			opt.options = append(opt.options, option)
			// the named option
			name := option.Name()
			if _, ok := opt.named_options[name]; ok {
				err = fmt.Errorf("duplicated option name: %v", name)
				return
			}
			opt.named_options[name] = option
			opt.Trace("set %#v as %v", name, option.Type())

			// the short-name option, if exist
			if name, ok := option.Lookup(TAG_SHORT); ok && name != "" {
				opt.Trace("set %#v as option", name)
				if _, ok := opt.named_options[name]; ok {
					err = fmt.Errorf("duplicated option name: %v", name)
					return
				}
				opt.named_options[name] = option
			}
		default:
			// append to the option-list
			opt.arg_cmd_opts = append(opt.arg_cmd_opts, option)
		}

		// add the callback function if need
		if fn_name, ok := option.Lookup(TAG_CALLBACK); ok {
			// always convert as the Title format
			fn_name = strings.Title(fn_name)
			opt.Trace("try add callback: %v", fn_name)

			// search the local callback
			if fn_value := based.MethodByName(fn_name); fn_value.IsValid() && !fn_value.IsZero() {
				opt.Debug("found possible local callback: %T", fn_value.Interface())
				// NOTE - using (func(Option) error) instead of (Callback)
				if fn, ok := fn_value.Interface().(func(Option) error); ok {
					// found the callback
					option.SetCallback(fn)
					continue
				}
			}

			// search the global callback
			if fn_value := reflect.ValueOf(opt).MethodByName(fn_name); fn_value.IsValid() && !fn_value.IsZero() {
				opt.Debug("found possible global callback: %T", fn_value.Interface())
				// NOTE - using (func(Option) error) instead of (Callback)
				if fn, ok := fn_value.Interface().(func(Option) error); ok {
					// found the callback
					option.SetCallback(fn)
					continue
				}
			}

			err = fmt.Errorf("cannot find the callback: %v", fn_name)
			return
		}
	}
	return
}

// generate new option by current struct field
func (opt *StructOpt) GetOption() (ch <-chan Option) {
	tmp := make(chan Option, 1)

	go func() {
		defer close(tmp)

		based := opt.Value.Elem()
		for idx := 0; idx < based.Type().NumField(); idx++ {
			field := based.Type().Field(idx)
			value := based.Field(idx)

			opt.Trace(
				"#%d field in %v: %-6v (%-8v canset: %v)",
				idx, based.Type(), field.Name, field.Type, value.CanSet(),
			)

			if !value.CanSet() {
				// ignore the field that cannot set
				opt.Debug("skip the cannot set field: %v", field.Name)
				continue
			} else if strings.TrimSpace(string(field.Tag)) == TAG_IGNORE {
				// parse the field but skip
				opt.Info("skip the field: %v", field.Name)
				continue
			}

			// process the field what we need
			opt.Debug("process field: %-6v (%v) `%v`", field.Name, field.Type, field.Tag)
			switch {
			case field.Type.Kind() == reflect.Struct && field.Anonymous:
				// the nested struct
				for f_idx := 0; f_idx < field.Type.NumField(); f_idx++ {
					sub_field := field.Type.Field(f_idx)
					sub_value := value.Field(f_idx)

					opt.Trace("#%d sub-field in %v: %v", f_idx, field.Name, sub_field.Name)
					if !sub_value.CanSet() {
						// cannot set the value, skip
						continue
					}

					tmp <- &Raw{
						Log:       opt.Log,
						Value:     sub_value,
						StructTag: sub_field.Tag,

						name:        strings.ToLower(sub_field.Name),
						option_type: Ignore,
						type_hint:   NONE,
						options:     map[string]struct{}{},
					}
				}
			default:
				tmp <- &Raw{
					Log:       opt.Log,
					Value:     value,
					StructTag: field.Tag,

					name:        strings.ToLower(field.Name),
					option_type: Ignore,
					type_hint:   NONE,
					options:     map[string]struct{}{},
				}
			}
		}
	}()

	ch = tmp
	return
}
