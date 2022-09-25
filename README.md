# Goke
Goke is a build automation tool, similar to Make, but without the Makefile clutter.

## What makes it different:

* Uses YAML to declare build configurations, instead of the Makefile syntax
* Built in Go, making it a blazing fast, multi-threaded tool
* Support for global hooks
* Intuitive environment variable declaration at any position in the configuration
* And more!

## Installation

#### Homebrew (Recommended)

```
brew tap dugajean/goke
brew install goke
```

#### GitHub releases

Download the appropriate executable for your system from the [releases page](https://github.com/dugajean/goke/releases).

## Example configuration (goke.yml)
```
global:
  environment:
    FOO: "foo"
    BAR: "$(echo 'BAR')"
    BAZ: "$(FOO)"
    LOKI: "Loki"

  events:
    before_each_run:
      - "echo 'This will run before each command in a given task'"
    after_each_run:
      - "echo 'This will run after each command in a given task'"
      - "greet-pepper"
    before_each_task:
      - "echo 'This will run once before the given task'"
    after_each_task:
      - "echo 'This will run once after the given task'"

greet-pepper:
  run:
    - "echo 'Hello Pepper'"

greet-loki:
  run:
    - "echo 'Hello ${LOKI}'"

greet-cats:
  files: [cmd/cli/*]
  run:
    - "echo 'Hello Frey'"
    - "echo 'Hello Sunny'"
    - "greet-loki"
```

## Running commands
From your project directory, you can now issue the following commands with the configuration shown above:
```
$ goke greet-cats
$ goke greet-loki
$ goke greet-pepper
```

#### `main` task

If you omit the task name and only run `goke`, it will look for a `main` task in the configuration file.

#### Available flags

| Flag | What it does |
|---|---|
| `--init` | Creates a simple `goke.yml` file in the current directory, if one doesn't already exist |
| `--version` | Prints the current version of goke |
| `--watch` | Runs the given command in _watch_ mode, meaning it will watch the files under `files:` and rerun the command whenever they change |
| `--force` | Runs the given command regardless whether the files under `files:` have changed |
| `--no-cache` | Goke caches the given configuration to speed up execution and avoid parsing the configuration on every run. Clear the cache if you are changing your configuration |

## Tests
Goke has some unit test coverage. PR’s are welcome to add more tests.

Run tests with:
```
go test ./internal
```

## Contributing
This project started a way for me to practice Go, but then I decided to turn it into a full fledged tool that can serve everyone.

I would really appreciate your contributions, either through PR’s, bug reporting, feature requests, etc.

For bug reports, please specify the exact steps on how to reproduce the problem.

You decided to contribute? Holy s$%&, thanks! 🚀 Please run this command from the root of your fork before you write any code:

```
git config --local core.hooksPath .githooks/
```

## License
GNU General Public License v3.0
