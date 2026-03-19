package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateDependenciesNoCycles(t *testing.T) {
	tasks := taskList{
		"build": {Run: []string{"go build"}, DependsOn: []string{"clean"}},
		"clean": {Run: []string{"rm -rf build"}},
		"test":  {Run: []string{"go test"}, DependsOn: []string{"build"}},
	}

	err := ValidateDependencies(tasks)
	require.NoError(t, err)
}

func TestValidateDependenciesDirectCycle(t *testing.T) {
	tasks := taskList{
		"a": {Run: []string{"echo a"}, DependsOn: []string{"b"}},
		"b": {Run: []string{"echo b"}, DependsOn: []string{"a"}},
	}

	err := ValidateDependencies(tasks)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestValidateDependenciesSelfCycle(t *testing.T) {
	tasks := taskList{
		"a": {Run: []string{"echo a"}, DependsOn: []string{"a"}},
	}

	err := ValidateDependencies(tasks)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestValidateDependenciesTransitiveCycle(t *testing.T) {
	tasks := taskList{
		"a": {Run: []string{"echo a"}, DependsOn: []string{"b"}},
		"b": {Run: []string{"echo b"}, DependsOn: []string{"c"}},
		"c": {Run: []string{"echo c"}, DependsOn: []string{"a"}},
	}

	err := ValidateDependencies(tasks)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestValidateDependenciesMissingTask(t *testing.T) {
	tasks := taskList{
		"a": {Run: []string{"echo a"}, DependsOn: []string{"nonexistent"}},
	}

	err := ValidateDependencies(tasks)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown task 'nonexistent'")
}

func TestValidateDependenciesDiamond(t *testing.T) {
	tasks := taskList{
		"a": {Run: []string{"echo a"}, DependsOn: []string{"b", "c"}},
		"b": {Run: []string{"echo b"}, DependsOn: []string{"d"}},
		"c": {Run: []string{"echo c"}, DependsOn: []string{"d"}},
		"d": {Run: []string{"echo d"}},
	}

	err := ValidateDependencies(tasks)
	require.NoError(t, err)
}

func TestValidateDependenciesEmpty(t *testing.T) {
	tasks := taskList{
		"a": {Run: []string{"echo a"}},
		"b": {Run: []string{"echo b"}},
	}

	err := ValidateDependencies(tasks)
	require.NoError(t, err)
}
