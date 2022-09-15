package internal

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"
)

type Options struct {
	ClearCache bool
	Watch      bool
	Force      bool
}

func ReadYamlConfig() string {
	content, err := os.ReadFile("goke.yml")

	if err != nil {
		fmt.Println("No presence of goke.yml sighted")
		os.Exit(1)
	}

	return string(content)
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Serialize a struct
func GOBSerialize[T any](structInstance T) string {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(structInstance)

	if err != nil {
		log.Fatal("failed gob encode", err)
	}

	return base64.StdEncoding.EncodeToString(b.Bytes())
}

// Deserialize a struct
func GOBDeserialize[T any](structStr string, structShell *T) T {
	by, err := base64.StdEncoding.DecodeString(structStr)

	if err != nil {
		log.Fatal("failed base64 decode", err)
	}

	b := bytes.Buffer{}
	b.Write(by)
	d := gob.NewDecoder(&b)
	err = d.Decode(structShell)

	if err != nil {
		log.Fatal("failed gob decode", err)
	}

	return *structShell
}

func PermutateArgs(args []string) int {
	args = args[1:]
	optind := 0

	for i := range args {
		if args[i][0] == '-' {
			tmp := args[i]
			args[i] = args[optind]
			args[optind] = tmp
			optind++
		}
	}

	return optind + 1
}

// Parses the command string into an array of [command, args, args]...
func ParseCommandLine(command string) ([]string, error) {
	var args []string
	state := "start"
	current := ""
	quote := "\""
	escapeNext := true

	for i := 0; i < len(command); i++ {
		c := command[i]

		if state == "quotes" {
			if string(c) != quote {
				current += string(c)
			} else {
				args = append(args, current)
				current = ""
				state = "start"
			}
			continue
		}

		if escapeNext {
			current += string(c)
			escapeNext = false
			continue
		}

		if c == '\\' {
			escapeNext = true
			continue
		}

		if c == '"' || c == '\'' {
			state = "quotes"
			quote = string(c)
			continue
		}

		if state == "arg" {
			if c == ' ' || c == '\t' {
				args = append(args, current)
				current = ""
				state = "start"
			} else {
				current += string(c)
			}
			continue
		}

		if c != ' ' && c != '\t' {
			state = "arg"
			current += string(c)
		}
	}

	if state == "quotes" {
		return []string{}, fmt.Errorf("unclosed quote in command: %s", command)
	}

	if current != "" {
		args = append(args, current)
	}

	return args, nil
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
