package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/theckman/yacspin"
)

type Executer struct {
	Parser Parser
}

var spinnerCfg = yacspin.Config{
	Frequency:         100 * time.Millisecond,
	Colors:            []string{"fgYellow"},
	CharSet:           yacspin.CharSets[11],
	Suffix:            " ",
	SuffixAutoColon:   true,
	Message:           "Running commands",
	StopCharacter:     "✓",
	StopColors:        []string{"fgGreen"},
	StopMessage:       "Done",
	StopFailCharacter: "✗",
	StopFailColors:    []string{"fgRed"},
	StopFailMessage:   "Failed",
}

// Executes all command strings under given arg.
// Each call happens in its own go routine.
func (e *Executer) Execute(arg string) {
	if _, ok := e.Parser.Commands[arg]; !ok {
		fmt.Printf("command '%s' not found\n", arg)
		os.Exit(1)
	}

	spinner := e.makeSpinner()

	e.scanFiles()
	e.dispatchCommands(arg, spinner)

	spinner.Stop()
}

func (e *Executer) scanFiles() {
	// todo
}

func (e *Executer) dispatchCommands(arg string, spinner *yacspin.Spinner) {
	outputs := make(chan string)
	for _, mainCmd := range e.Parser.Commands[arg].Run {
		for _, beforeEachCmd := range e.Parser.Global.Events.BeforeEach {
			go e.runSysCommand(beforeEachCmd, spinner, outputs)
			fmt.Print(<-outputs)
		}

		go e.runSysCommand(mainCmd, spinner, outputs)
		fmt.Print(<-outputs)

		for _, afterEachCmd := range e.Parser.Global.Events.AfterEach {
			go e.runSysCommand(afterEachCmd, spinner, outputs)
			fmt.Print(<-outputs)
		}
	}
}

// Executes the given string in the underlying OS.
func (e *Executer) runSysCommand(c string, spinner *yacspin.Spinner, outChan chan string) {
	splitCmd, err := e.parseCommandLine(c)

	if err != nil {
		log.Fatalln(err)
	}

	spinner.Message(fmt.Sprintf("Running: %s", c))
	out, err := exec.Command(splitCmd[0], splitCmd[1:]...).Output()

	if err != nil {
		log.Fatalln(err)
	}

	outChan <- "\n" + string(out) + "\n"
}

func (e *Executer) makeSpinner() *yacspin.Spinner {
	spinner, _ := yacspin.New(spinnerCfg)
	spinner.Start()

	return spinner
}

func (e *Executer) parseCommandLine(command string) ([]string, error) {
	var args []string
	state := "start"
	current := ""
	quote := "\""
	escapeNext := true

	for i := 0; i < len(command); i++ {
		c := command[i]

		if state == "quotes" {
			if string(c) != quote {
				current += string(c)
			} else {
				args = append(args, current)
				current = ""
				state = "start"
			}
			continue
		}

		if escapeNext {
			current += string(c)
			escapeNext = false
			continue
		}

		if c == '\\' {
			escapeNext = true
			continue
		}

		if c == '"' || c == '\'' {
			state = "quotes"
			quote = string(c)
			continue
		}

		if state == "arg" {
			if c == ' ' || c == '\t' {
				args = append(args, current)
				current = ""
				state = "start"
			} else {
				current += string(c)
			}
			continue
		}

		if c != ' ' && c != '\t' {
			state = "arg"
			current += string(c)
		}
	}

	if state == "quotes" {
		return []string{}, fmt.Errorf("unclosed quote in command line: %s", command)
	}

	if current != "" {
		args = append(args, current)
	}

	return args, nil
}
