package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var yamlConfig = ReadYamlConfig()

type command struct {
	Files []string `yaml:"files,omitempty"`
	Run   []string `yaml:"run"`
}

type global struct {
	Global struct {
		Environment map[string]string `yaml:"environment,omitempty"`
		Events      struct {
			BeforeEach []string `yaml:"before_each,omitempty"`
			AfterEach  []string `yaml:"after_each,omitempty"`
		} `yaml:"events,omitempty"`
	} `yaml:"global,omitempty"`
}

type commandList map[string]command

const osCommandRegexp = `\$\(([\w\d]+)\)`

type Parser struct {
	Commands  commandList
	FilePaths []string
	global
}

func (p *Parser) Bootstrap() {
	p.parseGlobal()
	p.parseCommands()
}

// Parses the individual user defined commands in the YAML config,
// and processes the dynamic parts of both "run" and "files" sections.
func (p *Parser) parseCommands() {
	var all commandList

	if err := yaml.Unmarshal([]byte(yamlConfig), &all); err != nil {
		fmt.Println(err.Error())
	}

	re := regexp.MustCompile(osCommandRegexp)
	filePaths := []string{}

	for k, c := range all {
		for i := range c.Files {
			p.replaceEnvironmentVariables(re, &all[k].Files[i])
			filePaths = append(filePaths, p.expandFilePaths(all[k].Files[i])...)
		}

		for i, r := range c.Run {
			all[k].Run[i] = strings.Replace(r, "$(files)", strings.Join(c.Files, " "), -1)
			p.replaceEnvironmentVariables(re, &all[k].Run[i])
		}
	}

	p.FilePaths = filePaths
	p.Commands = all
}

// Parses the "global" key in the yaml config and adds it to the parser.
// Also sets all variables under global.environment as OS environment variables.
func (p *Parser) parseGlobal() {
	var g global
	if err := yaml.Unmarshal([]byte(yamlConfig), &g); err != nil {
		log.Fatal(err)
	}

	re := regexp.MustCompile(osCommandRegexp)
	for k, v := range g.Global.Environment {
		_, cmd := p.parseSystemCmd(re, v)

		if cmd == "" {
			continue
		}

		out, err := exec.Command(cmd).Output()

		if err != nil {
			continue
		}

		os.Setenv(k, string(out))
	}

	p.Global = g.Global
}

// Parses the interpolated system commands, ie. "Hello $(echo 'World')" and returns it.
// Returns the command wrapper in $() and without the wrapper.
func (p *Parser) parseSystemCmd(re *regexp.Regexp, str string) (string, string) {
	match := re.FindAllStringSubmatch(str, -1)

	if len(match) > 0 && len(match[0]) > 0 {
		return match[0][0], match[0][1]
	}

	return "", ""
}

// Replace the placeholders with actual environment variable values in string pointer.
// Given that a string pointer must be provided, the replacement happens in place.
func (p *Parser) replaceEnvironmentVariables(re *regexp.Regexp, str *string) {
	resolved := *str
	raw, env := p.parseSystemCmd(re, resolved)

	if raw != "" && env != "" {
		*str = strings.Replace(resolved, raw, os.Getenv(env), -1)
	}
}

func (p *Parser) expandFilePaths(file string) []string {
	filePaths := []string{}

	if strings.Contains(file, "*") {
		files, err := filepath.Glob(file)
		if err != nil {
			log.Fatal(err)
		}

		if len(files) > 0 {
			filePaths = append(filePaths, files...)
		}
	} else if FileExists(file) {
		filePaths = append(filePaths, file)
	}

	return filePaths
}
