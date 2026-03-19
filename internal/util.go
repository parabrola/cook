package internal

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"strings"
)

func GokeFiles() []string {
	return []string{"goke.yml", "goke.yaml"}
}

func CurrentConfigFile() string {
	for _, f := range GokeFiles() {
		if FileExists(f) {
			return f
		}
	}

	return ""
}

func ReadYamlConfig() (string, error) {
	for _, f := range GokeFiles() {
		content, err := os.ReadFile(f)

		if err == nil && len(content) > 0 {
			return string(content), nil
		}
	}

	return "", errors.New("no presence of goke.yml sighted")
}

func CreateGokeConfig() error {
	const sampleConfig = `global:
environment:
  MY_BINARY: "my_binary"

build: 
  files: [cmd/cli/*.go, internal/*]
  run:
    - "go build -o ./build/${MY_BINARY} ./cmd/cli"
`

	for _, f := range GokeFiles() {
		if FileExists(f) {
			return fmt.Errorf("%s already present in this directory", f)
		}
	}

	return os.WriteFile("goke.yml", []byte(sampleConfig), 0644)
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func GOBSerialize[T any](structInstance T) string {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	if err := e.Encode(structInstance); err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(b.Bytes())
}

func GOBDeserialize[T any](structStr string, structShell *T) (T, error) {
	by, err := base64.StdEncoding.DecodeString(structStr)
	if err != nil {
		return *structShell, fmt.Errorf("failed base64 decode: %w", err)
	}

	b := bytes.Buffer{}
	b.Write(by)
	d := gob.NewDecoder(&b)
	if err = d.Decode(structShell); err != nil {
		return *structShell, fmt.Errorf("failed gob decode: %w", err)
	}

	return *structShell, nil
}

func JoinInnerArgs(args []string) string {
	return strings.Join(args, " ")
}
