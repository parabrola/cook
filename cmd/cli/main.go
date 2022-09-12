package main

import (
	"os"

	app "github.com/dugajean/goke/internal"
	"github.com/dugajean/goke/internal/cli"
)

func main() {
	argIndex := app.PermutateArgs(os.Args)
	opts := cli.GetOptions()
	rw := app.FileOSWrapper{}

	p := app.NewParser(app.ReadYamlConfig(), &opts, &rw) //
	p.Bootstrap()

	l := app.NewLockfile(p.FilePaths, &rw)
	l.Bootstrap()

	e := app.NewExecuter(&p, &l, &opts)
	e.Start(parseTaskName(argIndex))
}

func parseTaskName(argIndex int) string {
	arg := ""

	if len(os.Args) > (argIndex - 1) {
		arg = os.Args[argIndex]
	}

	return arg
}
