// This file adds a simple wrapper around some os's methods so that
// they can be stubbed easily in a testing scenario.
// FileOSWrapper is the concrete implementation
// MemOSWrapper is the one used in tests
package internal

import (
	"io/fs"
	"os"
	"time"
)

type OSWrapper interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	Getwd() (dir string, err error)
	Stat(name string) (fs.FileInfo, error)
	FileExists(filename string) bool
	Remove(name string) error
	TempDir() string
}

type FileOSWrapper struct{}

type MemOSWrapper struct {
	ReadData []byte
	CWD      string
	Exists   bool
}

func (osw *FileOSWrapper) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (osw *FileOSWrapper) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (osw *FileOSWrapper) Getwd() (dir string, err error) {
	return os.Getwd()
}

func (osw *FileOSWrapper) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (osw *FileOSWrapper) Remove(name string) error {
	return os.Remove(name)
}

func (osw *FileOSWrapper) TempDir() string {
	return os.TempDir()
}

func (osw *FileOSWrapper) FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (osw *MemOSWrapper) ReadFile(name string) ([]byte, error) {
	return osw.ReadData, nil
}

func (osw *MemOSWrapper) WriteFile(name string, data []byte, perm fs.FileMode) error {
	osw.ReadData = data
	return nil
}

func (osw *MemOSWrapper) Getwd() (dir string, err error) {
	return osw.CWD, nil
}

func (osw *MemOSWrapper) Stat(name string) (fs.FileInfo, error) {
	return &MemFileInfo{}, nil
}

func (osw *MemOSWrapper) FileExists(filename string) bool {
	return osw.Exists
}

func (osw *MemOSWrapper) Remove(name string) error {
	return nil
}

func (osw *MemOSWrapper) TempDir() string {
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
