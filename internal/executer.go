package internal

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/theckman/yacspin"
)

// This represent the default task, so when the user
// doesn't provide any args to the program, we default to this.
const DefaultTask = "main"

type Executer struct {
	parser   Parser
	lockfile Lockfile
	spinner  *yacspin.Spinner
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

// Executer constructor.
func NewExecuter(p *Parser, l *Lockfile) Executer {
	spinner, _ := yacspin.New(spinnerCfg)

	return Executer{
		parser:   *p,
		lockfile: *l,
		spinner:  spinner,
	}
}

// Executes all command strings under given taskName.
// Each call happens in its own go routine.
func (e *Executer) Execute(taskName string) {
	e.spinner.Start()

	if _, ok := e.parser.Tasks[taskName]; !ok {
		e.log("error", fmt.Sprintf("Command '%s' not found\n", taskName))
	}

	task := e.parser.Tasks[taskName]
	shouldDispatch, err := e.shouldDispatch(task)

	if err != nil {
		e.logErr(err)
	}

	if shouldDispatch {
		err := e.dispatchCommands(task, true)

		if err != nil {
			e.logErr(err)
		}
	} else {
		e.log("success", "Nothing to run")
	}
}

// Checks whether files have changed since the last run.
// Also updates the lockfile if files did get modified.
// If no "files" key is present in the task, simply returns true.
func (e *Executer) shouldDispatch(task Task) (bool, error) {
	if len(task.Files) == 0 {
		return true, nil
	}

	dispatchCh := make(chan Ref[bool])
	go e.shouldDispatchRoutine(task, dispatchCh)
	dispatch := <-dispatchCh

	if dispatch.Error() != nil {
		return false, dispatch.Error()
	}

	if dispatch.Equal(true) {
		e.lockfile.UpdateTimestampsForFiles(task.Files)
	}

	return dispatch.Value(), nil
}

// Go Routine function that determines whether the stored
// mtime is greater  than mtime if the file at this moment.
func (e *Executer) shouldDispatchRoutine(task Task, ch chan Ref[bool]) {
	lockedModTimes := e.lockfile.GetCurrentProject()

	for _, f := range task.Files {
		fo, err := os.Stat(f)
		if err != nil {
			ch <- NewRef(false, err)
		}

		modTimeNow := fo.ModTime().Unix()
		if lockedModTimes[f] < modTimeNow {
			ch <- NewRef(true, nil)
			return
		}
	}

	ch <- NewRef(false, nil)
}

// Dispatches the individual commands of the current task,
// including and events that need to be run.
func (e *Executer) dispatchCommands(task Task, initialRun bool) error {
	outputs := make(chan Ref[string])
	if initialRun {
		for _, beforeEachCmd := range e.parser.Global.Shared.Events.BeforeEachTask {
			err := e.runSysOrRecurse(beforeEachCmd, &outputs)

			if err != nil {
				return err
			}
		}
	}

	for _, mainCmd := range task.Run {
		if initialRun {
			for _, beforeEachCmd := range e.parser.Global.Shared.Events.BeforeEachRun {
				if err := e.runSysOrRecurse(beforeEachCmd, &outputs); err != nil {
					return err
				}
			}
		}

		if err := e.runSysOrRecurse(mainCmd, &outputs); err != nil {
			return err
		}

		if initialRun {
			for _, afterEachCmd := range e.parser.Global.Shared.Events.AfterEachRun {
				if err := e.runSysOrRecurse(afterEachCmd, &outputs); err != nil {
					return err
				}
			}
		}
	}

	for _, afterEachCmd := range e.parser.Global.Shared.Events.AfterEachTask {
		if err := e.runSysOrRecurse(afterEachCmd, &outputs); err != nil {
			return err
		}
	}

	return nil
}

// Determine what to execute: system command or another declared task in goke.yml.
func (e *Executer) runSysOrRecurse(cmd string, ch *chan Ref[string]) error {
	e.spinner.Message(fmt.Sprintf("Running: %s", cmd))

	if _, ok := e.parser.Tasks[cmd]; ok {
		return e.dispatchCommands(e.parser.Tasks[cmd], false)
	} else {
		go e.runSysCommand(cmd, *ch)
		output := <-*ch

		if output.Error() != nil {
			return output.Error()
		}

		fmt.Print(output.Value())
	}

	return nil
}

// Executes the given string in the underlying OS.
func (e *Executer) runSysCommand(c string, ch chan Ref[string]) {
	splitCmd, err := e.parseCommandLine(c)

	if err != nil {
		ch <- NewRef("", err)
		return
	}

	out, err := exec.Command(splitCmd[0], splitCmd[1:]...).Output()

	if err != nil {
		ch <- NewRef("", err)
		return
	}

	ch <- NewRef("\n"+string(out)+"\n", nil)
}

func (e *Executer) logErr(err error) {
	e.log("error", fmt.Sprintf("Error: %s\n", err.Error()))
}

func (e *Executer) log(status string, message string) {
	switch status {
	default:
	case "success":
		e.spinner.StopMessage(message)
		e.spinner.Stop()
		os.Exit(0)
	case "error":
		e.spinner.StopFailMessage(message)
		e.spinner.StopFail()
		os.Exit(1)
	}
}

// Parses the command string into an array of [command, args, args]...
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
		return []string{}, fmt.Errorf("unclosed quote in command: %s", command)
	}

	if current != "" {
		args = append(args, current)
	}

	return args, nil
}
