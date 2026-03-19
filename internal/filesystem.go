// This file adds a simple wrapper around some os's methods so that
// they can be stubbed easily in a testing scenario.
// LocalFileSystem is the concrete implementation
// MemFileSystem is the one used in tests
package internal

import (
	"io/fs"
	"os"
	"path/filepath"
)

type FileSystem interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	Getwd() (dir string, err error)
	Stat(name string) (fs.FileInfo, error)
	FileExists(filename string) bool
	Remove(name string) error
	TempDir() string
	Glob(path string) ([]string, error)
}

type LocalFileSystem struct{}

func (fs *LocalFileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (fs *LocalFileSystem) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (fs *LocalFileSystem) Getwd() (dir string, err error) {
	return os.Getwd()
}

func (fs *LocalFileSystem) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (fs *LocalFileSystem) Remove(name string) error {
	return os.Remove(name)
}

func (fs *LocalFileSystem) TempDir() string {
	return os.TempDir()
}

func (fs *LocalFileSystem) FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func (fs *LocalFileSystem) Glob(path string) ([]string, error) {
	return filepath.Glob(path)
}
