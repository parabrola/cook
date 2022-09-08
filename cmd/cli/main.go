package main

import (
	"os"

	app "github.com/dugajean/goke/internal"
)

func main() {
	clearCache := len(os.Args) > 2 && os.Args[2] == "-c"

	p := app.NewParser(clearCache)
	p.Bootstrap()

	l := app.Lockfile{Files: p.FilePaths}
	l.Bootstrap()

	e := app.NewExecuter(&p, &l)
	if len(os.Args) > 0 {
		e.Execute(os.Args[1])
	} else {
		e.Execute(app.DefaultTask)
	}
}
