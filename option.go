package structopt

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cmj0121/logger"
	"golang.org/x/text/width"
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
)

// The type-hint of the option type, max to 4-chars
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
	}
	return
}

// The option of the StructOpt and used to process the input arguments
type Option struct {
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
	// The set of the value can be used, may empty.
	// choices []string
}

// Generate the option by the reflect.StructOption, pass from the StructOpt.parse
func NewOption(sfield reflect.StructField, value reflect.Value, log *logger.Log) (option *Option, err error) {
	option = &Option{
		Value:     value,
		StructTag: sfield.Tag,

		name:        strings.ToLower(sfield.Name),
		option_type: Ignore,
		type_hint:   TYPEHINT_NONE,
	}

	switch value.Kind() {
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
		// as the flag
		option.option_type = Flag
		option.type_hint = TYPEHINT_NONE
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

func (option *Option) Name() (name string) {
	if name, _ = option.Lookup(TAG_NAME); name == "" {
		// using the raw field name
		name = option.name
	}
	return
}

// [UTILITY] calculate the multi-char length
func WidecharSize(s string) (size int) {
	for _, r := range s {
		switch p := width.LookupRune(r); p.Kind() {
		case width.EastAsianWide:
			size += 2
		default:
			size += 1
		}
	}
	return
}
