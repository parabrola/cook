package internal

import (
	"testing"

	"github.com/dugajean/goke/internal/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var files = []string{"./lockfile.go"}

var lockfileOpts = Options{
	ClearCache: true,
}

func TestNewLockfile(t *testing.T) {
	fsMock := tests.NewFileSystem(t)
	lockfile := NewLockfile(files, &lockfileOpts, fsMock)

	assert.NotNil(t, lockfile)
	assert.Equal(t, files, lockfile.files)
}

func TestGenerateLockfileWithTrue(t *testing.T) {
	fsMock := tests.NewFileSystem(t)
	fsMock.On("Getwd").Return("path/to/cwd", nil)
	fsMock.On("Stat", mock.Anything).Return(tests.MemFileInfo{}, nil)
	fsMock.On("WriteFile", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	lockfile := NewLockfile(files, &lockfileOpts, fsMock)
	err := lockfile.generateLockfile(true)

	assert.Nil(t, err)
}

func TestGenerateLockfileWithFalse(t *testing.T) {
	fsMock := tests.NewFileSystem(t)
	fsMock.On("Getwd").Return("path/to/cwd", nil)
	fsMock.On("Stat", mock.Anything).Return(tests.MemFileInfo{}, nil)
	fsMock.On("WriteFile", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	// fsMock.On("FileExists", mock.Anything).Return(false)

	lockfile := NewLockfile(files, &lockfileOpts, fsMock)
	err := lockfile.generateLockfile(true)

	assert.Nil(t, err)
}
