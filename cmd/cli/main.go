package main

import (
	"fmt"
	"os"

	app "github.com/dugajean/goke/internal"
	"github.com/dugajean/goke/internal/cli"
)

func main() {
	argIndex := app.PermutateArgs(os.Args)
	opts := cli.GetOptions()

	handleInit(opts)

	fs := app.LocalFileSystem{}
	p := app.NewParser(app.ReadYamlConfig(), &opts, &fs)
	p.Bootstrap()

	l := app.NewLockfile(p.FilePaths, &opts, &fs)
	l.Bootstrap()

	e := app.NewExecutor(&p, &l, &opts)
	e.Start(parseTaskName(argIndex))
}

func parseTaskName(argIndex int) string {
	arg := ""

	if len(os.Args) > (argIndex - 1) {
		arg = os.Args[argIndex]
	}

	return arg
}

func handleInit(opts app.Options) {
	if opts.Init {
		err := app.CreateGokeConfig()

		if err != nil && !opts.Quiet {
			fmt.Println(err)
		}

		os.Exit(0)
	}
}
