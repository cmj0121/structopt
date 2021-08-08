package structopt

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"golang.org/x/text/width"
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
		// check is the simple sign int
		var sign_val int64

		if sign_val, err = AtoI(s); err != nil {
			err = fmt.Errorf("not the RAT: %v", s)
			return
		}

		val = float64(sign_val)
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
