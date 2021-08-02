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
type OptionType int

const (
	// Ignore this option
	Ignore OptionType = iota
	// The flag of the option, only store true/false value.
	Flip
	// The value store and will auto-convert to fit type.
	Flag
	// The extension of option which recursive process the pass arguments.
	Subcommand
)

// The type-hint of the option
type OptionTypeHint int

const (
	// no-need to provide the type hint
	TYPEHINT_NONE OptionTypeHint = iota
	// the sign integer, can be save as int64
	TYPEHINT_INT
	// the sign integer, can be save as uint64
	TYPEHINT_UINT
	// the sign rantional number
	TYPEHINT_RAT
	// the string value
	TYPEHINT_STR
	// the file-path, an will be auto-open
	TYPEHINT_FILE
	// the file-permission
	TYPEHINT_FILE_MODE
	// the RFC-3389 format timestamp
	TYPEHINT_TIME
	// the network interface
	TYPEHINT_IFACE
	// the network IPv4 / IPv6 address
	TYPEHINT_IP
	// the network IPv4 / IPv6 address with mask, CIDR
	TYPEHINT_CIDR
)

// The type-hint of the option type, max to 5-chars
func (hint OptionTypeHint) String() (str string) {
	switch hint {
	case TYPEHINT_INT:
		str = "INT"
	case TYPEHINT_UINT:
		str = "UINT"
	case TYPEHINT_RAT:
		str = "RAT"
	case TYPEHINT_STR:
		str = "STR"
	case TYPEHINT_FILE:
		str = "FILE"
	case TYPEHINT_FILE_MODE:
		str = "FMODE"
	case TYPEHINT_TIME:
		str = "TIME"
	case TYPEHINT_IFACE:
		str = "IFACE"
	case TYPEHINT_IP:
		str = "IP"
	case TYPEHINT_CIDR:
		str = "CIDR"
	}
	return
}

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

	switch option.Value.Interface().(type) {
	case *os.File:
		// the flag / os.File
		option.option_type = Flag
		option.type_hint = TYPEHINT_FILE
	case *os.FileMode:
		// the flag / os.FileMode
		option.option_type = Flag
		option.type_hint = TYPEHINT_FILE_MODE
	case *time.Time:
		// the flag / os.File
		option.option_type = Flag
		option.type_hint = TYPEHINT_TIME
	case *net.Interface:
		// the flag / net.Interface
		option.option_type = Flag
		option.type_hint = TYPEHINT_IFACE
	case *net.IP:
		// the flag / net.IP
		option.option_type = Flag
		option.type_hint = TYPEHINT_IP
	case *net.IPNet:
		// the flag / net.IPNet
		option.option_type = Flag
		option.type_hint = TYPEHINT_CIDR
	default:
		switch typ := option.Value.Type(); typ.Kind() {
		case reflect.Bool:
			// the flip
			option.option_type = Flip
			option.type_hint = TYPEHINT_NONE
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			// the flag / sign-int
			option.option_type = Flag
			option.type_hint = TYPEHINT_INT
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			// the flag / sign-int
			option.option_type = Flag
			option.type_hint = TYPEHINT_UINT
		case reflect.Float32, reflect.Float64:
			// the flag / sign-rational number
			option.option_type = Flag
			option.type_hint = TYPEHINT_RAT
		case reflect.String:
			// the flag / string
			option.option_type = Flag
			option.type_hint = TYPEHINT_STR
		default:
			option.Error("unhandle field type %v", typ)
			err = fmt.Errorf("unhandle field type %v", typ)
		}
	}

	return
}

// Display the option in usage
func (option *Option) String() (str string) {
	// show as the formatted option which has three parts: margin, option and help
	help, _ := option.Lookup(TAG_HELP)

	type_hint := option.type_hint.String()
	short_name, _ := option.Lookup(TAG_SHORT)
	if len(short_name) > 0 {
		// add the leading -
		short_name_offset := WidecharSize(short_name) - len([]rune(short_name))
		short_name = fmt.Sprintf("-%-*v %v", 2-short_name_offset, short_name, type_hint)
	}
	short_width_offset := WidecharSize(short_name) - len([]rune(short_name))
	flag := fmt.Sprintf("%-*v --%v %v", 8-short_width_offset, short_name, option.Name(), type_hint)

	flag_width_offset := WidecharSize(flag) - len([]rune(flag))
	str = fmt.Sprintf("    %-*v %v", 28-flag_width_offset, flag, help)
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
func (option *Option) Set(value string) (err error) {
	if !option.Value.CanSet() {
		err = fmt.Errorf("%v cannot set", value)
		return
	}

	switch option.option_type {
	case Ignore:
		// need not operation and should not be called
		err = fmt.Errorf("OPTION %v (%v) should not call Set", option.Name(), option.type_hint)
	case Flip:
		if value != "" {
			err = fmt.Errorf("set FLIP should not pass any value")
			return
		}

		option.Trace("flip to %v", !option.Value.Bool())
		option.Value.SetBool(!option.Value.Bool())
	case Flag:
		switch option.type_hint {
		case TYPEHINT_STR:
			// set string as value
			option.Value.SetString(value)
		case TYPEHINT_INT:
			var val int64
			if val, err = AtoI(value); err == nil {
				// set string as Int64
				option.Value.SetInt(val)
				// check the value is overflow for the raw field type
				if _, ok := option.options[TAG_TRUNC]; !ok {
					// allow data can not allow be truncated
					if val != option.Value.Int() {
						err = fmt.Errorf("overflow %v", value)
						return
					}
				}
			}
		case TYPEHINT_UINT:
			var val uint64
			if val, err = AtoU(value); err == nil {
				// set string as Int64
				option.Value.SetUint(val)
				// check the value is overflow for the raw field type
				if _, ok := option.options[TAG_TRUNC]; !ok {
					// allow data can not allow be truncated
					if val != option.Value.Uint() {
						err = fmt.Errorf("overflow %v", value)
						return
					}
				}
			}
		case TYPEHINT_RAT:
			var val float64
			if val, err = AtoF(value); err == nil {
				// set string as Float64
				option.Value.SetFloat(val)
			}
		default:
			// not implemented
			err = fmt.Errorf("OPTION %v (%v) not implemented Set", option.Name(), option.type_hint)
			return
		}
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
