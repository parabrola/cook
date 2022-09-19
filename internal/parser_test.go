package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	parser := NewParser(yamlConfigStub, &options)
	assert.NotNil(t, parser)
}

func TestTaskParsing(t *testing.T) {
	parser := NewParser(yamlConfigStub, &options)
	assert.Nil(t, parser.Tasks)

	parser.parseTasks()
	assert.NotNil(t, parser.Tasks)

	assert.NotNil(t, parser.Tasks["greet-loki"])
	assert.NotNil(t, parser.Tasks["greet-cats"])
	assert.NotNil(t, parser.Tasks["greet-lisha"])
}

func TestGlobalsParsing(t *testing.T) {
	parser := NewParser(yamlConfigStub, &options)

	parser.parseGlobal()
	assert.Equal(t, "foo", parser.Global.Shared.Environment["FOO"])
	assert.Equal(t, "bar\n", parser.Global.Shared.Environment["BAR"])
	assert.Equal(t, "baz", parser.Global.Shared.Environment["BAZ"])
}

func TestTaskFilesExpansion(t *testing.T) {
	parser := NewParser(yamlConfigStub, &options)
	parser.parseTasks()

	greetCatsTask := parser.Tasks["greet-cats"]
	assert.NotNil(t, greetCatsTask)
}

func TestParserWithoutCache(t *testing.T) {
	parser := NewParser(yamlConfigStub, &options)
	parser.parseTasks()

	greetCatsTask := parser.Tasks["greet-cats"]
	assert.NotNil(t, greetCatsTask)
}
