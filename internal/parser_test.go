package internal

import (
	"os"
	"strings"
	"testing"

	"github.com/dugajean/goke/internal/tests"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var yamlConfigStub = `
global:
  environment:
    FOO: "foo"
    BAR: "$(echo 'bar')"
    BAZ: "baz"

events:
  before_each_run:
    - "echo 'before each 1'"
    - "echo 'before each 2'"
  after_each_run:
    - "echo 'after each 1'"
    - "greet-lisha"
  before_each_task:
    - "echo 'before task'"
  after_each_task:
    - "echo 'after task'"

greet-lisha:
  run:
    - "echo 'Hello Lisha!'"

greet-loki:
  run:
    - 'echo "Hello Boki"'

greet-cats:
  files: [cmd/cli/*]
  run:
    - 'echo "Hello Frey"'
    - 'echo "Hello Sunny"'
    - "greet-loki"

greet-thor:
  run:
    - 'echo "Hello ${THOR}"'
  env:
    THOR: "LORD OF THUNDER"`

var clearCacheOpts = Options{
	ClearCache: true,
}

var baseOptions = Options{}

func mockCacheDoesNotExist(t *testing.T) *tests.FileSystem {
	fsMock := tests.NewFileSystem(t)
	fsMock.On("TempDir").Return("path/to/temp")
	fsMock.On("Getwd").Return("path/to/cwd", nil)
	fsMock.On("FileExists", mock.Anything).Return(false).Twice()

	return fsMock
}

func mockCacheDoesNotExistOnce(t *testing.T) *tests.FileSystem {
	fsMock := tests.NewFileSystem(t)
	fsMock.On("TempDir").Return("path/to/temp")
	fsMock.On("Getwd").Return("path/to/cwd", nil)
	fsMock.On("FileExists", mock.Anything).Return(false).Once()
	fsMock.On("FileExists", mock.Anything).Return(true).Once()

	return fsMock
}

func mockCacheExists(t *testing.T) *tests.FileSystem {
	fsMock := tests.NewFileSystem(t)
	fsMock.On("TempDir").Return("path/to/temp")
	fsMock.On("Getwd").Return("path/to/cwd", nil)
	fsMock.On("FileExists", mock.Anything).Return(true).Twice()

	return fsMock
}

func TestNewParserWithoutCache(t *testing.T) {
	fsMock := mockCacheDoesNotExist(t)
	parser := NewParser(yamlConfigStub, &clearCacheOpts, fsMock)
	require.NotNil(t, parser)
}

func TestNewParserWithCache(t *testing.T) {
	fsMock := mockCacheDoesNotExistOnce(t)
	fsMock.On("ReadFile", mock.Anything).Return([]byte(tests.ReadFileBase64), nil)

	parser := NewParser(yamlConfigStub, &clearCacheOpts, fsMock)
	require.NotNil(t, parser)
}

func TestNewParserWithCacheAndWithoutClearCacheFlag(t *testing.T) {
	fsMock := mockCacheExists(t)
	fsMock.On("Stat", mock.Anything).Return(tests.MemFileInfo{}, nil).Twice()
	fsMock.On("ReadFile", mock.Anything).Return([]byte(tests.ReadFileBase64), nil).Once()

	parser := NewParser(yamlConfigStub, &baseOptions, fsMock)
	require.NotNil(t, parser)
}

func TestNewParserWithShouldClearCacheTrue(t *testing.T) {
	fsMock := tests.NewFileSystem(t)
	fsMock.On("TempDir").Return("path/to/temp")
	fsMock.On("Getwd").Return("path/to/cwd", nil)
	fsMock.On("FileExists", mock.Anything).Return(true).Once()
	fsMock.On("FileExists", mock.Anything).Return(false).Once()
	fsMock.On("Remove", mock.Anything).Return(nil)

	parser := NewParser(yamlConfigStub, &clearCacheOpts, fsMock)
	require.NotNil(t, parser)
}

func TestTaskParsing(t *testing.T) {
	fsMock := mockCacheDoesNotExist(t)
	fsMock.On("Glob", mock.Anything).Return([]string{"foo", "bar"}, nil).Once()
	parser := NewParser(yamlConfigStub, &clearCacheOpts, fsMock)

	parser.parseTasks()

	require.NotNil(t, parser.Tasks["greet-loki"])
	require.NotNil(t, parser.Tasks["greet-cats"])
	require.NotNil(t, parser.Tasks["greet-lisha"])
}

func TestGlobalsParsing(t *testing.T) {
	fsMock := mockCacheDoesNotExist(t)
	parser := NewParser(yamlConfigStub, &clearCacheOpts, fsMock)

	parser.parseGlobal()

	require.Equal(t, "foo", os.Getenv("FOO"))
	require.True(t, strings.Contains(os.Getenv("BAR"), "bar"))
	require.Equal(t, "baz", os.Getenv("BAZ"))

	require.Equal(t, "foo", parser.Global.Shared.Environment["FOO"])
	require.True(t, strings.Contains(parser.Global.Shared.Environment["BAR"], "bar"))
	require.Equal(t, "baz", parser.Global.Shared.Environment["BAZ"])
}

func TestTaskGlobFilesExpansion(t *testing.T) {
	expectedGlob := []string{"foo", "bar"}

	fsMock := mockCacheDoesNotExist(t)
	fsMock.On("Glob", mock.Anything).Return(expectedGlob, nil).Once()
	parser := NewParser(yamlConfigStub, &clearCacheOpts, fsMock)

	parser.parseTasks()
	greetCatsTask := parser.Tasks["greet-cats"]

	require.Equal(t, expectedGlob, greetCatsTask.Files)
}

func TestSetEnvVariables(t *testing.T) {
	fsMock := mockCacheDoesNotExist(t)
	parser := NewParser(yamlConfigStub, &clearCacheOpts, fsMock)

	values := map[string]string{
		"THOR":     "Lord of thunder",
		"THOR_CMD": "$(echo 'Hello Thor')",
	}

	want := map[string]string{
		"THOR":     "Lord of thunder",
		"THOR_CMD": "Hello Thor",
	}

	got, _ := parser.setEnvVariables(values)
	require.Equal(t, want["THOR"], os.Getenv("THOR"))
	require.Equal(t, want["THOR_CMD"], os.Getenv("THOR_CMD"))

	for k := range got {
		require.Equal(t, want[k], got[k])
	}
}
