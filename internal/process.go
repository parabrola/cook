package internal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

type Process interface {
	Execute(name string, args ...string) error
	Fprint(w io.Writer, a ...any) (n int, err error)
	Exit(code int)
}

type ShellProcess struct{}

func (sp *ShellProcess) Execute(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (sp *ShellProcess) Fprint(w io.Writer, a ...any) (n int, err error) {
	return fmt.Fprint(w, a...)
}

func (sp *ShellProcess) Exit(code int) {
	os.Exit(code)
}
