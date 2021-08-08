package structopt

// The phsudo-option used in the StructOpt which may flip/flag, arguments
// or sub-command.
type Option interface {
	Prepare() error
	// The name of the option which should be unique in StructOpt.
	Name() string
	// The help message show in StructOpt.
	String() string
	// Set the option by pass arguments, return number of arguments used
	// or return error.
	Set(...string) (int, error)
	// Show the option type
	Type() Type
	// Show the type-hint
	TypeHint() TypeHint
	// Set the callback function
	SetCallback(Callback)
	// Lookup the tag by key
	Lookup(string) (string, bool)
}

// The callback function which is used when option been set
type Callback func(option Option) error

// The enum type of the option
//go:generate stringer -type=Type
type Type int

const (
	// Ignore this option
	Ignore Type = iota
	// The flag of the option, only store true/false value.
	Flip
	// The value store and will auto-convert to fit type.
	Flag
	// The argument
	Argument
	// The extension of option which recursive process the pass arguments.
	Subcommand
)

// The type-hint of the option
//go:generate stringer -type=TypeHint
type TypeHint int

const (
	// no-need to provide the type hint
	NONE TypeHint = iota
	// the sign integer, can be save as int64
	INT
	// the sign integer, can be save as uint64
	UINT
	// the sign rantional number
	RAT
	// the string value
	STR
	// the file-path, an will be auto-open
	FILE
	// the file-permission
	FMODE
	// the RFC-3389 format timestamp
	TIME
	// the time duration string
	SPAN
	// the network interface
	IFACE
	// the network IPv4 / IPv6 address
	IP
	// the network IPv4 / IPv6 address with mask, CIDR
	CIDR
)
