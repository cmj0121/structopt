package structopt

import (
	"regexp"

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
	TAG_IGNORE = "-"
	// reserved key of the field tag
	TAG_NAME  = "name"
	TAG_SHORT = "short"
	TAG_HELP  = "help"

	// special tag which no-need provide the valie
	TAG_OPTION     = "option"
	TAG_OPTION_SEP = ","
	// used to node the field allow data truncated
	TAG_TRUNC = "trunc"
)

// pre-define the INT/UINT format
var (
	RE_INT = regexp.MustCompile(`0|[1-9][0-9]*`)
	RE_BIN = regexp.MustCompile(`(:?0[bB])[01]+`)
	RE_OCT = regexp.MustCompile(`(:?0[oO])[0-7]+`)
	RE_HEX = regexp.MustCompile(`(:?0[xX])[0-9a-fA-F]+`)

	RE_FLOAT = regexp.MustCompile(`-?(?:0|[1-9][0-9]*)?\.[0-9]+`)
	RE_RAT   = regexp.MustCompile(`-?(?:0|[1-9][0-9]*)/-?[1-9][0-9]*`)
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
