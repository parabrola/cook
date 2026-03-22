# Cook
Cook is a build automation tool, similar to Make, but without the Makefile clutter.

## What makes it different:

* Uses YAML to declare build configurations, instead of the Makefile syntax
* Built in Go, making it a blazing fast, multi-threaded tool
* Task dependencies with automatic parallel execution
* File watching with OS-level filesystem events
* Support for global hooks
* Intuitive environment variable declaration at any position in the configuration
* Automatic `.env` file loading

## Installation

#### Homebrew (macOS/Linux)

```
brew tap parabrola/cook
brew install cook
```

#### Scoop (Windows)

```
scoop bucket add cook https://github.com/parabrola/scoop-cook
scoop install cook
```

#### deb (Debian/Ubuntu)

Download the `.deb` file from the [releases page](https://github.com/parabrola/cook/releases), then:

```
sudo dpkg -i cook_*.deb
```

#### rpm (Fedora/RHEL)

Download the `.rpm` file from the [releases page](https://github.com/parabrola/cook/releases), then:

```
sudo rpm -i cook_*.rpm
```

#### Go install

```
go install github.com/parabrola/cook/cmd/cli@latest
```

#### GitHub releases

Download the appropriate binary for your system from the [releases page](https://github.com/parabrola/cook/releases).

## Example configuration (cook.yml)
```yaml
global:
  environment:
    APP_NAME: "myapp"
    VERSION: "$(git describe --tags --always)"
    GOFLAGS: "-trimpath"

  events:
    before_each_task:
      - "echo '==> Starting task'"

clean:
  run:
    - "rm -rf ./build ./dist ./coverage.out"

generate:
  files: [internal/*.go]
  run:
    - "go generate ./..."

lint:
  depends_on: [generate]
  files: [cmd/**/*.go, internal/**/*.go]
  run:
    - "golangci-lint run ./..."

test:
  depends_on: [generate]
  files: [internal/*_test.go]
  run:
    - "go test -race -coverprofile=coverage.out ./..."

build:
  depends_on: [lint, test]
  files: [cmd/cli/*.go, internal/*.go]
  run:
    - "go build -ldflags '-s -w -X main.version=${VERSION}' -o ./build/${APP_NAME} ./cmd/cli"

docker:
  depends_on: [build]
  run:
    - "docker build -t ${APP_NAME}:${VERSION} ."

deploy-staging:
  depends_on: [docker]
  run:
    - "kubectl set image deployment/${APP_NAME} app=${APP_NAME}:${VERSION} -n staging"
  env:
    KUBECONFIG: "~/.kube/staging.yaml"

dev:
  files: [cmd/**/*.go, internal/**/*.go]
  run:
    - "go build -o ./build/${APP_NAME} ./cmd/cli"
    - "./build/${APP_NAME}"
```

## Running commands
From your project directory, you can now issue commands with the configuration shown above:
```
$ cook test
$ cook greet-cats
$ cook greet-loki
```

#### `main` task

If you omit the task name and only run `cook`, it will look for a `main` task in the configuration file.

## Task dependencies

Use `depends_on` to declare that a task requires other tasks to complete first:

```yaml
test:
  depends_on: [compile, lint]
  run:
    - "go test ./..."
```

Dependencies that don't depend on each other run in parallel automatically. In the example above, `compile` and `lint` both depend on `clean` but not on each other, so they run concurrently after `clean` finishes.

Each dependency runs at most once per invocation, even if multiple tasks depend on it (diamond dependency deduplication).

Circular dependencies are detected at parse time and will produce an error.

## File watching

Use the `-w` flag to watch for file changes and re-run the task automatically:

```
$ cook test -w
```

The `files` key specifies which files to monitor. Cook uses OS-level filesystem events (kqueue on macOS, inotify on Linux) for instant detection with zero CPU overhead when idle.

## Environment variables

Environment variables can be declared globally or per-task:

```yaml
global:
  environment:
    MY_VAR: "hello"
    DYNAMIC: "$(echo 'computed at parse time')"

my-task:
  run:
    - "echo ${MY_VAR}"
  env:
    LOCAL_VAR: "only available in this task"
```

Use `${VAR}` to reference environment variables in commands, and `$(command)` to capture command output.

### `.env` file support

Cook automatically loads `.env` files if they exist in your project directory:

1. `.env` — shared defaults
2. `.env.local` — personal overrides (should be gitignored)

Later files override earlier ones. You can also specify additional files explicitly:

```yaml
global:
  env_file:
    - .env.production
  environment:
    APP_URL: "${BASE_URL}/api"
```

Explicit files listed in `env_file` are loaded after the defaults, and must exist or Cook will return an error. Variables from `.env` files are available for use in `environment`, `run`, and `files` sections.

## Events

Global hooks that run before/after tasks and individual commands:

```yaml
global:
  events:
    before_each_task:
      - "echo 'starting task'"
    after_each_task:
      - "echo 'task complete'"
    before_each_run:
      - "echo 'before each command'"
    after_each_run:
      - "echo 'after each command'"
```

## Available flags

```
-h --help      Show help screen
-v --version   Show version
-i --init      Creates a cook.yaml file in the current directory
-t --tasks     Outputs a list of all task names
-w --watch     Run task in watch mode
-c --no-cache  Clears the program's cache
-f --force     Runs the task even if files have not been changed
-a --args=<a>  The arguments and options to pass to the underlying commands
-q --quiet     Suppresses all output from tasks
```

## Tests

Run tests with:
```
go test ./...
```

## Contributing
I would really appreciate your contributions, either through PR's, bug reporting, feature requests, etc.

For bug reports, please specify the exact steps on how to reproduce the problem.

You decided to contribute? Please run this command from the root of your fork before you write any code:

```
git config --local core.hooksPath .githooks/
```

## License
GNU General Public License v3.0
