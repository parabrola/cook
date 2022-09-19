package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var files = []string{"./lockfile.go"}

func TestNewLockfile(t *testing.T) {
	lockfile := NewLockfile(files)

	assert.NotNil(t, lockfile)
	assert.Equal(t, files, lockfile.files)
}

func TestGenerateLockfileWithTrue(t *testing.T) {
	lockfile := NewLockfile(files)
	err := lockfile.generateLockfile(true)

	assert.Nil(t, err)
}

func TestGenerateLockfileWithFalse(t *testing.T) {
	lockfile := NewLockfile(files)
	err := lockfile.generateLockfile(true)

	assert.Nil(t, err)
}
