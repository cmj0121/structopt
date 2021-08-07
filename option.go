package structopt

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/cmj0121/logger"
)

// The callback function which is used when option been set
type Callback func(option *Option) error

// The enum type of the option
//go:generate stringer -type=OptionType
type OptionType int

const (
	// Ignore this option
	Ignore OptionType = iota
	// The flag of the option, only store true/false value.
	Flip
	// The value store and will auto-convert to fit type.
	Flag
	// The argument
	Argument
	// The extension of option which recursive process the pass arguments.
	Subcommand
)

// The type-hint of the option
//go:generate stringer -type=OptionTypeHint
type OptionTypeHint int

const (
	// no-need to provide the type hint
	NONE OptionTypeHint = iota
	// the sign integer, can be save as int64
	INT
	// the sign integer, can be save as uint64
	UINT
	// the sign rantional number
	RAT
	// the string value
	STR
	// the file-path, an will be auto-open
	FILE
	// the file-permission
	FMODE
	// the RFC-3389 format timestamp
	TIME
	// the time duration string
	SPAN
	// the network interface
	IFACE
	// the network IPv4 / IPv6 address
	IP
	// the network IPv4 / IPv6 address with mask, CIDR
	CIDR
)

// The option of the StructOpt and used to process the input arguments
type Option struct {
	*logger.Log

	// The related reflect.Value of the StructOpt, can set the value directly.
	reflect.Value
	// The save field tag
	reflect.StructTag
	// The callback function, trigger when set value
	Callback

	// The raw field name
	name string
	// The type hint of the reflect.Value
	type_hint OptionTypeHint
	// The type of the option.
	option_type OptionType
	// the customized options
	options map[string]struct{}
	// The set of the value can be used, may empty.
	// choices []string
}

// Prepare the option by the known field
func (option *Option) Prepare() (err error) {
	if val, ok := option.Lookup(TAG_OPTION); ok {
		for _, opt := range strings.Split(val, TAG_OPTION_SEP) {
			// set as key in map
			option.options[strings.TrimSpace(opt)] = struct{}{}
		}
	}

	switch option.Value.Kind() {
	case reflect.Bool:
		// should be flip
		option.option_type = Flip
		option.type_hint = NONE
	case reflect.Ptr:
		typ := option.Value.Type().Elem()
		// create the dummy value
		value := reflect.New(typ).Elem()

		// force to be option?
		if _, ok := option.options[TAG_FLAG]; ok {
			// force set as flag
			option.option_type = Flag
			err = option.prepare(value)
			return
		}

		// maybe the argument or sub-command
		switch typ.Kind() {
		case reflect.Struct:
			err = fmt.Errorf("not implemented: %v (%T)", typ, value.Interface())
			option.option_type = Subcommand
			return
		default:
			err = option.prepare(value)
			option.option_type = Argument
			return
		}
	default:
		// should be flag
		err = option.prepare(option.Value)
	}
	return
}

func (option *Option) prepare(value reflect.Value) (err error) {
	switch value.Interface().(type) {
	case os.File:
		// the flag / os.File
		option.option_type = Flag
		option.type_hint = FILE
	case os.FileMode:
		// the flag / os.FileMode
		option.option_type = Flag
		option.type_hint = FMODE
	case time.Time:
		// the flag / os.File
		option.option_type = Flag
		option.type_hint = TIME
	case time.Duration:
		// the flag / os.File
		option.option_type = Flag
		option.type_hint = SPAN
	case net.Interface:
		// the flag / net.Interface
		option.option_type = Flag
		option.type_hint = IFACE
	case net.IP:
		// the flag / net.IP
		option.option_type = Flag
		option.type_hint = IP
	case net.IPNet:
		// the flag / net.IPNet
		option.option_type = Flag
		option.type_hint = CIDR
	default:
		switch value.Kind() {
		case reflect.Bool:
			// the flip
			option.option_type = Flip
			option.type_hint = NONE
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			// the flag / sign-int
			option.option_type = Flag
			option.type_hint = INT
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			// the flag / sign-int
			option.option_type = Flag
			option.type_hint = UINT
		case reflect.Float32, reflect.Float64:
			// the flag / sign-rational number
			option.option_type = Flag
			option.type_hint = RAT
		case reflect.String:
			// the flag / string
			option.option_type = Flag
			option.type_hint = STR
		default:
			err = fmt.Errorf("prepare: unhandle field type %v (%T)", value.Kind(), value.Interface())
			return
		}
	}

	return
}

// Display the option in usage
func (option *Option) String() (str string) {
	// show as the formatted option which has three parts: margin, option and help
	help, _ := option.Lookup(TAG_HELP)
	flag := ""
	flag_width := 28

	type_hint := option.TypeHint().String()
	if option.TypeHint() == NONE {
		type_hint = ""
	}

	switch option.Type() {
	case Flip, Flag:
		short_name, _ := option.Lookup(TAG_SHORT)
		if len(short_name) > 0 {
			// add the leading -
			short_name = fmt.Sprintf("-%-v %v", short_name, type_hint)
			short_name = strings.TrimSpace(short_name)
		}
		short_width_offset := WidecharSize(short_name) - len([]rune(short_name))
		flag = fmt.Sprintf("%*v --%v %v", 8-short_width_offset, short_name, option.Name(), type_hint)
	default:
		flag = fmt.Sprintf("%v (%v)", strings.ToUpper(option.Name()), type_hint)
		flag_width = 18
	}

	flag_width_offset := WidecharSize(flag) - len([]rune(flag))
	str = fmt.Sprintf("    %-*v %v", flag_width-flag_width_offset, flag, help)
	str = strings.TrimRight(str, " ")
	return
}

// The option primary name
func (option *Option) Name() (name string) {
	if name, _ = option.Lookup(TAG_NAME); name == "" {
		// using the raw field name
		name = option.name
	}
	return
}

// Set the field from the pass value
func (option *Option) Set(args ...string) (count int, err error) {
	if !option.Value.CanSet() {
		err = fmt.Errorf("%v cannot set", option.Value)
		return
	}

	switch option.option_type {
	case Ignore:
		// need not operation and should not be called
		err = fmt.Errorf("OPTION %v (%v) should not call Set", option.Name(), option.type_hint)
	case Flip, Flag, Argument:
		value := option.Value

		for value.Kind() == reflect.Ptr {
			if value.IsZero() {
				option.Debug("%v is zero, create dummy", value.Type())
				value.Set(reflect.New(value.Type().Elem()))
			}

			value = value.Elem()
		}

		count, err = option.set(value, args...)
	default:
		// not implemented
		err = fmt.Errorf("OPTION %v (%v) not implemented Set", option.Name(), option.option_type)
		return
	}

	if option.Callback != nil {
		// run the callback
		err = option.Callback(option)
	}
	return
}

// the exactly set the value to the option. In the idea case the value should pass
// the reflect.Ptr and the option.set handle the arbitrary number of pointer
func (option *Option) set(value reflect.Value, args ...string) (count int, err error) {
	if option.TypeHint() == NONE {
		option.Trace("flip to %v", !value.Bool())
		value.SetBool(!value.Bool())
		return
	}

	if len(args) == 0 {
		err = fmt.Errorf("should pass %v", option.TypeHint())
		return
	}

	count = 1
	arg := args[0]
	switch option.TypeHint() {
	case STR:
		// set string as value
		value.Set(reflect.ValueOf(arg))
	case INT:
		var val int64
		if val, err = AtoI(arg); err != nil {
			// cannot encode as int64
			return
		}
		// set string as Int64
		value.SetInt(val)

		// check the value is overflow for the raw field type
		if _, ok := option.options[TAG_TRUNC]; !ok {
			// allow data can not allow be truncated
			if val != value.Int() {
				err = fmt.Errorf("overflow %v", arg)
				return
			}
		}
	case UINT:
		var val uint64
		if val, err = AtoU(arg); err != nil {
			// cannot encode as uint64
			return
		}
		// set string as Uint64
		value.SetUint(val)

		// check the value is overflow for the raw field type
		if _, ok := option.options[TAG_TRUNC]; !ok {
			// allow data can not allow be truncated
			if val != option.Value.Uint() {
				err = fmt.Errorf("overflow %v", arg)
				return
			}
		}
	case RAT:
		var val float64
		if val, err = AtoF(arg); err != nil {
			// cannot encode as float
			return
		}

		// set string as Float64
		value.SetFloat(val)
	case FILE:
		info, e := os.Stat(arg)
		switch {
		case os.IsNotExist(e):
			err = fmt.Errorf("file %#v does not exist", value)
			return
		case info.IsDir():
			err = fmt.Errorf("%#v is not file", value)
			return
		}

		fd, e := os.Open(arg)
		if e != nil {
			err = fmt.Errorf("cannot open file %#v: %v", value, e)
			return
		}

		value.Set(reflect.ValueOf(*fd))
	case FMODE:
		var val uint64

		val, err = AtoU(arg)
		if err != nil || val >= (1<<32) {
			err = fmt.Errorf("invalid file-mode: %v (%v)", value, err)
			return
		}

		filemode := os.FileMode(val)
		value.Set(reflect.ValueOf(filemode))
	case TIME:
		var timestamp time.Time
		if timestamp, err = time.Parse(time.RFC3339, arg); err != nil {
			err = fmt.Errorf("invalid time: %v (%v)", arg, err)
			return
		}
		value.Set(reflect.ValueOf(timestamp))
	case SPAN:
		var duration time.Duration

		if duration, err = time.ParseDuration(arg); err != nil {
			err = fmt.Errorf("invalid time duration: %v (%v)", arg, err)
			return
		}
		value.Set(reflect.ValueOf(duration))
	case IFACE:
		var iface *net.Interface
		iface, err = net.InterfaceByName(arg)
		if err != nil {
			err = fmt.Errorf("invalid IFace: %v", arg)
			return
		}
		value.Set(reflect.ValueOf(*iface))
	case IP:
		ip := net.ParseIP(arg)
		if ip == nil {
			// resoved by hostname
			var ips []net.IP

			ips, err = net.LookupIP(arg)
			if err != nil || len(ips) == 0 {
				err = fmt.Errorf("invalid IP: %v", arg)
				return
			}
			ip = ips[0]
		}

		value.Set(reflect.ValueOf(ip))
	case CIDR:
		var inet *net.IPNet

		// skip the IP field
		if _, inet, err = net.ParseCIDR(arg); err != nil {
			// err = fmt.Errorf("invalid CIDR: %v (%v)", value, err)
			return
		}
		value.Set(reflect.ValueOf(*inet))
	default:
		// not implemented
		err = fmt.Errorf("OPTION %v (%v) not implemented Set", option.Name(), option.type_hint)
		return
	}
	return
}

// The type of the option
func (option *Option) Type() (option_type OptionType) {
	option_type = option.option_type
	return
}

// The type-hint of the option
func (option *Option) TypeHint() (type_hint OptionTypeHint) {
	type_hint = option.type_hint
	return
}
