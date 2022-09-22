package cli

import (
	"flag"

	"github.com/dugajean/goke/internal"
)

func GetOptions() internal.Options {
	var opts internal.Options

	flag.BoolVar(&opts.ClearCache, "no-cache", false, "Clear Goke's cache. Default: false")
	flag.BoolVar(&opts.Watch, "watch", false, "Goke remains on and watches the task's specified files for changes, then reruns the command. Default: false")
	flag.BoolVar(&opts.Force, "force", false, "Executes the task regardless whether the files have changed or not. Default: false")
	flag.BoolVar(&opts.Init, "init", false, "Initializes a goke.yml file in the current directory")
	flag.BoolVar(&opts.Quiet, "quiet", false, "Disables all output to the console. Default: false")
	flag.Parse()

	return opts
}
