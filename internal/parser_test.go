package internal

import (
	"testing"

	"github.com/dugajean/goke/internal/mocks"
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

var parserBase64encoded = `Sf+BAwEBBlBhcnNlcgH/ggABBAEFVGFza3MB/4gAAQlGaWxlUGF0aHMB/4YAAQpZQU1MQ29uZmlnAQwAAQZHbG9iYWwB/4oAAAAZ/4cEAQEIdGFza0xpc3QB/4gAAQwB/4QAACn/gwMBAv+EAAEDAQROYW1lAQwAAQVGaWxlcwH/hgABA1J1bgH/hgAAABb/hQIBAQhbXXN0cmluZwH/hgABDAAAIP+JAwEBBkdsb2JhbAH/igABAQEGU2hhcmVkAf+MAAAA/gGY/4sDAQH+AWtzdHJ1Y3QgeyBFbnZpcm9ubWVudCBtYXBbc3RyaW5nXXN0cmluZyAieWFtbDpcImVudmlyb25tZW50LG9taXRlbXB0eVwiIjsgRXZlbnRzIHN0cnVjdCB7IEJlZm9yZUVhY2hSdW4gW11zdHJpbmcgInlhbWw6XCJiZWZvcmVfZWFjaF9ydW4sb21pdGVtcHR5XCIiOyBBZnRlckVhY2hSdW4gW11zdHJpbmcgInlhbWw6XCJhZnRlcl9lYWNoX3J1bixvbWl0ZW1wdHlcIiI7IEJlZm9yZUVhY2hUYXNrIFtdc3RyaW5nICJ5YW1sOlwiYmVmb3JlX2VhY2hfdGFzayxvbWl0ZW1wdHlcIiI7IEFmdGVyRWFjaFRhc2sgW11zdHJpbmcgInlhbWw6XCJhZnRlcl9lYWNoX3Rhc2ssb21pdGVtcHR5XCIiIH0gInlhbWw6XCJldmVudHMsb21pdGVtcHR5XCIiIH0B/4wAAQIBC0Vudmlyb25tZW50Af+OAAEGRXZlbnRzAf+QAAAAIf+NBAEBEW1hcFtzdHJpbmddc3RyaW5nAf+OAAEMAQwAAP4BWP+PAwEB//1zdHJ1Y3QgeyBCZWZvcmVFYWNoUnVuIFtdc3RyaW5nICJ5YW1sOlwiYmVmb3JlX2VhY2hfcnVuLG9taXRlbXB0eVwiIjsgQWZ0ZXJFYWNoUnVuIFtdc3RyaW5nICJ5YW1sOlwiYWZ0ZXJfZWFjaF9ydW4sb21pdGVtcHR5XCIiOyBCZWZvcmVFYWNoVGFzayBbXXN0cmluZyAieWFtbDpcImJlZm9yZV9lYWNoX3Rhc2ssb21pdGVtcHR5XCIiOyBBZnRlckVhY2hUYXNrIFtdc3RyaW5nICJ5YW1sOlwiYWZ0ZXJfZWFjaF90YXNrLG9taXRlbXB0eVwiIiB9Af+QAAEEAQ1CZWZvcmVFYWNoUnVuAf+GAAEMQWZ0ZXJFYWNoUnVuAf+GAAEOQmVmb3JlRWFjaFRhc2sB/4YAAQ1BZnRlckVhY2hUYXNrAf+GAAAA/gMv/4IBBQZnbG9iYWwBBmdsb2JhbAAGZXZlbnRzAQZldmVudHMAC2dyZWV0LWxpc2hhAQtncmVldC1saXNoYQIBE2VjaG8gJ0hlbGxvIExpc2hhIScACmdyZWV0LWxva2kBCmdyZWV0LWxva2kCARFlY2hvICJIZWxsbyBCb2tpIgAKZ3JlZXQtY2F0cwEKZ3JlZXQtY2F0cwEBD2NtZC9jbGkvbWFpbi5nbwEDEWVjaG8gIkhlbGxvIEZyZXkiEmVjaG8gIkhlbGxvIFN1bm55IgpncmVldC1sb2tpAAEBD2NtZC9jbGkvbWFpbi5nbwH+AhIKZ2xvYmFsOgogIGVudmlyb25tZW50OgogICAgRk9POiAiZm9vIgogICAgQkFSOiAiJChlY2hvICdiYXInKSIKICAgIEJBWjogImJheiIKCmV2ZW50czoKICBiZWZvcmVfZWFjaF9ydW46CiAgICAtICJlY2hvICdiZWZvcmUgZWFjaCAxJyIKICAgIC0gImVjaG8gJ2JlZm9yZSBlYWNoIDInIgogIGFmdGVyX2VhY2hfcnVuOgogICAgLSAiZWNobyAnYWZ0ZXIgZWFjaCAxJyIKICAgIC0gImdyZWV0LWxpc2hhIgogIGJlZm9yZV9lYWNoX3Rhc2s6CiAgICAtICJlY2hvICdiZWZvcmUgdGFzayciCiAgYWZ0ZXJfZWFjaF90YXNrOgogICAgLSAiZWNobyAnYWZ0ZXIgdGFzayciCgpncmVldC1saXNoYToKICBydW46CiAgICAtICJlY2hvICdIZWxsbyBMaXNoYSEnIgoKZ3JlZXQtbG9raToKICBydW46CiAgICAtICdlY2hvICJIZWxsbyBCb2tpIicKCmdyZWV0LWNhdHM6CiAgZmlsZXM6IFtjbWQvY2xpLypdCiAgcnVuOgogICAgLSAnZWNobyAiSGVsbG8gRnJleSInCiAgICAtICdlY2hvICJIZWxsbyBTdW5ueSInCiAgICAtICJncmVldC1sb2tpIgEBAQMDQkFaA2JhegNGT08DZm9vA0JBUg0kKGVjaG8gJ2JhcicpAQAAAAA=`

var options = Options{
	ClearCache: true,
}

func TestNewParser(t *testing.T) {
	mockStdlib := getMock(t, true)

	parser := NewParser(yamlConfigStub, &options, mockStdlib)
	assert.NotNil(t, parser)
}

func TestTaskParsing(t *testing.T) {
	mockStdlib := getMock(t, true)

	parser := NewParser(yamlConfigStub, &options, mockStdlib)
	assert.NotNil(t, parser.Tasks)

	parser.parseTasks()
	assert.NotNil(t, parser.Tasks)

	assert.NotNil(t, parser.Tasks["greet-loki"])
	assert.NotNil(t, parser.Tasks["greet-cats"])
	assert.NotNil(t, parser.Tasks["greet-lisha"])
}

func TestGlobalsParsing(t *testing.T) {
	mockStdlib := mocks.NewStdlibWrapper(t)
	mockStdlib.On("TempDir").Return("path/to/temp")
	mockStdlib.On("Getwd").Return("path/to/cwd", nil)
	mockStdlib.On("FileExists", mock.Anything).Return(false)
	mockStdlib.On("Setenv", mock.Anything, mock.Anything).Return(nil)

	parser := NewParser(yamlConfigStub, &options, mockStdlib)

	parser.parseGlobal()
	assert.Equal(t, "foo", parser.Global.Shared.Environment["FOO"])
	assert.Equal(t, "$(echo 'bar')", parser.Global.Shared.Environment["BAR"])
	assert.Equal(t, "baz", parser.Global.Shared.Environment["BAZ"])
	mockStdlib.AssertNumberOfCalls(t, "Setenv", 3)
}

func TestTaskFilesExpansion(t *testing.T) {
	mockStdlib := getMock(t, true)

	parser := NewParser(yamlConfigStub, &options, mockStdlib)
	parser.parseTasks()

	greetCatsTask := parser.Tasks["greet-cats"]
	assert.NotNil(t, greetCatsTask)
}

func TestParserWithoutCache(t *testing.T) {
	mockStdlib := mocks.NewStdlibWrapper(t)
	mockStdlib.On("TempDir").Return("path/to/temp")
	mockStdlib.On("Getwd").Return("path/to/cwd", nil)
	mockStdlib.On("FileExists", mock.Anything).Return(false)

	parser := NewParser(yamlConfigStub, &options, mockStdlib)
	parser.parseTasks()

	greetCatsTask := parser.Tasks["greet-cats"]
	assert.NotNil(t, greetCatsTask)
	mockStdlib.AssertExpectations(t)
	mockStdlib.AssertNumberOfCalls(t, "TempDir", 1)
	mockStdlib.AssertNumberOfCalls(t, "Getwd", 1)
	mockStdlib.AssertNumberOfCalls(t, "FileExists", 2)
}

func getMock(t *testing.T, fileExists bool) StdlibWrapper {
	mockStdlib := mocks.NewStdlibWrapper(t)

	mockStdlib.On("TempDir").Return("path/to/temp")
	mockStdlib.On("Getwd").Return("path/to/cwd", nil)
	mockStdlib.On("FileExists", mock.Anything).Return(fileExists)
	mockStdlib.On("Remove", mock.Anything).Return(nil)
	mockStdlib.On("ReadFile", mock.Anything).Return([]byte(parserBase64encoded), nil)

	return mockStdlib
}
