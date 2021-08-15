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

// The option of flip
type FlipFlag struct {
	// The raw value of the input struct, should be the pointer of the value.
	reflect.Value

	// The field of the option in the struct
	reflect.StructTag

	// The callback function, may nil
	Callback

	// Name of the command-line, default is the name of struct.
	name string
	// The pre-defined value can used.
	choices []string
	// The default value
	default_value string
	// option is required
	required         bool
	option_type      Type
	option_type_hint TypeHint
}

func (option *FlipFlag) Name() (name string) {
	name = option.name
	if n := option.StructTag.Get(TAG_NAME); n != "" {
		// override the option name
		name = n
	}
	return
}

func (option *FlipFlag) ShortName() (name string) {
	name = option.StructTag.Get(TAG_SHORT)
	return
}

func (option *FlipFlag) String() (str string) {
	// show as the formatted option which has three parts: margin, option and help
	help, _ := option.Lookup(TAG_HELP)
	flag := ""
	flag_width := 24

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
		flag = fmt.Sprintf("%v [%v]", strings.ToUpper(option.Name()), option.TypeHint())
		flag_width = 12
	}

	flag_width_offset := WidecharSize(flag) - len([]rune(flag))
	str = fmt.Sprintf("    %-*v %v", flag_width-flag_width_offset, flag, help)

	if len(option.choices) > 0 {
		// show the choices
		str = fmt.Sprintf("%v %v", str, option.choices)
	}

	if option.default_value != "" {
		// has default value
		str = fmt.Sprintf("%v (default: %v)", str, option.default_value)
	}

	str = strings.TrimRight(str, " ")
	return
}

func (option *FlipFlag) Set(args ...string) (count int, err error) {
	value := option.Value
	for value.Kind() == reflect.Ptr {
		if value.IsZero() {
			// create dummy instance
			value.Set(reflect.New(value.Type().Elem()))
		}
		value = value.Elem()
	}

	switch option.Type() {
	case Flip:
		// flip the value
		value.SetBool(!value.Bool())
	case Flag, Argument:
		if len(args) == 0 {
			err = fmt.Errorf("%v should pass %v", option.Name(), option.TypeHint())
			return
		}
		arg := args[0]

		if len(option.choices) > 0 {
			idx := sort.SearchStrings(option.choices, arg)
			if idx == len(option.choices) || option.choices[idx] != arg {
				err = fmt.Errorf("set %v: %v not in %v", option.Name(), arg, option.choices)
				return
			}
		}

		switch option.TypeHint() {
		case INT:
			var val int64

			if val, err = AtoI(arg); err != nil {
				err = fmt.Errorf("pass %v: %v", arg, err)
				return
			}
			value.SetInt(val)
		case UINT:
			var val uint64

			if val, err = AtoU(arg); err != nil {
				err = fmt.Errorf("pass %#v as INT: %v", arg, err)
				return
			}
			value.SetUint(val)
		case STR:
			// just set the raw string
			value.Set(reflect.ValueOf(arg))
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
			err = fmt.Errorf("not implemented set %v", option.TypeHint())
			return
		}
		count++
	default:
		err = fmt.Errorf("should not be here: %v", option.Type())
		return
	}

	if option.Callback != nil {
		// call the callback
		log.Trace("execute callback %v", option.Callback)
		option.Callback(option)
	}
	return
}

// Show the option type
func (option *FlipFlag) Type() (typ Type) {
	typ = option.option_type
	return
}

// Show the type-hint
func (option *FlipFlag) TypeHint() (typ TypeHint) {
	typ = option.option_type_hint
	return
}

// Set callback fn
func (option *FlipFlag) SetCallback(fn Callback) {
	// override the callback
	option.Callback = fn
}

func (opt *FlipFlag) IsRequired() (required bool) {
	required = opt.required
	return
}
