package main

import "os"

func main() {
	p := Parser{}
	p.Bootstrap()

	l := Lockfile{Files: p.FilePaths}
	l.Bootstrap()

	e := Executer{Parser: p}
	if len(os.Args) > 1 {
		e.Execute(os.Args[1])
	} else {
		e.Execute("build")
	}
}
