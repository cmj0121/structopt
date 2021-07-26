package structopt

import (
	"reflect"
	"strings"

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

// The option of the StructOpt and used to process the input arguments
type Option struct {
	// The related reflect.Value of the StructOpt, can set the value directly.
	reflect.Value
	// The type of the option.
	OptionType
	// The help message, may empty.
	Help string
	// The default value, may empty.
	Default string

	// The set of the value can be used, may empty.
	// choices []string
	// The processed tag key-value, which value may empty.
	tags map[string]string
}

// Generate the option by the reflect.StructOption, pass from the StructOpt.parse
func NewOption(sfield reflect.StructField, value reflect.Value, log *logger.Log) (option *Option, err error) {
	tags := sep_tags(string(sfield.Tag))
	option = &Option{
		Value:      value,
		OptionType: Ignore,
		tags:       tags,
	}

	if name, ok := option.tags[TAG_NAME]; !ok || name == "" {
		// set the lower-case as the name of option
		option.tags[TAG_NAME] = strings.ToLower(sfield.Name)
	}

	return
}

// The name of the option
func (option *Option) Name() (name string) {
	name = option.tags[TAG_NAME]
	return
}

// The short-name of the option
func (option *Option) ShortName() (name string) {
	name = option.tags[TAG_SHORT]
	return
}

// Utility for separate the StructTag to named map
func sep_tags(tag string) (tags map[string]string) {
	tags = map[string]string{}
	// separate the tag into the named map
	for len(tag) > 0 {
		idx := 0

		// step 1 - trim left
		for idx < len(tag) && (tag[idx] == ' ' || tag[idx] == '\t') {
			// skip the leading empty space
			idx++
		}
		tag = tag[idx:]
		idx = 0

		// step 2 - find the key
		for idx < len(tag) && !(tag[idx] == ' ' || tag[idx] == '\t' || tag[idx] == ':') {
			// the valid token in key
			idx++
		}
		key, value := tag[:idx], ""
		idx, tag = 0, tag[idx:]

		// step 3 - find the value
		if len(tag) > 0 {
			if tag[0] == ':' {
				// set the value
				tag = tag[1:]

				switch tag[0] {
				case '"':
					// find the value within double-quote
					idx = 1
					for idx < len(tag) && tag[idx] != '"' {
						if tag[idx] == '\\' {
							idx++
						}
						idx++
					}

					if tag[idx] != '"' {
						// invalid pair, skip
						tag = tag[idx:]
						continue
					}
					value = tag[1:idx]
					tag = tag[idx+1:]
				default:
					// invalid value
				}
			}
		}

		// step 4 - set key-value pair
		tags[key] = value
	}
	return
}
