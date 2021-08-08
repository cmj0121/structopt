package structopt

import (
	"fmt"
)

// The display-name of the field
func (opt *StructOpt) Name() (name string) {
	name = opt.name
	return
}

// The short-name of the field
func (opt *StructOpt) ShortName() (name string) {
	// always be empty
	return
}

// Show the usage message
func (opt *StructOpt) String() (str string) {
	str = fmt.Sprintf("usage: %v\n", opt.Name())
	return
}

// Set the input argument and setup the value for secified fields, or return error.
func (opt *StructOpt) Set(args ...string) (idx int, err error) {
	disable_short_option := false
	disable_option := false

	for idx < len(args) {
		arg := args[idx]
		log.Trace("parse #%v argument: %#v", idx, arg)

		switch {
		case len(arg) == 0:
			// empty argument, skip
		case !disable_short_option && arg == "-":
			// disable short option
			log.Debug("#%v argument %#v: disable short option", idx, arg)
		case !disable_option && arg == "--":
			// disable option
			disable_short_option = true
			disable_option = true
			log.Debug("#%v argument %#v: disable option", idx, arg)
		case len(arg) > 1 && !disable_option && arg[:2] == "--":
			// long option
			log.Info("#%v argument %#v", idx, arg)
		case !disable_short_option && arg[:1] == "-":
			// short option
			log.Trace("#%v argument %#v", idx, arg)
			switch len([]rune(arg[1:])) {
			case 1:
				// single short option
				log.Info("#%v argument %#v: single short option", idx, arg)
			default:
				// multi- short options
				for short_opt_idx, short_opt := range arg[1:] {
					log.Info("#%v argument %#v: #%v short option: %#v", idx, arg, short_opt_idx, string(short_opt))
				}
			}
		default:
			// argument
			log.Info("#%v argument %#v", idx, arg)
		}

		idx++
	}

	return
}

// Show the type of the structopt, alwasy be Subcommand
func (opt *StructOpt) Type() (typ Type) {
	typ = Subcommand
	return
}

// Show the type hint of the structopt, alwasy be NONE
func (opt *StructOpt) TypeHint() (typ TypeHint) {
	typ = NONE
	return
}

// Set callback fn
func (opt *StructOpt) SetCallback(fn Callback) {
	// set callback function
	opt.Callback = fn
}
