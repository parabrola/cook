// This file adds a simple wrapper around some os's methods so that
// they can be stubbed easily in a testing scenario.
// LocalFileSystem is the concrete implementation
// MemFileSystem is the one used in tests
package internal

import (
	"io/fs"
	"os"
)

type FileSystem interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	Getwd() (dir string, err error)
	Stat(name string) (fs.FileInfo, error)
	FileExists(filename string) bool
	Remove(name string) error
	TempDir() string
}

type LocalFileSystem struct{}

func (std *LocalFileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (std *LocalFileSystem) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (std *LocalFileSystem) Getwd() (dir string, err error) {
	return os.Getwd()
}

func (std *LocalFileSystem) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (std *LocalFileSystem) Remove(name string) error {
	return os.Remove(name)
}

func (std *LocalFileSystem) TempDir() string {
	return os.TempDir()
}

func (std *LocalFileSystem) FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
