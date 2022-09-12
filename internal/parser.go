package internal

import (
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Task struct {
	Name  string
	Files []string `yaml:"files,omitempty"`
	Run   []string `yaml:"run"`
}

type Global struct {
	Shared struct {
		Environment map[string]string `yaml:"environment,omitempty"`
		Events      struct {
			BeforeEachRun  []string `yaml:"before_each_run,omitempty"`
			AfterEachRun   []string `yaml:"after_each_run,omitempty"`
			BeforeEachTask []string `yaml:"before_each_task,omitempty"`
			AfterEachTask  []string `yaml:"after_each_task,omitempty"`
		} `yaml:"events,omitempty"`
	} `yaml:"global,omitempty"`
}

type taskList map[string]Task

const osCommandRegexp = `\$\(([\w\d]+)\)`

var parserString string

type Parser struct {
	Tasks      taskList
	FilePaths  []string
	YAMLConfig string
	os         OSWrapper
	Global
}

// Provide a parser instance which can be either a blank one,
// or one provided  from the cache, which gets deserialized.
func NewParser(YAMLConfig string, opts *Options, osw OSWrapper) Parser {
	parser := Parser{}
	parser.os = osw
	parser.YAMLConfig = YAMLConfig

	tempFile := path.Join(osw.TempDir(), parser.getTempFileName())

	// Clear cache if CLI flag was provided.
	if opts.ClearCache && osw.FileExists(tempFile) {
		osw.Remove(tempFile)
	}

	if !osw.FileExists(tempFile) {
		return parser
	}

	pBytes, err := osw.ReadFile(tempFile)
	if err != nil {
		log.Fatal(err)
	}

	pStr := string(pBytes)
	parserString = pStr

	return GOBDeserialize(pStr, &parser)
}

// Do the parsing process or skip if cached.
func (p *Parser) Bootstrap() {
	// Nothing too bootstrap if cached.
	if parserString != "" {
		return
	}

	err := p.parseGlobal()
	if err != nil {
		log.Fatal(err)
	}

	err = p.parseTasks()
	if err != nil {
		log.Fatal(err)
	}

	pStr := GOBSerialize(*p)
	err = p.os.WriteFile(path.Join(p.os.TempDir(), p.getTempFileName()), []byte(pStr), 0644)

	if err != nil {
		log.Fatal(err)
	}
}

// Parses the individual user defined tasks in the YAML config,
// and processes the dynamic parts of both "run" and "files" sections.
func (p *Parser) parseTasks() error {
	var tasks taskList

	if err := yaml.Unmarshal([]byte(p.YAMLConfig), &tasks); err != nil {
		return err
	}

	re := regexp.MustCompile(osCommandRegexp)
	allFilesPaths := []string{}

	for k, c := range tasks {
		filePaths := []string{}
		for i := range c.Files {
			p.replaceEnvironmentVariables(re, &tasks[k].Files[i])
			expanded, err := p.expandFilePaths(tasks[k].Files[i])

			if err != nil {
				return err
			}

			filePaths = append(filePaths, expanded...)
			allFilesPaths = append(allFilesPaths, expanded...)
		}

		c.Files = filePaths
		tasks[k] = c

		for i, r := range c.Run {
			tasks[k].Run[i] = strings.Replace(r, "$(files)", strings.Join(c.Files, " "), -1)
			p.replaceEnvironmentVariables(re, &tasks[k].Run[i])
		}

		c.Name = k
		tasks[k] = c
	}

	p.FilePaths = allFilesPaths
	p.Tasks = tasks

	return nil
}

// Parses the "global" key in the yaml config and adds it to the parser.
// Also sets all variables under global.environment as OS environment variables.
func (p *Parser) parseGlobal() error {
	var g Global

	if err := yaml.Unmarshal([]byte(p.YAMLConfig), &g); err != nil {
		return err
	}

	re := regexp.MustCompile(osCommandRegexp)
	for k, v := range g.Shared.Environment {
		_, cmd := p.parseSystemCmd(re, v)

		if cmd == "" {
			continue
		}

		out, err := exec.Command(cmd).Output()
		if err != nil {
			continue
		}

		g.Shared.Environment[k] = string(out)
		os.Setenv(k, string(out))
	}

	p.Global = g

	return nil
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

func (p *Parser) expandFilePaths(file string) ([]string, error) {
	filePaths := []string{}

	if strings.Contains(file, "*") {
		files, err := filepath.Glob(file)
		if err != nil {
			return nil, err
		}

		if len(files) > 0 {
			filePaths = append(filePaths, files...)
		}
	} else if FileExists(file) {
		filePaths = append(filePaths, file)
	}

	return filePaths, nil
}

func (p *Parser) getTempFileName() string {
	cwd, _ := p.os.Getwd()
	return "goke-" + strings.Replace(cwd, string(filepath.Separator), "-", -1)
}
