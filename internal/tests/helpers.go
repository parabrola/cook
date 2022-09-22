package tests

import (
	fs "io/fs"
	"testing"
	"time"

	mock "github.com/stretchr/testify/mock"
)

const ReadFileBase64 = "Sf+BAwEBBlBhcnNlcgH/ggABBAEFVGFza3MB/4gAAQlGaWxlUGF0aHMB/4YAAQpZQU1MQ29uZmlnAQwAAQZHbG9iYWwB/4oAAAAZ/4cEAQEIdGFza0xpc3QB/4gAAQwB/4QAACn/gwMBAv+EAAEDAQROYW1lAQwAAQVGaWxlcwH/hgABA1J1bgH/hgAAABb/hQIBAQhbXXN0cmluZwH/hgABDAAAIP+JAwEBBkdsb2JhbAH/igABAQEGU2hhcmVkAf+MAAAA/gGY/4sDAQH+AWtzdHJ1Y3QgeyBFbnZpcm9ubWVudCBtYXBbc3RyaW5nXXN0cmluZyAieWFtbDpcImVudmlyb25tZW50LG9taXRlbXB0eVwiIjsgRXZlbnRzIHN0cnVjdCB7IEJlZm9yZUVhY2hSdW4gW11zdHJpbmcgInlhbWw6XCJiZWZvcmVfZWFjaF9ydW4sb21pdGVtcHR5XCIiOyBBZnRlckVhY2hSdW4gW11zdHJpbmcgInlhbWw6XCJhZnRlcl9lYWNoX3J1bixvbWl0ZW1wdHlcIiI7IEJlZm9yZUVhY2hUYXNrIFtdc3RyaW5nICJ5YW1sOlwiYmVmb3JlX2VhY2hfdGFzayxvbWl0ZW1wdHlcIiI7IEFmdGVyRWFjaFRhc2sgW11zdHJpbmcgInlhbWw6XCJhZnRlcl9lYWNoX3Rhc2ssb21pdGVtcHR5XCIiIH0gInlhbWw6XCJldmVudHMsb21pdGVtcHR5XCIiIH0B/4wAAQIBC0Vudmlyb25tZW50Af+OAAEGRXZlbnRzAf+QAAAAIf+NBAEBEW1hcFtzdHJpbmddc3RyaW5nAf+OAAEMAQwAAP4BWP+PAwEB//1zdHJ1Y3QgeyBCZWZvcmVFYWNoUnVuIFtdc3RyaW5nICJ5YW1sOlwiYmVmb3JlX2VhY2hfcnVuLG9taXRlbXB0eVwiIjsgQWZ0ZXJFYWNoUnVuIFtdc3RyaW5nICJ5YW1sOlwiYWZ0ZXJfZWFjaF9ydW4sb21pdGVtcHR5XCIiOyBCZWZvcmVFYWNoVGFzayBbXXN0cmluZyAieWFtbDpcImJlZm9yZV9lYWNoX3Rhc2ssb21pdGVtcHR5XCIiOyBBZnRlckVhY2hUYXNrIFtdc3RyaW5nICJ5YW1sOlwiYWZ0ZXJfZWFjaF90YXNrLG9taXRlbXB0eVwiIiB9Af+QAAEEAQ1CZWZvcmVFYWNoUnVuAf+GAAEMQWZ0ZXJFYWNoUnVuAf+GAAEOQmVmb3JlRWFjaFRhc2sB/4YAAQ1BZnRlckVhY2hUYXNrAf+GAAAA/gMv/4IBBQZnbG9iYWwBBmdsb2JhbAAGZXZlbnRzAQZldmVudHMAC2dyZWV0LWxpc2hhAQtncmVldC1saXNoYQIBE2VjaG8gJ0hlbGxvIExpc2hhIScACmdyZWV0LWxva2kBCmdyZWV0LWxva2kCARFlY2hvICJIZWxsbyBCb2tpIgAKZ3JlZXQtY2F0cwEKZ3JlZXQtY2F0cwEBD2NtZC9jbGkvbWFpbi5nbwEDEWVjaG8gIkhlbGxvIEZyZXkiEmVjaG8gIkhlbGxvIFN1bm55IgpncmVldC1sb2tpAAEBD2NtZC9jbGkvbWFpbi5nbwH+AhIKZ2xvYmFsOgogIGVudmlyb25tZW50OgogICAgRk9POiAiZm9vIgogICAgQkFSOiAiJChlY2hvICdiYXInKSIKICAgIEJBWjogImJheiIKCmV2ZW50czoKICBiZWZvcmVfZWFjaF9ydW46CiAgICAtICJlY2hvICdiZWZvcmUgZWFjaCAxJyIKICAgIC0gImVjaG8gJ2JlZm9yZSBlYWNoIDInIgogIGFmdGVyX2VhY2hfcnVuOgogICAgLSAiZWNobyAnYWZ0ZXIgZWFjaCAxJyIKICAgIC0gImdyZWV0LWxpc2hhIgogIGJlZm9yZV9lYWNoX3Rhc2s6CiAgICAtICJlY2hvICdiZWZvcmUgdGFzayciCiAgYWZ0ZXJfZWFjaF90YXNrOgogICAgLSAiZWNobyAnYWZ0ZXIgdGFzayciCgpncmVldC1saXNoYToKICBydW46CiAgICAtICJlY2hvICdIZWxsbyBMaXNoYSEnIgoKZ3JlZXQtbG9raToKICBydW46CiAgICAtICdlY2hvICJIZWxsbyBCb2tpIicKCmdyZWV0LWNhdHM6CiAgZmlsZXM6IFtjbWQvY2xpLypdCiAgcnVuOgogICAgLSAnZWNobyAiSGVsbG8gRnJleSInCiAgICAtICdlY2hvICJIZWxsbyBTdW5ueSInCiAgICAtICJncmVldC1sb2tpIgEBAQMDQkFaA2JhegNGT08DZm9vA0JBUg0kKGVjaG8gJ2JhcicpAQAAAAA="

func GetFileSystemMock(t *testing.T) any {
	fsMock := NewFileSystem(t)

	fsMock.On("TempDir").Return("path/to/temp")
	fsMock.On("Getwd").Return("path/to/cwd", nil)
	fsMock.On("FileExists", mock.Anything).Return(true)
	fsMock.On("Remove", mock.Anything).Return(nil)
	fsMock.On("Stat", mock.Anything).Return(MemFileInfo{}, nil)
	fsMock.On("WriteFile", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	fsMock.On("ReadFile", mock.Anything).Return([]byte(ReadFileBase64), nil)

	return fsMock
}

type MemFileInfo struct{}

func (fi MemFileInfo) Name() string {
	return "foo"
}

func (fi MemFileInfo) Size() int64 {
	return 10000
}

func (fi MemFileInfo) Mode() fs.FileMode {
	return 0644
}

func (fi MemFileInfo) ModTime() time.Time {
	return time.Date(2022, time.December, 24, 1, 1, 1, 1, time.UTC)
}

func (fi MemFileInfo) IsDir() bool {
	return false
}

func (fi MemFileInfo) Sys() any {
	return nil
}
