package internal

import (
	"fmt"
	"sort"

	"github.com/docopt/docopt-go"
)

const CURRENT_VERSION = "0.2.6"

const usage = `Goke

Usage:
  goke [<task>] [-w|--watch] [-c|--no-cache] [-f|--force] [-q|--quiet] [-a|--args=<a>...]
  goke -i | --init
  goke -h | --help
  goke -v | --version
  goke -t | --tasks [-c|--no-cache]

Options:
  -h --help      Show this screen
  -v --version   Show version
  -i --init      Creates a goke.yaml file in the current directory
  -t --tasks     Outputs a list of all task names
  -w --watch     Run task in watch mode
  -c --no-cache  Clears the program's cache
  -f --force     Runs the task even if files have not been changed
  -a --args=<a>  The arguments and options to pass to the underlying commands
  -q --quiet     Suppresses all output from tasks`

type OptionHandler struct {
	NeedsParser bool
	Func        func(*Parseable) (int, error)
}

func newOptionHandler(fn func(*Parseable) (int, error), needsParser bool) OptionHandler {
	return OptionHandler{
		NeedsParser: needsParser,
		Func:        fn,
	}
}

type Options struct {
	TaskName string   `docopt:"<task>"`
	Watch    bool     `docopt:"-w,--watch"`
	NoCache  bool     `docopt:"-c,--no-cache"`
	Force    bool     `docopt:"-f,--force"`
	Quiet    bool     `docopt:"-q,--quiet"`
	Args     []string `docopt:"-a,--args"`
	Init     bool     `docopt:"-i,--init"`
	Tasks    bool     `docopt:"-t,--tasks"`
}

func NewCliOptions() Options {
	var opts Options

	parsedDoc, _ := docopt.ParseArgs(usage, nil, CURRENT_VERSION)
	parsedDoc.Bind(&opts)

	return opts
}

// Handlers groups the handlers into a slice so that we can run them all at once when used.
func (opts Options) Handlers(p *Parseable) []OptionHandler {
	var handlers []OptionHandler

	handlers = append(handlers, newOptionHandler(opts.initHandler, false))
	handlers = append(handlers, newOptionHandler(opts.tasksHandler, true))

	return handlers
}

// initHandler creates goke.yaml if it doesn't exist.
// Can be invoked via the -i option.
func (opts Options) initHandler(p *Parseable) (int, error) {
	if !opts.Init {
		return -1, nil
	}

	err := CreateGokeConfig()
	if err != nil && !opts.Quiet {
		return 1, err
	}

	return 0, nil
}

// tasksHandler outputs a list of all tasks in the current goke.yaml file.
// Can be invoked via the -t option.
func (opts Options) tasksHandler(p *Parseable) (int, error) {
	if !opts.Tasks {
		return -1, nil
	}

	parser := (*p).(*parser)
	tasks := make([]string, 0, len(parser.Tasks))

	for k := range parser.Tasks {
		tasks = append(tasks, k)
	}
	sort.Strings(tasks)

	for _, task := range tasks {
		fmt.Println(task)
	}

	return 0, nil
}
