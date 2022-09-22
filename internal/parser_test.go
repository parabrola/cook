package internal

import (
	"testing"

	"github.com/dugajean/goke/internal/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
    - "greet-loki"`

var options = Options{
	ClearCache: true,
}

func TestNewParser(t *testing.T) {
	fsMock := tests.NewFileSystem(t)
	fsMock.On("TempDir").Return("path/to/temp")
	fsMock.On("Getwd").Return("path/to/cwd", nil)
	fsMock.On("FileExists", mock.Anything).Return(false)

	parser := NewParser(yamlConfigStub, &options, fsMock)
	assert.NotNil(t, parser)
}

func TestTaskParsing(t *testing.T) {
	fsMock := tests.NewFileSystem(t)
	fsMock.On("TempDir").Return("path/to/temp")
	fsMock.On("Getwd").Return("path/to/cwd", nil)
	fsMock.On("FileExists", mock.Anything).Return(true)
	fsMock.On("Remove", mock.Anything).Return(nil)
	fsMock.On("ReadFile", mock.Anything).Return([]byte(tests.ReadFileBase64), nil)

	parser := NewParser(yamlConfigStub, &options, fsMock)

	parser.parseTasks()
	assert.NotNil(t, parser.Tasks)

	assert.NotNil(t, parser.Tasks["greet-loki"])
	assert.NotNil(t, parser.Tasks["greet-cats"])
	assert.NotNil(t, parser.Tasks["greet-lisha"])
}

func TestGlobalsParsing(t *testing.T) {
	fsMock := tests.NewFileSystem(t)
	fsMock.On("TempDir").Return("path/to/temp")
	fsMock.On("Getwd").Return("path/to/cwd", nil)
	fsMock.On("FileExists", mock.Anything).Return(false)

	parser := NewParser(yamlConfigStub, &options, fsMock)

	parser.parseGlobal()
	assert.Equal(t, "foo", parser.Global.Shared.Environment["FOO"])
	assert.Equal(t, "bar\n", parser.Global.Shared.Environment["BAR"])
	assert.Equal(t, "baz", parser.Global.Shared.Environment["BAZ"])
}

func TestTaskFilesExpansion(t *testing.T) {
	fsMock := tests.NewFileSystem(t)
	fsMock.On("TempDir").Return("path/to/temp")
	fsMock.On("Getwd").Return("path/to/cwd", nil)
	fsMock.On("FileExists", mock.Anything).Return(true)
	fsMock.On("Remove", mock.Anything).Return(nil)
	fsMock.On("ReadFile", mock.Anything).Return([]byte(tests.ReadFileBase64), nil)

	parser := NewParser(yamlConfigStub, &options, fsMock)
	parser.parseTasks()

	greetCatsTask := parser.Tasks["greet-cats"]
	assert.NotNil(t, greetCatsTask)
}

func TestParserWithoutCache(t *testing.T) {
	fsMock := tests.NewFileSystem(t)
	fsMock.On("TempDir").Return("path/to/temp")
	fsMock.On("Getwd").Return("path/to/cwd", nil)
	fsMock.On("FileExists", mock.Anything).Return(true)
	fsMock.On("Remove", mock.Anything).Return(nil)
	fsMock.On("ReadFile", mock.Anything).Return([]byte(tests.ReadFileBase64), nil)

	parser := NewParser(yamlConfigStub, &options, fsMock)
	parser.parseTasks()

	greetCatsTask := parser.Tasks["greet-cats"]
	assert.NotNil(t, greetCatsTask)
}
