package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var lockfile_rw = MemOSWrapper{}

func TestTaskFilesExpansion1(t *testing.T) {
	parser := NewParser(yamlConfigStub, &options, &lockfile_rw)

	parser.parseTasks()

	greetCatsTask := parser.Tasks["greet-cats"]

	assert.NotNil(t, greetCatsTask)
}
