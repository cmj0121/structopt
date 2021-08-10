package structopt

import (
	"fmt"
	"reflect"
	"strings"
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
		flag = fmt.Sprintf("%v", strings.ToUpper(option.Name()))
		flag_width = 12
	}

	flag_width_offset := WidecharSize(flag) - len([]rune(flag))
	str = fmt.Sprintf("    %-*v %v", flag_width-flag_width_offset, flag, help)
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
	case Flag:
		if len(args) == 0 {
			err = fmt.Errorf("%v should pass %v", option.Name(), option.TypeHint())
			return
		}

		arg := args[0]
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
		default:
			err = fmt.Errorf("not implemented set %v", option.TypeHint())
			return
		}
		count ++
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
