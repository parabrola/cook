package internal

import (
	"os"
	"testing"

	"github.com/parabrola/goke/internal/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func dotenvMockNoDefaults(t *testing.T) *tests.FileSystem {
	fsMock := tests.NewFileSystem(t)
	fsMock.On("FileExists", ".env").Return(false)
	fsMock.On("FileExists", ".env.local").Return(false)
	return fsMock
}

func TestLoadDotenvFilesSkipsMissingDefaults(t *testing.T) {
	fsMock := dotenvMockNoDefaults(t)

	err := LoadDotenvFiles(nil, fsMock)
	require.NoError(t, err)
}

func TestLoadDotenvFilesExplicitFileMustExist(t *testing.T) {
	fsMock := dotenvMockNoDefaults(t)
	fsMock.On("FileExists", ".env.production").Return(false)

	err := LoadDotenvFiles([]string{".env.production"}, fsMock)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), ".env.production")
	assert.Contains(t, err.Error(), "not found")
}

func TestLoadDotenvFilesLoadsExplicitFiles(t *testing.T) {
	os.Unsetenv("DOTENV_TEST_A")
	os.Unsetenv("DOTENV_TEST_B")

	dir := t.TempDir()
	envPath := dir + "/base.env"
	localPath := dir + "/local.env"

	os.WriteFile(envPath, []byte("DOTENV_TEST_A=base\nDOTENV_TEST_B=original\n"), 0644)
	os.WriteFile(localPath, []byte("DOTENV_TEST_B=overridden\n"), 0644)

	fsMock := dotenvMockNoDefaults(t)
	fsMock.On("FileExists", envPath).Return(true)
	fsMock.On("FileExists", localPath).Return(true)

	err := LoadDotenvFiles([]string{envPath, localPath}, fsMock)
	require.NoError(t, err)

	assert.Equal(t, "base", os.Getenv("DOTENV_TEST_A"))
	assert.Equal(t, "overridden", os.Getenv("DOTENV_TEST_B"))
}

func TestLoadDotenvFilesLaterOverridesEarlier(t *testing.T) {
	os.Unsetenv("DOTENV_OVERRIDE")

	dir := t.TempDir()
	first := dir + "/first.env"
	second := dir + "/second.env"

	os.WriteFile(first, []byte("DOTENV_OVERRIDE=first\n"), 0644)
	os.WriteFile(second, []byte("DOTENV_OVERRIDE=second\n"), 0644)

	fsMock := dotenvMockNoDefaults(t)
	fsMock.On("FileExists", first).Return(true)
	fsMock.On("FileExists", second).Return(true)

	err := LoadDotenvFiles([]string{first, second}, fsMock)
	require.NoError(t, err)

	assert.Equal(t, "second", os.Getenv("DOTENV_OVERRIDE"))
}

func TestParseGlobalWithEnvFile(t *testing.T) {
	os.Unsetenv("FROM_DOTENV")

	dir := t.TempDir()
	envPath := dir + "/.env.custom"
	os.WriteFile(envPath, []byte("FROM_DOTENV=it_works\n"), 0644)

	config := `
global:
  env_file:
    - ` + envPath + `
  environment:
    COMBINED: "${FROM_DOTENV}_and_more"
`
	fsMock := mockCacheDoesNotExist(t)
	fsMock.On("FileExists", ".env").Return(false)
	fsMock.On("FileExists", ".env.local").Return(false)
	fsMock.On("FileExists", envPath).Return(true)

	parser := NewParser(config, &clearCacheOpts, fsMock)

	err := parser.parseGlobal()
	require.NoError(t, err)

	assert.Equal(t, "it_works", os.Getenv("FROM_DOTENV"))
}
