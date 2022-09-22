package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var files = []string{"./lockfile.go"}

var lockfileOpts = Options{
	ClearCache: true,
}

func TestNewLockfile(t *testing.T) {
	lockfile := NewLockfile(files, &lockfileOpts)

	assert.NotNil(t, lockfile)
	assert.Equal(t, files, lockfile.files)
}

func TestGenerateLockfileWithTrue(t *testing.T) {
	lockfile := NewLockfile(files, &lockfileOpts)
	err := lockfile.generateLockfile(true)

	assert.Nil(t, err)
}

func TestGenerateLockfileWithFalse(t *testing.T) {
	lockfile := NewLockfile(files, &lockfileOpts)
	err := lockfile.generateLockfile(true)

	assert.Nil(t, err)
}
