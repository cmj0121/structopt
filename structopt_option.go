package structopt

import (
	"fmt"
	"strings"
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
func (opt *StructOpt) Usage() (str string) {
	var help_message []string

	usage := fmt.Sprintf("usage: %v", opt.Name())
	if len(opt.ff_options) > 0 {
		// add option
		usage = fmt.Sprintf("%v [OPTION]", usage)
	}

	for _, option := range opt.arg_options {
		// add argument
		usage = fmt.Sprintf("%v %v", usage, strings.ToUpper(option.Name()))
	}

	if len(opt.sub_options) > 0 {
		// add option
		usage = fmt.Sprintf("%v [SUB]", usage)
	}

	help_message = append(help_message, usage)

	if len(opt.ff_options) > 0 {
		help_message = append(help_message, "")
		help_message = append(help_message, "options:")

		for _, option := range opt.ff_options {
			// add the option row
			help_message = append(help_message, option.String())
		}
	}

	if len(opt.sub_options) > 0 || len(opt.arg_options) > 0 {
		help_message = append(help_message, "")
		help_message = append(help_message, "arguments:")

		for _, option := range opt.arg_options {
			// add the option row
			help_message = append(help_message, option.String())
		}

		for _, option := range opt.sub_options {
			// add the option row
			help_message = append(help_message, option.String())
		}
	}

	str = fmt.Sprintf("%v\n", strings.Join(help_message, "\n"))
	return
}

func (opt *StructOpt) String() (str string) {
	str = fmt.Sprintf("    %-12v %v", opt.Name(), opt.help)
	str = strings.TrimRight(str, " ")
	return
}

// Set the input argument and setup the value for secified fields, or return error.
func (opt *StructOpt) Set(args ...string) (idx int, err error) {
	disable_short_option := false
	disable_option := false

	for idx < len(args) {
		var count int

		arg := args[idx]
		log.Info("parse #%v argument: %#v", idx, arg)

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
			log.Debug("#%v argument %#v", idx, arg)

			if option, ok := opt.named_options[arg[2:]]; !ok {
				err = fmt.Errorf("unknown option: %v", arg)
				return
			} else if count, err = option.Set(args[idx+1:]...); err != nil {
				// cannot set value
				return
			}
			idx += count
		case !disable_short_option && arg[:1] == "-":
			// short option
			log.Trace("#%v argument %#v", idx, arg)
			switch len([]rune(arg[1:])) {
			case 1:
				// single short option
				log.Debug("#%v argument %#v: single short option", idx, arg)

				if option, ok := opt.named_options[arg[1:]]; !ok {
					err = fmt.Errorf("unknown option: %v", arg)
					return
				} else if count, err = option.Set(args[idx+1:]...); err != nil {
					// cannot set value
					return
				}
				idx += count
			default:
				// multi- short options
				for short_opt_idx, short_opt := range arg[1:] {
					log.Debug("#%v argument %#v: #%v short option: %#v", idx, arg, short_opt_idx, string(short_opt))

					if option, ok := opt.named_options[arg[1:]]; !ok {
						err = fmt.Errorf("unknown option: %v", arg)
						return
					} else if count, err = option.Set(); err != nil {
						err = fmt.Errorf("set %v: %v", arg, err)
						return
					}
					idx += count
				}
			}
		default:
			// argument
			log.Debug("#%v argument %#v", idx, arg)
			// sub-command
			if option, ok := opt.named_options[arg]; !ok {
				err = fmt.Errorf("unknown argument: %v", arg)
				return
			} else if _, err = option.Set(args[idx+1:]...); err != nil {
				err = fmt.Errorf("set %v: %v", arg, err)
				return
			}
			// NOTE - in sub-command case, there are no remains args
			idx = len(args)
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
