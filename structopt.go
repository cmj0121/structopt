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

	// callback function when set
	Callback

	// Name of the command-line, default is the name of struct.
	name string
	// The properties of the Option used in StructOpt.
	named_options map[string]Option

	ff_options  []Option
	sub_options []Option
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

	// generate the options
	based := opt.Value.Elem()
	for idx := 0; idx < based.NumField(); idx++ {
		field := based.Type().Field(idx)
		value := opt.Value.Elem().Field(idx)
		log.Trace("process #%d field: %v (%v)", idx, field.Name, field.Type)

		switch {
		case field.Type.Kind() == reflect.Struct && field.Anonymous:
			for sub_idx := 0; sub_idx < field.Type.NumField(); sub_idx++ {
				sub_field := field.Type.Field(sub_idx)
				sub_value := value.Field(sub_idx)
				log.Trace("process #%d sub-field in %v: %v (%v)", sub_idx, field.Type, sub_field.Name, sub_field.Type)

				// add option
				if err = opt.new_option(based, sub_value, sub_field); err != nil {
					log.Warn("cannot set %v as option: %v", sub_field.Name, err)
					err = fmt.Errorf("cannot set %v as option: %v", sub_field.Name, err)
					return
				}
			}
		default:
			// add option
			if err = opt.new_option(based, value, field); err != nil {
				log.Warn("cannot set %v as option: %v", field.Name, err)
				err = fmt.Errorf("cannot set %v as option: %v", field.Name, err)
				return
			}
		}
	}
	return
}

func (opt *StructOpt) new_option(based reflect.Value, value reflect.Value, field reflect.StructField) (err error) {
	var option Option
	log.Info("process %v (%v) as option", field.Name, field.Type)

	tags := map[string]struct{}{}
	for _, tag := range strings.Split(field.Tag.Get(TAG_OPTION), TAG_OPTION_SEP) {
		// tag = strings.TrimSpace(tag)
		tags[tag] = struct{}{}
	}

	_, skip := tags[TAG_SKIP]

	switch {
	case TAG_IGNORE == strings.TrimSpace(string(field.Tag)):
		log.Debug("option %v set ignore", field.Name)
		return
	case skip:
		log.Debug("option %v set skip", field.Name)
		return
	default:
		switch field.Type.Kind() {
		case reflect.Bool:
			flip := &FlipFlag{
				Value:     value,
				StructTag: field.Tag,
				name:      field.Name,
			}
			option = flip
		case reflect.Ptr:
			// may sub-command or argument
			_, flag := tags[TAG_FLAG]
			switch {
			case flag:
				// force set as flag
				log.Crit("not implemented: %v (%v) as flag", field.Name, field.Type)
				return
			case field.Type.Elem().Kind() == reflect.Struct:
				if value.IsZero() {
					// create dummy instance, and not set back
					value = reflect.New(field.Type.Elem())
					log.Trace("create dummy instance from %v: %v", field.Type.Elem(), value)
				}

				var sub *StructOpt
				if sub, err = New(value.Interface()); err != nil {
					log.Warn("create sub-command from %v: %v", field.Type.Elem(), err)
					err = fmt.Errorf("create sub-command from %v: %v", field.Type.Elem(), err)
					return
				}

				if name := field.Tag.Get(TAG_NAME); name != "" {
					// override the name
					sub.name = name
				}
				option = sub
			default:
				log.Crit("not implemented: %v (%v) as argument", field.Name, field.Type)
				return
			}
		default:
			log.Crit("not implemented: %v (%v) as flag", field.Name, field.Type)
			return
		}
	}

	// setup the callback
	if err = opt.set_callback(based, field.Tag.Get(TAG_CALLBACK), option); err != nil {
		err = fmt.Errorf("cannot set option %v: %v", option.Name(), err)
		return
	}

	switch option.Type() {
	case Flip, Subcommand:
		name := option.Name()
		if old, ok := opt.named_options[name]; ok {
			log.Warn("duplicated field: %v (%v)", name, old)
			err = fmt.Errorf("duplicated field: %v", name)
			return
		}
		opt.named_options[name] = option
		log.Info("add new named option: %v", name)
		switch option.Type() {
		case Flip:
			// add the option as sub-command
			opt.ff_options = append(opt.ff_options, option)
		case Subcommand:
			// add the option as sub-command
			opt.sub_options = append(opt.sub_options, option)
		}

		if name = option.ShortName(); name != "" {
			if old, ok := opt.named_options[name]; ok {
				log.Warn("duplicated field: %v (%v)", name, old)
				err = fmt.Errorf("duplicated field: %v", name)
				return
			}
			opt.named_options[name] = option
			log.Info("add new named option: %v", name)
		}
	default:
		log.Warn("not implemented set option: %v", option.Type())
		err = fmt.Errorf("not implemented set option: %v", option.Type())
		return
	}

	return
}

func (opt *StructOpt) set_callback(based reflect.Value, fn string, option Option) (err error) {
	if fn == "" {
		// no-need to process callback
		return
	}

	fn = strings.Title(fn)
	log.Trace("try set callback: %v", fn)
	local_fn := based.MethodByName(fn)
	if local_fn.IsValid() && !local_fn.IsZero() {
		if callback, ok := local_fn.Interface().(func(Option)); ok {
			log.Debug("set local callback: %v", callback)
			option.SetCallback(callback)
			return
		}
	}

	global_fn := reflect.ValueOf(opt).MethodByName(fn)
	if global_fn.IsValid() && !global_fn.IsZero() {
		if callback, ok := global_fn.Interface().(func(Option)); ok {
			log.Debug("set global callback: %v", callback)
			option.SetCallback(callback)
			return
		}
	}

	err = fmt.Errorf("cannot found callback: %v", fn)
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
