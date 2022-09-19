package main

import (
	"os"

	app "github.com/dugajean/goke/internal"
	"github.com/dugajean/goke/internal/cli"
)

func main() {
	argIndex := app.PermutateArgs(os.Args)
	opts := cli.GetOptions()

	p := app.NewParser(app.ReadYamlConfig(), &opts) //
	p.Bootstrap()

	l := app.NewLockfile(p.FilePaths)
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
