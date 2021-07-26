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

// The option of the StructOpt and used to process the input arguments
type Option struct {
	// The related reflect.Value of the StructOpt, can set the value directly.
	reflect.Value
	// The type of the option.
	OptionType
	// The save field tag
	reflect.StructTag

	// The raw field name
	name string
	// The set of the value can be used, may empty.
	// choices []string
	// The processed tag key-value, which value may empty.
}

// Generate the option by the reflect.StructOption, pass from the StructOpt.parse
func NewOption(sfield reflect.StructField, value reflect.Value, log *logger.Log) (option *Option, err error) {
	option = &Option{
		Value:      value,
		OptionType: Ignore,
		StructTag:  sfield.Tag,

		name: strings.ToLower(sfield.Name),
	}

	switch value.Kind() {
	case reflect.Bool:
		// the flip
		option.OptionType = Flip
	default:
		// as the flag
		option.OptionType = Flag
	}

	return
}

// Display the option in usage
func (option *Option) String() (str string) {
	// show as the formatted option which has three parts: margin, option and help
	help, _ := option.Lookup(TAG_HELP)
	str = fmt.Sprintf("    -f  --flip        %v", help)
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
