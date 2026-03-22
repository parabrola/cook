package internal

import (
	"testing"

	"github.com/parabrola/cook/internal/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var files = []string{"./lockfile.go"}

var lockfileOpts = Options{
	NoCache: true,
}

var dotCookFile = `{
  "/path/to/project1": {
    "path/to/file": 1664738433
  },
  "/path/to/project2": {
    "./path/to/file": 1663812584
  }
}`

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
	fsMock.On("WriteFile", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	lockfile := NewLockfile(files, &lockfileOpts, fsMock)
	lockfile.JSON = lockFileJson{
		"path/to/cwd": {"./lockfile.go": 1664738433},
	}
	err := lockfile.generateLockfile(false)

	assert.Nil(t, err)
}

func TestBootstrapLoadsExistingLockfile(t *testing.T) {
	fsMock := tests.NewFileSystem(t)
	fsMock.On("FileExists", mock.Anything).Return(true)
	fsMock.On("ReadFile", mock.Anything).Return([]byte(dotCookFile), nil)

	lockfile := NewLockfile(files, &lockfileOpts, fsMock)
	require.NoError(t, lockfile.Bootstrap())

	assert.NotNil(t, lockfile.JSON)
	assert.Contains(t, lockfile.JSON, "/path/to/project1")
	assert.Contains(t, lockfile.JSON, "/path/to/project2")
}

func TestGetCurrentProject(t *testing.T) {
	fsMock := tests.NewFileSystem(t)
	fsMock.On("Getwd").Return("/path/to/project1", nil)
	fsMock.On("FileExists", mock.Anything).Return(true)
	fsMock.On("ReadFile", mock.Anything).Return([]byte(dotCookFile), nil)

	lockfile := NewLockfile(files, &lockfileOpts, fsMock)
	require.NoError(t, lockfile.Bootstrap())

	project := lockfile.GetCurrentProject()
	assert.NotNil(t, project)
	assert.Equal(t, int64(1664738433), project["path/to/file"])
}

func TestUpdateTimestampsForFiles(t *testing.T) {
	fsMock := tests.NewFileSystem(t)
	fsMock.On("Getwd").Return("/path/to/project1", nil)
	fsMock.On("FileExists", mock.Anything).Return(true)
	fsMock.On("ReadFile", mock.Anything).Return([]byte(dotCookFile), nil)
	fsMock.On("Stat", mock.Anything).Return(tests.MemFileInfo{}, nil)
	fsMock.On("WriteFile", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	lockfile := NewLockfile(files, &lockfileOpts, fsMock)
	require.NoError(t, lockfile.Bootstrap())

	err := lockfile.UpdateTimestampsForFiles([]string{"path/to/file"})
	assert.Nil(t, err)

	project := lockfile.GetCurrentProject()
	assert.NotNil(t, project["path/to/file"])
}
