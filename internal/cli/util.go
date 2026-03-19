package cli

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var osCommandRegexp = regexp.MustCompile(`\$\((.+)\)`)
var osEnvRegexp = regexp.MustCompile(`\$\{(.+)\}`)

// Parses the interpolated system commands, ie. "Hello $(echo 'World')" and returns it.
// Returns the command wrapper in $() and without the wrapper.
func parseSystemCmd(re *regexp.Regexp, str string) (string, string) {
	match := re.FindAllStringSubmatch(str, -1)

	if len(match) > 0 && len(match[0]) > 0 {
		return match[0][0], match[0][1]
	}

	return "", ""
}

// prase system commands and store results to env
func SetEnvVariables(vars map[string]string) (map[string]string, error) {
	retVars := make(map[string]string)
	for k, v := range vars {
		_, cmd := parseSystemCmd(osCommandRegexp, v)

		if cmd == "" {
			retVars[k] = v
			_ = os.Setenv(k, v)
			continue
		}

		splitCmd, err := ParseCommandLine(os.ExpandEnv(cmd))
		if err != nil {
			return retVars, err
		}

		out, err := exec.Command(splitCmd[0], splitCmd[1:]...).Output()
		if err != nil {
			return retVars, err
		}

		outStr := strings.TrimSpace(string(out))
		retVars[k] = outStr
		_ = os.Setenv(k, outStr)
	}

	return retVars, nil
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

// Replace the placeholders with actual environment variable values in string pointer.
// Given that a string pointer must be provided, the replacement happens in place.
func ReplaceEnvironmentVariables(re *regexp.Regexp, str *string) {
	resolved := *str
	raw, env := parseSystemCmd(re, resolved)

	if raw != "" && env != "" {
		*str = strings.ReplaceAll(resolved, raw, os.Getenv(env))
	}
}
