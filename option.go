package structopt

import (
	"fmt"
	"math/big"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cmj0121/logger"
)

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

// Generate the option by the reflect.StructOption, pass from the StructOpt.parse
func NewOption(sfield reflect.StructField, value reflect.Value, log *logger.Log) (option *Option, err error) {
	option = &Option{
		Log:       log,
		Value:     value,
		StructTag: sfield.Tag,

		name:        strings.ToLower(sfield.Name),
		option_type: Ignore,
		type_hint:   TYPEHINT_NONE,
		options:     map[string]struct{}{},
	}

	if val, ok := option.Lookup(TAG_OPTION); ok {
		for _, opt := range strings.Split(val, TAG_OPTION_SEP) {
			// set as key in map
			option.options[strings.TrimSpace(opt)] = struct{}{}
		}
	}

	err = option.setValue(value)
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
		short_name_offset := len(short_name) - WidecharSize(short_name)
		short_name = fmt.Sprintf("-%-*v %v", 2-short_name_offset, short_name, type_hint)
	}
	short_width_offset := len(short_name) - WidecharSize(short_name)
	flag := fmt.Sprintf("%-*v --%v %v", 8-short_width_offset, short_name, option.Name(), type_hint)

	flag_width_offset := len(flag) - WidecharSize(flag)
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
			if val, err = option.AtoI(value); err == nil {
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
			if val, err = option.AtoU(value); err == nil {
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
			if val, err = option.AtoF(value); err == nil {
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

// the strconv.Atoi wrapper for process the hexadecimal or other format
func (option *Option) AtoI(s string) (val int64, err error) {
	minus := false
	if len(s) > 0 && s[0] == '-' {
		minus = true
		s = s[1:]
	}

	switch {
	case RE_HEX.MatchString(s):
		if s = s[2:]; minus {
			s = "-" + s
		}
		val, err = strconv.ParseInt(s, 16, 64)
	case RE_OCT.MatchString(s):
		if s = s[2:]; minus {
			s = "-" + s
		}
		val, err = strconv.ParseInt(s, 8, 64)
	case RE_BIN.MatchString(s):
		if s = s[2:]; minus {
			s = "-" + s
		}
		val, err = strconv.ParseInt(s, 2, 64)
	case RE_INT.MatchString(s):
		if minus {
			s = "-" + s
		}
		val, err = strconv.ParseInt(s, 10, 64)
	default:
		err = fmt.Errorf("not the sign INT: %v", s)
		return
	}

	return
}

// the strconv.Atoi wrapper for process the hexadecimal or other format
func (option *Option) AtoU(s string) (val uint64, err error) {
	switch {
	case RE_HEX.MatchString(s):
		val, err = strconv.ParseUint(s[2:], 16, 64)
	case RE_OCT.MatchString(s):
		val, err = strconv.ParseUint(s[2:], 8, 64)
	case RE_BIN.MatchString(s):
		val, err = strconv.ParseUint(s[2:], 2, 64)
	case RE_INT.MatchString(s):
		val, err = strconv.ParseUint(s, 10, 64)
	default:
		err = fmt.Errorf("not the sign INT: %v", s)
		return
	}
	return
}

func (option *Option) AtoF(s string) (val float64, err error) {
	switch {
	case RE_FLOAT.MatchString(s):
		val, err = strconv.ParseFloat(s, 64)
	case RE_RAT.MatchString(s):
		pattern := strings.Split(s, "/")
		var num int64
		var denom int64

		if num, err = option.AtoI(pattern[0]); err != nil {
			// invalid numerator
			return
		}
		if denom, err = option.AtoI(pattern[1]); err != nil {
			// invalid denominator
			return
		}

		rat := big.NewRat(num, denom)
		val, _ = rat.Float64()

	default:
		err = fmt.Errorf("not the RAT: %v", s)
		return
	}

	return
}

// The customized StructTag Lookup method, which key can be search if

// set the option type and type hint
func (option *Option) setValue(value reflect.Value) (err error) {
	switch value.Interface().(type) {
	case *os.File:
		// the flag / os.File
		option.option_type = Flag
		option.type_hint = TYPEHINT_FILE
	case *net.Interface:
		// the flag / os.File
		option.option_type = Flag
		option.type_hint = TYPEHINT_IFACE
	case os.FileMode:
		// the flag / os.FileMode
		option.option_type = Flag
		option.type_hint = TYPEHINT_FILE_MODE
	case time.Time:
		// the flag / os.File
		option.option_type = Flag
		option.type_hint = TYPEHINT_TIME
	case net.Interface:
		// the flag / net.Interface
		option.option_type = Flag
		option.type_hint = TYPEHINT_IFACE
	case net.IP:
		// the flag / net.IP
		option.option_type = Flag
		option.type_hint = TYPEHINT_IP
	case net.IPNet:
		// the flag / net.IPNet
		option.option_type = Flag
		option.type_hint = TYPEHINT_CIDR
	default:
		switch typ := value.Type(); typ.Kind() {
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
