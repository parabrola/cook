package internal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/parabrola/goke/internal/cli"
	"github.com/theckman/yacspin"
)

// This represent the default task, so when the user
// doesn't provide any args to the program, we default to this.
const DefaultTask = "main"

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

type Executor struct {
	parser    Parseable
	lockfile  Lockfile
	spinner   *yacspin.Spinner
	options   Options
	process   Process
	fs        FileSystem
	context   context.Context
	completed sync.Map
	mu        sync.Mutex
}

// Executor constructor.
func NewExecutor(p *Parseable, l *Lockfile, opts *Options, proc Process, fs FileSystem, ctx *context.Context) Executor {
	spinner, _ := yacspin.New(spinnerCfg)

	return Executor{
		parser:   *p,
		lockfile: *l,
		spinner:  spinner,
		options:  *opts,
		process:  proc,
		fs:       fs,
		context:  *ctx,
	}
}

// Starts the command for a single run or as a watcher.
func (e *Executor) Start(taskName string) {
	e.completed = sync.Map{}

	arg := DefaultTask
	if taskName != "" {
		arg = taskName
	}

	if e.options.Watch {
		if err := e.watch(arg); err != nil {
			e.logErr(err)
		}
	} else {
		if err := e.execute(arg); err != nil {
			e.logErr(err)
		}
	}
}

// Executes all command strings under given taskName.
// Each call happens in its own go routine.
func (e *Executor) execute(taskName string) error {
	task := e.initTask(taskName)
	didDispatch, err := e.checkAndDispatch(task)

	if err != nil {
		return err
	}

	if !didDispatch {
		e.logExit("success", "Nothing to run")
	}

	e.spinner.StopMessage("Done!")
	e.spinner.Stop()

	return nil
}

// Begins an infinite loop that watches for the file changes
// in the "files" section of the task's configuration.
func (e *Executor) watch(taskName string) error {
	task := e.initTask(taskName)

	if len(task.Files) == 0 {
		return errors.New("task has no files to watch")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// Watch the directories containing the tracked files,
	// since watching individual files can miss editor save patterns
	// (editors often write to a temp file then rename).
	dirs := make(map[string]struct{})
	for _, f := range task.Files {
		dir := filepath.Dir(f)
		dirs[dir] = struct{}{}
	}
	for dir := range dirs {
		if err := watcher.Add(dir); err != nil {
			return err
		}
	}

	// Build a set for fast lookup of tracked files.
	tracked := make(map[string]struct{}, len(task.Files))
	for _, f := range task.Files {
		abs, err := filepath.Abs(f)
		if err != nil {
			continue
		}
		tracked[abs] = struct{}{}
	}

	// Run once on startup.
	if _, err := e.checkAndDispatch(task); err != nil {
		e.logErr(err)
	}

	e.spinner.Message("Watching for file changes...")

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			abs, _ := filepath.Abs(event.Name)
			if _, isTracked := tracked[abs]; !isTracked {
				continue
			}

			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}

			if _, err := e.checkAndDispatch(task); err != nil {
				e.logErr(err)
			}

			e.spinner.Message("Watching for file changes...")

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			return err

		case <-e.context.Done():
			return nil
		}
	}
}

// Checks whether the task will be dispatched or not,
// and then dispatches is true. Returns true if dispatched.
func (e *Executor) checkAndDispatch(task Task) (bool, error) {
	shouldDispatch, err := e.shouldDispatch(task)
	if err != nil {
		return false, err
	}

	if shouldDispatch || e.options.Force {
		if err := e.dispatchTask(task, true); err != nil {
			return false, err
		}
	}

	return (shouldDispatch || e.options.Force), nil
}

// Fetch the task from the parser based on task name.
func (e *Executor) initTask(taskName string) Task {
	if !e.options.Quiet {
		e.spinner.Start()
	}

	e.mustExist(taskName)
	task, _ := e.parser.GetTask(taskName)
	return task
}

// Checks whether files have changed since the last run.
// Also updates the lockfile if files did get modified.
// If no "files" key is present in the task, simply returns true.
func (e *Executor) shouldDispatch(task Task) (bool, error) {
	if len(task.Files) == 0 {
		return true, nil
	}

	dispatchCh := make(chan Ref[bool])
	go e.shouldDispatchRoutine(task, dispatchCh)
	dispatch := <-dispatchCh

	if dispatch.Error() != nil {
		return false, dispatch.Error()
	}

	if dispatch.Value() {
		e.lockfile.UpdateTimestampsForFiles(task.Files)
	}

	return dispatch.Value(), nil
}

// Go Routine function that determines whether the stored
// mtime is greater  than mtime if the file at this moment.
func (e *Executor) shouldDispatchRoutine(task Task, ch chan Ref[bool]) {
	lockedModTimes := e.lockfile.GetCurrentProject()

	for _, f := range task.Files {
		fo, err := e.fs.Stat(f)
		if err != nil {
			ch <- NewRef(false, err)
			return
		}

		modTimeNow := fo.ModTime().Unix()

		if lockedModTimes[f] < modTimeNow {
			ch <- NewRef(true, nil)
			return
		}
	}

	ch <- NewRef(false, nil)
}

func (e *Executor) executeDependencies(task Task) error {
	if len(task.DependsOn) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(task.DependsOn))

	for _, depName := range task.DependsOn {
		if _, loaded := e.completed.LoadOrStore(depName, true); loaded {
			continue
		}

		depTask, ok := e.parser.GetTask(depName)
		if !ok {
			return fmt.Errorf("dependency '%s' not found", depName)
		}

		wg.Add(1)
		go func(dt Task) {
			defer wg.Done()

			if err := e.executeDependencies(dt); err != nil {
				errCh <- err
				return
			}

			if err := e.dispatchTask(dt, false); err != nil {
				errCh <- err
			}
		}(depTask)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

// Dispatches the individual commands of the current task,
// including any events that need to be run.
func (e *Executor) dispatchTask(task Task, initialRun bool) error {
	if err := e.executeDependencies(task); err != nil {
		return err
	}

	global := e.parser.GetGlobal()

	if initialRun {
		for _, beforeEachCmd := range global.Shared.Events.BeforeEachTask {
			if err := e.runSysOrRecurse(beforeEachCmd); err != nil {
				return err
			}
		}
	}

	for _, mainCmd := range task.Run {
		if initialRun {
			for _, beforeEachCmd := range global.Shared.Events.BeforeEachRun {
				if err := e.runSysOrRecurse(beforeEachCmd); err != nil {
					return err
				}
			}
		}

		if err := e.runSysOrRecurse(mainCmd); err != nil {
			return err
		}

		if initialRun {
			for _, afterEachCmd := range global.Shared.Events.AfterEachRun {
				if err := e.runSysOrRecurse(afterEachCmd); err != nil {
					return err
				}
			}
		}
	}

	for _, afterEachCmd := range global.Shared.Events.AfterEachTask {
		if err := e.runSysOrRecurse(afterEachCmd); err != nil {
			return err
		}
	}

	return nil
}

// Determine what to execute: system command or another declared task in goke.yml.
func (e *Executor) runSysOrRecurse(cmd string) error {
	if !e.options.Quiet {
		message := cmd
		if len(e.options.Args) > 0 {
			message = fmt.Sprintf("%s %s", message, JoinInnerArgs(e.options.Args))
		}

		e.spinner.Message(fmt.Sprintf("Running: %s", message))
	}

	if task, ok := e.parser.GetTask(cmd); ok {
		return e.dispatchTask(task, false)
	}

	e.mu.Lock()
	if !e.options.Quiet {
		e.spinner.Pause()
	}
	err := e.runSysCommand(cmd)
	if !e.options.Quiet {
		e.spinner.Unpause()
	}
	e.mu.Unlock()

	return err
}

// Executes the given string in the underlying OS.
func (e *Executor) runSysCommand(c string) error {
	splitCmd, err := cli.ParseCommandLine(os.ExpandEnv(c))
	if err != nil {
		return err
	}

	wholeCmd := append(splitCmd[1:], e.options.Args...)
	return e.process.Execute(splitCmd[0], wholeCmd...)
}

func (e *Executor) mustExist(taskName string) {
	if _, ok := e.parser.GetTask(taskName); !ok {
		e.logExit("error", fmt.Sprintf("Command '%s' not found\n", taskName))
	}
}

// Shortcut to logging an error using spinner logger.
func (e *Executor) logErr(err error) {
	e.logExit("error", fmt.Sprintf("Error: %s\n", err.Error()))
}

// Log to the console using the spinner instance.
func (e *Executor) logExit(status string, message string) {
	switch status {
	default:
	case "success":
		if !e.options.Quiet {
			e.spinner.StopMessage(message)
			e.spinner.Stop()
		}
		e.process.Exit(0)
	case "error":
		if !e.options.Quiet {
			e.spinner.StopFailMessage(message)
			e.spinner.StopFail()
		}
		e.process.Exit(1)
	}
}
