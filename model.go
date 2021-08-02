package structopt

type Help struct {
	// the build-in option to show the help message
	Help bool `short:"h" name:"help" callback:"help" help:"show this message"`
}
