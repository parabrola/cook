// This file adds a simple wrapper around some os's methods so that
// they can be stubbed easily in a testing scenario.
// ConcreteStdlibWrapper is the concrete implementation
// MemStdlibWrapper is the one used in tests
package internal

import (
	"io/fs"
	"os"
	"time"
)

type StdlibWrapper interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	Getwd() (dir string, err error)
	Stat(name string) (fs.FileInfo, error)
	FileExists(filename string) bool
	Remove(name string) error
	TempDir() string
	Setenv(key string, value string) error
}

type ConcreteStdlibWrapper struct{}

type MemStdlibWrapper struct {
	ReadData []byte
	CWD      string
	Exists   bool
}

func (std *ConcreteStdlibWrapper) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (std *ConcreteStdlibWrapper) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (std *ConcreteStdlibWrapper) Getwd() (dir string, err error) {
	return os.Getwd()
}

func (std *ConcreteStdlibWrapper) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (std *ConcreteStdlibWrapper) Remove(name string) error {
	return os.Remove(name)
}

func (std *ConcreteStdlibWrapper) TempDir() string {
	return os.TempDir()
}

func (std *ConcreteStdlibWrapper) FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (std *ConcreteStdlibWrapper) Setenv(key string, value string) error {
	return os.Setenv(key, value)
}

func (std *MemStdlibWrapper) ReadFile(name string) ([]byte, error) {
	return std.ReadData, nil
}

func (std *MemStdlibWrapper) WriteFile(name string, data []byte, perm fs.FileMode) error {
	std.ReadData = data
	return nil
}

func (std *MemStdlibWrapper) Getwd() (dir string, err error) {
	return std.CWD, nil
}

func (std *MemStdlibWrapper) Stat(name string) (fs.FileInfo, error) {
	return &MemFileInfo{}, nil
}

func (std *MemStdlibWrapper) FileExists(filename string) bool {
	return std.Exists
}

func (std *MemStdlibWrapper) Remove(name string) error {
	return nil
}

func (std *MemStdlibWrapper) TempDir() string {
	return "path/to/temp/dir"
}

type MemFileInfo struct{}

func (fi *MemFileInfo) Name() string {
	return "foo"
}

func (fi *MemFileInfo) Size() int64 {
	return 10000
}

func (fi *MemFileInfo) Mode() fs.FileMode {
	return 0644
}

func (fi *MemFileInfo) ModTime() time.Time {
	return time.Date(2022, time.December, 24, 1, 1, 1, 1, time.UTC)
}

func (fi *MemFileInfo) IsDir() bool {
	return false
}

func (fi *MemFileInfo) Sys() any {
	return nil
}
