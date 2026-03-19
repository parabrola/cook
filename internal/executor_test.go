package internal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/parabrola/goke/internal/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func getDependencies(t *testing.T, opts *Options) (*Parseable, *Lockfile, *tests.Process, FileSystem) {
	fsMock := mockCacheDoesNotExist(t)
	fsMock.On("FileExists", mock.Anything).Return(false)
	fsMock.On("WriteFile", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	fsMock.On("Stat", mock.Anything).Return(tests.MemFileInfo{}, nil)
	fsMock.On("ReadFile", mock.Anything).Return([]byte(dotGokeFile), nil)
	fsMock.On("Glob", mock.Anything).Return(tests.ExpectedGlob, nil)

	process := tests.NewProcess(t)

	parser := NewParser(tests.YamlConfigStub, opts, fsMock)
	lockfile := NewLockfile(files, opts, fsMock)

	require.NoError(t, parser.Bootstrap())
	require.NoError(t, lockfile.Bootstrap())

	return &parser, &lockfile, process, fsMock
}

func TestStartNonWatch(t *testing.T) {
	parser, lockfile, process, fsMock := getDependencies(t, &clearCacheOpts)

	process.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	ctx := context.Background()
	executor := NewExecutor(parser, lockfile, &clearCacheOpts, process, fsMock, &ctx)
	executor.Start("greet-loki")

	process.AssertNumberOfCalls(t, "Execute", 1)

	process.AssertExpectations(t)
}

func TestStartWatchWithNoFiles(t *testing.T) {
	watchOpts := Options{
		Watch:   true,
		NoCache: true,
	}

	parser, lockfile, process, fsMock := getDependencies(t, &watchOpts)
	process.On("Exit", mock.Anything).Return()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	executor := NewExecutor(parser, lockfile, &watchOpts, process, fsMock, &ctx)
	executor.Start("greet-loki")
	cancel()

	process.AssertNotCalled(t, "Execute")
	process.AssertNumberOfCalls(t, "Exit", 1)
}

func TestStartDefaultTask(t *testing.T) {
	parser, lockfile, process, fsMock := getDependencies(t, &clearCacheOpts)
	process.On("Exit", mock.Anything).Return()

	ctx := context.Background()
	executor := NewExecutor(parser, lockfile, &clearCacheOpts, process, fsMock, &ctx)
	executor.Start("")

	process.AssertNumberOfCalls(t, "Exit", 1)
}

func TestStartTaskNotFound(t *testing.T) {
	parser, lockfile, process, fsMock := getDependencies(t, &clearCacheOpts)
	process.On("Exit", mock.Anything).Return()

	ctx := context.Background()
	executor := NewExecutor(parser, lockfile, &clearCacheOpts, process, fsMock, &ctx)
	executor.Start("nonexistent-task")

	process.AssertCalled(t, "Exit", 1)
}

func TestExecuteCommandError(t *testing.T) {
	parser, lockfile, process, fsMock := getDependencies(t, &clearCacheOpts)
	process.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(errors.New("command failed"))
	process.On("Exit", mock.Anything).Return()

	ctx := context.Background()
	executor := NewExecutor(parser, lockfile, &clearCacheOpts, process, fsMock, &ctx)
	executor.Start("greet-loki")

	process.AssertCalled(t, "Exit", 1)
}

func TestExecuteWithForceFlag(t *testing.T) {
	forceOpts := Options{
		NoCache: true,
		Force:   true,
	}

	parser, lockfile, process, fsMock := getDependencies(t, &forceOpts)
	process.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	ctx := context.Background()
	executor := NewExecutor(parser, lockfile, &forceOpts, process, fsMock, &ctx)
	executor.Start("greet-loki")

	process.AssertNumberOfCalls(t, "Execute", 1)
}

func TestExecuteWithQuietFlag(t *testing.T) {
	quietOpts := Options{
		NoCache: true,
		Quiet:   true,
	}

	parser, lockfile, process, fsMock := getDependencies(t, &quietOpts)
	process.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	ctx := context.Background()
	executor := NewExecutor(parser, lockfile, &quietOpts, process, fsMock, &ctx)
	executor.Start("greet-loki")

	process.AssertNumberOfCalls(t, "Execute", 1)
}

func TestExecuteRecursiveTask(t *testing.T) {
	parser, lockfile, process, fsMock := getDependencies(t, &clearCacheOpts)
	process.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	ctx := context.Background()
	executor := NewExecutor(parser, lockfile, &clearCacheOpts, process, fsMock, &ctx)
	executor.Start("greet-cats")

	process.AssertNumberOfCalls(t, "Execute", 3)
	process.AssertExpectations(t)
}

func TestExecuteWithArgs(t *testing.T) {
	argsOpts := Options{
		NoCache: true,
		Args:    []string{"--verbose", "--debug"},
	}

	parser, lockfile, process, fsMock := getDependencies(t, &argsOpts)
	process.On("Execute", "echo", "Hello Boki", "--verbose", "--debug").Return(nil)

	ctx := context.Background()
	executor := NewExecutor(parser, lockfile, &argsOpts, process, fsMock, &ctx)
	executor.Start("greet-loki")

	process.AssertNumberOfCalls(t, "Execute", 1)
}

func TestExecuteWithEvents(t *testing.T) {
	psr, lockfile, process, fsMock := getDependencies(t, &clearCacheOpts)
	process.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	concrete := (*psr).(*parser)
	concrete.Global.Shared.Events.BeforeEachTask = []string{"echo 'before task'"}
	concrete.Global.Shared.Events.AfterEachTask = []string{"echo 'after task'"}

	ctx := context.Background()
	executor := NewExecutor(psr, lockfile, &clearCacheOpts, process, fsMock, &ctx)
	executor.Start("greet-loki")

	assert.GreaterOrEqual(t, len(process.Calls), 3)
}

func TestExecuteWithBeforeAfterEachRun(t *testing.T) {
	psr, lockfile, process, fsMock := getDependencies(t, &clearCacheOpts)
	process.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	concrete := (*psr).(*parser)
	concrete.Global.Shared.Events.BeforeEachRun = []string{"echo 'before run'"}
	concrete.Global.Shared.Events.AfterEachRun = []string{"echo 'after run'"}

	ctx := context.Background()
	executor := NewExecutor(psr, lockfile, &clearCacheOpts, process, fsMock, &ctx)
	executor.Start("greet-loki")

	assert.GreaterOrEqual(t, len(process.Calls), 3)
}

func getDependenciesWithConfig(t *testing.T, opts *Options, yamlConfig string) (*Parseable, *Lockfile, *tests.Process, FileSystem) {
	fsMock := mockCacheDoesNotExist(t)
	fsMock.On("FileExists", mock.Anything).Return(false)
	fsMock.On("WriteFile", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	fsMock.On("Stat", mock.Anything).Return(tests.MemFileInfo{}, nil)
	fsMock.On("ReadFile", mock.Anything).Return([]byte(dotGokeFile), nil)

	process := tests.NewProcess(t)

	parser := NewParser(yamlConfig, opts, fsMock)
	lockfile := NewLockfile(files, opts, fsMock)

	require.NoError(t, parser.Bootstrap())
	require.NoError(t, lockfile.Bootstrap())

	return &parser, &lockfile, process, fsMock
}

func TestExecuteWithDependsOn(t *testing.T) {
	config := `
clean:
  run:
    - "echo cleaning"

build:
  depends_on: [clean]
  run:
    - "echo building"
`
	parser, lockfile, process, fsMock := getDependenciesWithConfig(t, &clearCacheOpts, config)
	process.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	ctx := context.Background()
	executor := NewExecutor(parser, lockfile, &clearCacheOpts, process, fsMock, &ctx)
	executor.Start("build")

	process.AssertNumberOfCalls(t, "Execute", 2)
}

func TestExecuteWithDiamondDependency(t *testing.T) {
	config := `
base:
  run:
    - "echo base"

left:
  depends_on: [base]
  run:
    - "echo left"

right:
  depends_on: [base]
  run:
    - "echo right"

top:
  depends_on: [left, right]
  run:
    - "echo top"
`
	parser, lockfile, process, fsMock := getDependenciesWithConfig(t, &clearCacheOpts, config)
	process.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	ctx := context.Background()
	executor := NewExecutor(parser, lockfile, &clearCacheOpts, process, fsMock, &ctx)
	executor.Start("top")

	process.AssertNumberOfCalls(t, "Execute", 4)
}

func TestExecuteWithDependsOnChain(t *testing.T) {
	config := `
step1:
  run:
    - "echo step1"

step2:
  depends_on: [step1]
  run:
    - "echo step2"

step3:
  depends_on: [step2]
  run:
    - "echo step3"
`
	parser, lockfile, process, fsMock := getDependenciesWithConfig(t, &clearCacheOpts, config)
	process.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	ctx := context.Background()
	executor := NewExecutor(parser, lockfile, &clearCacheOpts, process, fsMock, &ctx)
	executor.Start("step3")

	process.AssertNumberOfCalls(t, "Execute", 3)
}

func TestExecuteDependencyFailsStopsExecution(t *testing.T) {
	config := `
failing:
  run:
    - "echo fail"

do-stuff:
  depends_on: [failing]
  run:
    - "echo success"
`
	parser, lockfile, process, fsMock := getDependenciesWithConfig(t, &clearCacheOpts, config)
	process.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(errors.New("command failed"))
	process.On("Exit", mock.Anything).Return()

	ctx := context.Background()
	executor := NewExecutor(parser, lockfile, &clearCacheOpts, process, fsMock, &ctx)
	executor.Start("do-stuff")

	process.AssertNumberOfCalls(t, "Execute", 1)
	process.AssertCalled(t, "Exit", 1)
}
