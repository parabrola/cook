package internal

import (
	"os"
	"strings"
	"testing"

	"github.com/parabrola/cook/internal/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var clearCacheOpts = Options{
	NoCache: true,
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
	parser := NewParser(tests.YamlConfigStub, &clearCacheOpts, fsMock)
	require.NotNil(t, parser)
}

func TestNewParserWithCache(t *testing.T) {
	fsMock := mockCacheDoesNotExistOnce(t)
	fsMock.On("ReadFile", mock.Anything).Return([]byte(tests.ReadFileBase64), nil)

	parser := NewParser(tests.YamlConfigStub, &clearCacheOpts, fsMock)
	require.NotNil(t, parser)
}

func TestNewParserWithCacheAndWithoutClearCacheFlag(t *testing.T) {
	fsMock := mockCacheExists(t)
	fsMock.On("Stat", mock.Anything).Return(tests.MemFileInfo{}, nil).Twice()
	fsMock.On("ReadFile", mock.Anything).Return([]byte(tests.ReadFileBase64), nil).Once()

	parser := NewParser(tests.YamlConfigStub, &baseOptions, fsMock)
	require.NotNil(t, parser)
}

func TestNewParserWithShouldClearCacheTrue(t *testing.T) {
	fsMock := tests.NewFileSystem(t)
	fsMock.On("TempDir").Return("path/to/temp")
	fsMock.On("Getwd").Return("path/to/cwd", nil)
	fsMock.On("FileExists", mock.Anything).Return(true).Once()
	fsMock.On("FileExists", mock.Anything).Return(false).Once()
	fsMock.On("Remove", mock.Anything).Return(nil)

	parser := NewParser(tests.YamlConfigStub, &clearCacheOpts, fsMock)
	require.NotNil(t, parser)
}

func TestTaskParsing(t *testing.T) {
	fsMock := mockCacheDoesNotExist(t)
	fsMock.On("Glob", mock.Anything).Return([]string{"foo", "bar"}, nil).Once()
	parser := NewParser(tests.YamlConfigStub, &clearCacheOpts, fsMock)

	parser.parseTasks()

	greetLoki, _ := parser.GetTask("greet-loki")
	greetCats, _ := parser.GetTask("greet-cats")
	greetLisha, _ := parser.GetTask("greet-lisha")
	require.NotNil(t, greetLoki)
	require.NotNil(t, greetCats)
	require.NotNil(t, greetLisha)
}

func TestGlobalsParsing(t *testing.T) {
	fsMock := mockCacheDoesNotExist(t)
	fsMock.On("FileExists", ".env").Return(false)
	fsMock.On("FileExists", ".env.local").Return(false)
	parser := NewParser(tests.YamlConfigStub, &clearCacheOpts, fsMock)

	parser.parseGlobal()

	require.Equal(t, "foo", os.Getenv("FOO"))
	require.True(t, strings.Contains(os.Getenv("BAR"), "bar"))
	require.Equal(t, "baz", os.Getenv("BAZ"))

	global := parser.GetGlobal()
	require.Equal(t, "foo", global.Shared.Env["FOO"])
	require.True(t, strings.Contains(global.Shared.Env["BAR"], "bar"))
	require.Equal(t, "baz", global.Shared.Env["BAZ"])
}

func TestTaskGlobFilesExpansion(t *testing.T) {
	fsMock := mockCacheDoesNotExist(t)
	fsMock.On("Glob", mock.Anything).Return(tests.ExpectedGlob, nil)
	parser := NewParser(tests.YamlConfigStub, &clearCacheOpts, fsMock)

	parser.parseTasks()
	greetCatsTask, _ := parser.GetTask("greet-cats")

	require.Equal(t, tests.ExpectedGlob, greetCatsTask.Files)
}

func TestGetTaskNotFound(t *testing.T) {
	fsMock := mockCacheDoesNotExist(t)
	fsMock.On("Glob", mock.Anything).Return(tests.ExpectedGlob, nil)
	parser := NewParser(tests.YamlConfigStub, &clearCacheOpts, fsMock)

	parser.parseTasks()

	_, ok := parser.GetTask("nonexistent")
	assert.False(t, ok)
}

func TestFilesPlaceholderReplacement(t *testing.T) {
	config := `
my-task:
  files: [cmd/cli/*]
  run:
    - "go vet {FILES}"
`
	fsMock := mockCacheDoesNotExist(t)
	fsMock.On("Glob", mock.Anything).Return([]string{"cmd/cli/main.go", "cmd/cli/util.go"}, nil)
	parser := NewParser(config, &clearCacheOpts, fsMock)

	parser.parseTasks()
	task, ok := parser.GetTask("my-task")

	assert.True(t, ok)
	assert.Contains(t, task.Run[0], "cmd/cli/main.go")
	assert.Contains(t, task.Run[0], "cmd/cli/util.go")
	assert.NotContains(t, task.Run[0], "{FILES}")
}

func TestTaskEnvVariables(t *testing.T) {
	fsMock := mockCacheDoesNotExist(t)
	fsMock.On("Glob", mock.Anything).Return(tests.ExpectedGlob, nil)
	parser := NewParser(tests.YamlConfigStub, &clearCacheOpts, fsMock)

	parser.parseTasks()
	task, ok := parser.GetTask("greet-thor")

	assert.True(t, ok)
	assert.Equal(t, "LORD OF THUNDER", os.Getenv("THOR"))
	assert.Equal(t, "LORD OF THUNDER", task.Env["THOR"])
}

func TestParseTasksRejectsCyclicDependencies(t *testing.T) {
	config := `
a:
  depends_on: [b]
  run:
    - "echo a"

b:
  depends_on: [a]
  run:
    - "echo b"
`
	fsMock := mockCacheDoesNotExist(t)
	parser := NewParser(config, &clearCacheOpts, fsMock)

	err := parser.parseTasks()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestParseTasksRejectsMissingDependency(t *testing.T) {
	config := `
a:
  depends_on: [nonexistent]
  run:
    - "echo a"
`
	fsMock := mockCacheDoesNotExist(t)
	parser := NewParser(config, &clearCacheOpts, fsMock)

	err := parser.parseTasks()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown task")
}

func TestJSONTaskParsing(t *testing.T) {
	config := `{
  "global": {
    "environment": {
      "FOO": "bar"
    }
  },
  "greet": {
    "run": ["echo hello"]
  },
  "build": {
    "files": ["cmd/cli/*"],
    "depends_on": ["greet"],
    "run": ["go build ./..."]
  }
}`
	fsMock := mockCacheDoesNotExist(t)
	fsMock.On("Glob", mock.Anything).Return([]string{"cmd/cli/main.go"}, nil)
	fsMock.On("FileExists", ".env").Return(false)
	fsMock.On("FileExists", ".env.local").Return(false)
	parser := NewParser(config, &clearCacheOpts, fsMock)

	err := parser.parseGlobal()
	require.NoError(t, err)

	err = parser.parseTasks()
	require.NoError(t, err)

	greet, ok := parser.GetTask("greet")
	assert.True(t, ok)
	assert.Equal(t, []string{"echo hello"}, greet.Run)

	build, ok := parser.GetTask("build")
	assert.True(t, ok)
	assert.Equal(t, []string{"go build ./..."}, build.Run)
	assert.Equal(t, []string{"greet"}, build.DependsOn)
	assert.Equal(t, []string{"cmd/cli/main.go"}, build.Files)
}

func TestJSONSchemaKeyExcludedFromTasks(t *testing.T) {
	config := `{
  "$schema": "cook.schema.json",
  "greet": {
    "run": ["echo hello"]
  }
}`
	fsMock := mockCacheDoesNotExist(t)
	parser := NewParser(config, &clearCacheOpts, fsMock)

	err := parser.parseTasks()
	require.NoError(t, err)

	_, ok := parser.GetTask("$schema")
	assert.False(t, ok)

	greet, ok := parser.GetTask("greet")
	assert.True(t, ok)
	assert.Equal(t, []string{"echo hello"}, greet.Run)
}

func TestLooksLikeJSON(t *testing.T) {
	assert.True(t, looksLikeJSON(`{"foo": "bar"}`))
	assert.True(t, looksLikeJSON(`  { "foo": "bar" }`))
	assert.False(t, looksLikeJSON(`global:`))
	assert.False(t, looksLikeJSON(`foo: bar`))
	assert.False(t, looksLikeJSON(``))
}

func TestGlobalKeyExcludedFromTasks(t *testing.T) {
	fsMock := mockCacheDoesNotExist(t)
	fsMock.On("Glob", mock.Anything).Return(tests.ExpectedGlob, nil)
	parser := NewParser(tests.YamlConfigStub, &clearCacheOpts, fsMock)

	parser.parseTasks()

	_, ok := parser.GetTask("global")
	assert.False(t, ok)
}
