package structopt

import (
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/width"
)

// pre-define meta
const (
	// the project name
	PROJ_NAME = "structopt"
	// the version info
	MAJOR = 0
	MINOR = 1
	MACRO = 0
)

// pre-define TAG key
const (
	TAG_IGNORE = "-"
	// reserved key of the field tag
	TAG_NAME     = "name"
	TAG_SHORT    = "short"
	TAG_HELP     = "help"
	TAG_CALLBACK = "callback"

	// special tag which no-need provide the valie
	TAG_OPTION     = "option"
	TAG_OPTION_SEP = ","
	// used to node the field allow data truncated
	TAG_TRUNC = "trunc"
)

// pre-define the INT/UINT format
var (
	RE_INT = regexp.MustCompile(`^0|[1-9][0-9]*$`)
	RE_BIN = regexp.MustCompile(`^(:?0[bB])([01]+)$`)
	RE_OCT = regexp.MustCompile(`^(:?0[oO]?)([0-7]+)$`)
	RE_HEX = regexp.MustCompile(`^(:?0[xX])([0-9a-fA-F]+)$`)

	RE_FLOAT = regexp.MustCompile(`^-?(?:0|[1-9][0-9]*)?\.[0-9]+$`)
	RE_RAT   = regexp.MustCompile(`^-?(?:0|[1-9][0-9]*)/-?[1-9][0-9]*$`)
)

// the strconv.Atoi wrapper for process the hexadecimal or other format
func AtoI(s string) (val int64, err error) {
	minus := false
	if len(s) > 0 && s[0] == '-' {
		minus = true
		s = s[1:]
	}

	switch {
	case RE_HEX.MatchString(s):
		if s = RE_HEX.FindStringSubmatch(s)[2]; minus {
			s = "-" + s
		}
		val, err = strconv.ParseInt(s, 16, 64)
	case RE_OCT.MatchString(s):
		if s = RE_OCT.FindStringSubmatch(s)[2]; minus {
			s = "-" + s
		}
		val, err = strconv.ParseInt(s, 8, 64)
	case RE_BIN.MatchString(s):
		if s = RE_BIN.FindStringSubmatch(s)[2]; minus {
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
func AtoU(s string) (val uint64, err error) {
	switch {
	case RE_HEX.MatchString(s):
		val, err = strconv.ParseUint(RE_HEX.FindStringSubmatch(s)[2], 16, 64)
	case RE_OCT.MatchString(s):
		val, err = strconv.ParseUint(RE_OCT.FindStringSubmatch(s)[2], 8, 64)
	case RE_BIN.MatchString(s):
		val, err = strconv.ParseUint(RE_BIN.FindStringSubmatch(s)[2], 2, 64)
	case RE_INT.MatchString(s):
		val, err = strconv.ParseUint(s, 10, 64)
	default:
		err = fmt.Errorf("not the sign INT: %v", s)
		return
	}
	return
}

func AtoF(s string) (val float64, err error) {
	switch {
	case RE_FLOAT.MatchString(s):
		val, err = strconv.ParseFloat(s, 64)
	case RE_RAT.MatchString(s):
		pattern := strings.Split(s, "/")
		var num int64
		var denom int64

		if num, err = AtoI(pattern[0]); err != nil {
			// invalid numerator
			return
		}
		if denom, err = AtoI(pattern[1]); err != nil {
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
