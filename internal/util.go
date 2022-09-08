package internal

import (
	"fmt"
	"os"
)

func ReadYamlConfig() string {
	content, err := os.ReadFile("goke.yml")

	if err != nil {
		fmt.Println("no presence of goke sighted")
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
