package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSystemCmd(t *testing.T) {
	cmds := []string{
		"$(echo 'Hello Thor')",
		"hello world",
	}

	want := [][]string{
		{"$(echo 'Hello Thor')", "echo 'Hello Thor'"},
		{"", ""},
	}

	for i, cmd := range cmds {
		got0, got1 := parseSystemCmd(osCommandRegexp, cmd)
		assert.Equal(t, want[i][0], got0, "expected "+want[i][0]+", got ", got0)
		assert.Equal(t, want[i][1], got1, "expected "+want[i][1]+", got ", got1)
	}
}

func TestSetEnvVariables(t *testing.T) {
	values := map[string]string{
		"THOR":     "Lord of thunder",
		"THOR_CMD": "$(echo 'Hello Thor')",
	}

	want := map[string]string{
		"THOR":     "Lord of thunder",
		"THOR_CMD": "Hello Thor",
	}

	got, _ := SetEnvVariables(values)
	assert.Equal(t, want["THOR"], os.Getenv("THOR"))
	assert.Equal(t, want["THOR_CMD"], os.Getenv("THOR_CMD"))

	for k := range got {
		assert.Equal(t, want[k], got[k])
	}
}

func TestParseCommandLineSimple(t *testing.T) {
	args, err := ParseCommandLine("echo hello world")
	require.NoError(t, err)
	assert.Equal(t, []string{"echo", "hello", "world"}, args)
}

func TestParseCommandLineDoubleQuotes(t *testing.T) {
	args, err := ParseCommandLine(`echo "hello world" foo`)
	require.NoError(t, err)
	assert.Equal(t, []string{"echo", "hello world", "foo"}, args)
}

func TestParseCommandLineSingleQuotes(t *testing.T) {
	args, err := ParseCommandLine("echo 'hello world' foo")
	require.NoError(t, err)
	assert.Equal(t, []string{"echo", "hello world", "foo"}, args)
}

func TestParseCommandLineEscapedCharacters(t *testing.T) {
	args, err := ParseCommandLine(`echo hello\ world`)
	require.NoError(t, err)
	assert.Equal(t, []string{"echo", "hello world"}, args)
}

func TestParseCommandLineUnclosedQuote(t *testing.T) {
	_, err := ParseCommandLine(`echo "hello world`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unclosed quote")
}

func TestParseCommandLineEmpty(t *testing.T) {
	args, err := ParseCommandLine("")
	require.NoError(t, err)
	assert.Empty(t, args)
}

func TestParseCommandLineMultipleTabs(t *testing.T) {
	args, err := ParseCommandLine("echo\t\thello\tworld")
	require.NoError(t, err)
	assert.Equal(t, []string{"echo", "hello", "world"}, args)
}

func TestReplaceEnvironmentVariables(t *testing.T) {
	values := map[string]string{
		"THOR": "Lord of thunder",
		"LOKI": "Lord of deception",
	}

	for k, v := range values {
		t.Setenv(k, v)
	}

	str := "I am ${THOR}"
	want := "I am Lord of thunder"

	ReplaceEnvironmentVariables(osEnvRegexp, &str)

	assert.Equal(t, want, str, "wrong env value is injected")
}

func TestReplaceEnvironmentVariablesNoMatch(t *testing.T) {
	str := "no variables here"
	original := str

	ReplaceEnvironmentVariables(osEnvRegexp, &str)

	assert.Equal(t, original, str)
}
