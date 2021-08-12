package structopt

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"
)

// The struct parser as the argument options.
type StructOpt struct {
	// The raw value of the input struct, should be the pointer of the value.
	reflect.Value
	// the reference instance of the parent StructOpt
	ref reflect.Value

	// callback function when set
	Callback

	// Name of the command-line, default is the name of struct.
	name string
	// The help message
	help string
	// The properties of the Option used in StructOpt.
	named_options map[string]Option

	ff_options  []Option
	arg_options []Option
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
		case !value.CanSet():
			// field cannot set, skip
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
	log.Debug("process %v (%v) as option", field.Name, field.Type)

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
			var flip *FlipFlag
			if flip, err = opt.new_flip_flag_arg(value, field); err != nil {
				// cannot create ff option
				err = fmt.Errorf("cannot create option %v: %v", field.Name, err)
				return
			}
			option = flip
		case reflect.Ptr:
			// may sub-command or argument
			_, flag := tags[TAG_FLAG]
			switch {
			case flag:
				// force set as flag
				if option, err = opt.new_flip_flag_arg(value, field); err != nil {
					// cannot create ff option
					err = fmt.Errorf("cannot create option %v: %v", field.Name, err)
					return
				}
			case field.Type.Elem().Kind() == reflect.Struct:
				ref := value
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
				sub.ref = ref
				sub.help = field.Tag.Get(TAG_HELP)
				option = sub
			default:
				var args *FlipFlag

				if args, err = opt.new_flip_flag_arg(value, field); err != nil {
					// cannot create ff option
					err = fmt.Errorf("cannot create option %v: %v", field.Name, err)
					return
				}
				args.option_type = Argument
				option = args
			}
		default:
			if option, err = opt.new_flip_flag_arg(value, field); err != nil {
				// cannot create ff option
				err = fmt.Errorf("cannot create option %v: %v", field.Name, err)
				return
			}
		}
	}

	// setup the callback
	if err = opt.set_callback(based, field.Tag.Get(TAG_CALLBACK), option); err != nil {
		err = fmt.Errorf("cannot set option %v: %v", option.Name(), err)
		return
	}

	name := option.Name()
	if old, ok := opt.named_options[name]; ok {
		log.Warn("duplicated field: %v (%v)", name, old)
		err = fmt.Errorf("duplicated field: %v", name)
		return
	}
	opt.named_options[name] = option
	switch option.Type() {
	case Ignore:
		// ignore the option
	case Flip, Flag:
		// add the option as sub-command
		opt.ff_options = append(opt.ff_options, option)
		log.Info("add new named option: --%v", name)
	case Argument:
		// add the option as argument
		opt.arg_options = append(opt.arg_options, option)
		log.Info("add new argument: %v", name)
	case Subcommand:
		// add the option as sub-command
		opt.sub_options = append(opt.sub_options, option)
		log.Info("add new sub-command: %v", name)
	default:
		log.Warn("not implemented set option: %v", option.Type())
		err = fmt.Errorf("not implemented set option: %v", option.Type())
		return
	}

	if name = option.ShortName(); name != "" {
		if old, ok := opt.named_options[name]; ok {
			log.Warn("duplicated field: %v (%v)", name, old)
			err = fmt.Errorf("duplicated field: %v", name)
			return
		}
		opt.named_options[name] = option
		log.Debug("add new short named option: %v", name)
	}

	return
}

func (opt *StructOpt) new_flip_flag_arg(value reflect.Value, field reflect.StructField) (option *FlipFlag, err error) {
	elm := value
	typ := field.Type
	log.Trace("try create option from %v (%v)", field.Name, value)

	for elm.Kind() == reflect.Ptr {
		switch {
		case elm.IsZero():
			elm = reflect.New(typ.Elem()).Elem()
		default:
			elm = elm.Elem()
		}
	}

	option = &FlipFlag{
		Value:     value,
		StructTag: field.Tag,

		name: strings.ToLower(field.Name),
	}
	if value.IsValid() && !value.IsZero() {
		// set the default value
		option.default_value = fmt.Sprintf("%v", value)
	}

	if val := option.StructTag.Get(TAG_CHOICE); val != "" {
		choices := strings.Split(val, " ")
		sort.Strings(choices)
		option.choices = choices
	}

	log.Debug("try create option %v: %T/%v", option.Name(), elm.Interface(), elm.Kind())

	switch elm.Interface().(type) {
	case os.File:
		// the flag / os.File
		option.option_type = Flag
		option.option_type_hint = FILE
	case os.FileMode:
		// the flag / os.FileMode
		option.option_type = Flag
		option.option_type_hint = FMODE
	case time.Time:
		// the flag / os.File
		option.option_type = Flag
		option.option_type_hint = TIME
	case time.Duration:
		// the flag / os.File
		option.option_type = Flag
		option.option_type_hint = SPAN
	case net.Interface:
		// the flag / net.Interface
		option.option_type = Flag
		option.option_type_hint = IFACE
	case net.IP:
		// the flag / net.IP
		option.option_type = Flag
		option.option_type_hint = IP
	case net.IPNet:
		// the flag / net.IPNet
		option.option_type = Flag
		option.option_type_hint = CIDR
	default:
		switch elm.Kind() {
		case reflect.Bool:
			option.option_type = Flip
			option.option_type_hint = NONE
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			option.option_type = Flag
			option.option_type_hint = INT
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			option.option_type = Flag
			option.option_type_hint = UINT
		case reflect.Float32, reflect.Float64:
			// the flag / sign-rational number
			option.option_type = Flag
			option.option_type_hint = RAT
		case reflect.String:
			option.option_type = Flag
			option.option_type_hint = STR
		default:
			log.Crit("not implemented: %v (%v) as flag", field.Name, typ)
			return
		}
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
	os.Stderr.WriteString(opt.Usage())
	os.Exit(0)
}

// Run as default command-line parser, read from os.Args and show error and usage when parse error.
func (opt *StructOpt) Run() {
	if _, err := opt.Set(os.Args[1:]...); err != nil {
		// show the error message
		os.Stderr.WriteString(fmt.Sprintf("error: %v\n%v", err, opt.Usage()))
		// and then exit the program
		os.Exit(1)
	}
}
