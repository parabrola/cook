package internal

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/parabrola/cook/internal/cli"
	"gopkg.in/yaml.v3"
)

type Parseable interface {
	Bootstrap() error
	GetGlobal() *Global
	GetTask(string) (Task, bool)
	GetFilePaths() []string
	parseTasks() error
	parseGlobal() error
	expandFilePaths(string) ([]string, error)
	getTempFileName() string
	shouldClearCache(string) bool
}

type Task struct {
	Name      string
	Files     []string          `yaml:"files,omitempty" json:"files,omitempty"`
	Run       []string          `yaml:"run" json:"run"`
	Env       map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	DependsOn []string          `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
}

type Global struct {
	Shared struct {
		EnvFile []string          `yaml:"env_file,omitempty" json:"env_file,omitempty"`
		Env     map[string]string `yaml:"environment,omitempty" json:"environment,omitempty"`
		Events  struct {
			BeforeEachRun  []string `yaml:"before_each_run,omitempty" json:"before_each_run,omitempty"`
			AfterEachRun   []string `yaml:"after_each_run,omitempty" json:"after_each_run,omitempty"`
			BeforeEachTask []string `yaml:"before_each_task,omitempty" json:"before_each_task,omitempty"`
			AfterEachTask  []string `yaml:"after_each_task,omitempty" json:"after_each_task,omitempty"`
		} `yaml:"events,omitempty" json:"events,omitempty"`
	} `yaml:"global,omitempty" json:"global,omitempty"`
}

type parser struct {
	Tasks     taskList
	FilePaths []string
	config    string
	isJSON    bool
	options   Options
	fs        FileSystem
	Global
}

// looksLikeJSON returns true if the config content appears to be JSON.
func looksLikeJSON(cfg string) bool {
	return len(strings.TrimSpace(cfg)) > 0 && strings.TrimSpace(cfg)[0] == '{'
}

// unmarshal dispatches to JSON or YAML based on the config file format.
func (p *parser) unmarshal(data []byte, v any) error {
	if p.isJSON {
		return json.Unmarshal(data, v)
	}
	return yaml.Unmarshal(data, v)
}

type taskList map[string]Task

var osCommandRegexp = regexp.MustCompile(`\$\((.+)\)`)
var parserString string

// NewParser creates a parser instance which can be either a blank one,
// or one provided  from the cache, which gets deserialized.
func NewParser(cfg string, opts *Options, fs FileSystem) Parseable {
	p := parser{}
	p.fs = fs
	p.config = cfg
	p.isJSON = looksLikeJSON(cfg)
	p.options = *opts

	tempFile := path.Join(p.fs.TempDir(), p.getTempFileName())

	if p.shouldClearCache(tempFile) {
		_ = p.fs.Remove(tempFile)
	}

	if !p.fs.FileExists(tempFile) {
		return &p
	}

	pBytes, err := p.fs.ReadFile(tempFile)
	if err != nil {
		_ = p.fs.Remove(tempFile)
		return &p
	}

	pStr := string(pBytes)
	deserialized, err := GOBDeserialize(pStr, &p)
	if err != nil {
		_ = p.fs.Remove(tempFile)
		return &p
	}

	p = deserialized
	parserString = pStr
	return &p
}

// Bootstrap does the parsing process or skip if cached.
func (p *parser) Bootstrap() error {
	if parserString != "" {
		p.restoreEnv()
		return nil
	}

	if err := p.parseGlobal(); err != nil {
		return err
	}

	if err := p.parseTasks(); err != nil {
		return err
	}

	pStr := GOBSerialize(p)
	_ = p.fs.WriteFile(path.Join(p.fs.TempDir(), p.getTempFileName()), []byte(pStr), 0644)

	return nil
}

func (p *parser) GetGlobal() *Global {
	return &p.Global
}

func (p *parser) GetTask(taskName string) (Task, bool) {
	task, ok := p.Tasks[taskName]
	return task, ok
}

func (p *parser) GetFilePaths() []string {
	return p.FilePaths
}

func (p *parser) restoreEnv() {
	for k, v := range p.Global.Shared.Env {
		os.Setenv(k, v)
	}
	for _, task := range p.Tasks {
		for k, v := range task.Env {
			os.Setenv(k, v)
		}
	}
}

// unmarshalTasks unmarshals the config into a taskList, handling JSON's
// $schema key which is a string (not a task object) and would cause a type error.
func (p *parser) unmarshalTasks(data []byte) (taskList, error) {
	if p.isJSON {
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, err
		}
		delete(raw, "$schema")
		delete(raw, "global")

		tasks := make(taskList)
		for k, v := range raw {
			var t Task
			if err := json.Unmarshal(v, &t); err != nil {
				return nil, err
			}
			tasks[k] = t
		}
		return tasks, nil
	}

	var tasks taskList
	if err := yaml.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

// Parses the individual user defined tasks in the config,
// and processes the dynamic parts of both "run" and "files" sections.
func (p *parser) parseTasks() error {
	tasks, err := p.unmarshalTasks([]byte(p.config))
	if err != nil {
		return err
	}

	allFilesPaths := []string{}

	for k, c := range tasks {
		filePaths := []string{}
		for i := range c.Files {
			cli.ReplaceEnvironmentVariables(osCommandRegexp, &tasks[k].Files[i])
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
			tasks[k].Run[i] = strings.ReplaceAll(r, "{FILES}", strings.Join(c.Files, " "))
			cli.ReplaceEnvironmentVariables(osCommandRegexp, &tasks[k].Run[i])
		}

		if len(c.Env) != 0 {
			vars, err := cli.SetEnvVariables(c.Env)
			if err != nil {
				return err
			}
			c.Env = vars
		}
		c.Name = k
		tasks[k] = c
	}

	delete(tasks, "global")

	if err := ValidateDependencies(tasks); err != nil {
		return err
	}

	p.FilePaths = allFilesPaths
	p.Tasks = tasks

	return nil
}

// Parses the "global" key in the yaml config and adds it to the parser.
// Also sets all variables under global.environment as OS environment variables.
func (p *parser) parseGlobal() error {
	var g Global

	if err := p.unmarshal([]byte(p.config), &g); err != nil {
		return err
	}

	if err := LoadDotenvFiles(g.Shared.EnvFile, p.fs); err != nil {
		return err
	}

	vars, err := cli.SetEnvVariables(g.Shared.Env)
	if err != nil {
		return err
	}

	g.Shared.Env = vars
	p.Global = g

	return nil
}

// Expand the path glob and returns all paths in an array
func (p *parser) expandFilePaths(file string) ([]string, error) {
	filePaths := []string{}

	if strings.Contains(file, "*") {
		files, err := p.fs.Glob(file)
		if err != nil {
			return nil, err
		}

		if len(files) > 0 {
			filePaths = append(filePaths, files...)
		}
	} else if p.fs.FileExists(file) {
		filePaths = append(filePaths, file)
	}

	return filePaths, nil
}

// Retrieves the temp file name
func (p *parser) getTempFileName() string {
	cwd, _ := p.fs.Getwd()
	return "cook-" + strings.ReplaceAll(cwd, string(filepath.Separator), "-")
}

// Determines whether the parser cache should be cleaned or not
func (p *parser) shouldClearCache(tempFile string) bool {
	tempFileExists := p.fs.FileExists(tempFile)
	mustCleanCache := false

	if !p.options.NoCache && tempFileExists {
		tempStat, _ := p.fs.Stat(tempFile)
		tempModTime := tempStat.ModTime().Unix()

		configStat, _ := p.fs.Stat(CurrentConfigFile())
		configModTime := configStat.ModTime().Unix()

		mustCleanCache = tempModTime < configModTime
	}

	if p.options.NoCache && tempFileExists {
		mustCleanCache = true
	}

	return mustCleanCache
}
