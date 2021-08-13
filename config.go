package structopt

import (
	"regexp"

	"github.com/cmj0121/logger"
)

// pre-define meta
const (
	// the project name
	PROJ_NAME = "structopt"
	// the version info
	MAJOR = 0
	MINOR = 2
	MACRO = 4
)

// pre-define TAG key
const (
	TAG_IGNORE = "-"
	// reserved key of the field tag
	TAG_NAME     = "name"
	TAG_SHORT    = "short"
	TAG_HELP     = "help"
	TAG_CALLBACK = "callback"
	TAG_CHOICE   = "choice"

	// special tag which no-need provide the valie
	TAG_OPTION     = "option"
	TAG_OPTION_SEP = ","
	// used to node the field allow data truncated
	TAG_SKIP  = "skip"
	TAG_FLAG  = "flag"
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

// The inner log sub-system, used for trace and warning log.
var log = logger.New(PROJ_NAME)
