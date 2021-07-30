package structopt

import (
	"golang.org/x/text/width"
)

// pre-define meta
const (
	// the project name
	PROJ_NAME = "structopt"
	// the version info
	MAJOR = 0
	MINOR = 0
	MACRO = 0
)

// pre-define TAG key
const (
	// reserved key of the field tag
	TAG_OPTION = "option"
	TAG_NAME   = "name"
	TAG_SHORT  = "short"
	TAG_HELP   = "help"
	TAG_IGNORE = "-"
)

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
