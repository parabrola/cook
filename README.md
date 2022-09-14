# Goke
Goke is a build automation tool, similar to Make, but without the Makefile clutter.

## What makes it different:

* Uses YAML to declare build configurations, instead of the Makefile syntax
* Built in Go, making it a blazing fast, multi-threaded tool
* Support for global hooks [Read more]
* Intuitive environment variable declaration at any position in the configuration
* And more!

## Installation
Download the appropriate executable for your system from the releases page.         

## Example configuration (goke.yml)
```
global:
  environment:
    FOO: "foo"
    BAR: "$(echo 'BAR')"
    BAZ: "$(FOO)"

  events:
    before_each_run:
      - "echo 'before each 1'"
    after_each_run:
      - "echo 'after each 1'"
      - "greet-pepper"
    before_each_task:
      - "echo 'before task'"
    after_each_task:
      - "echo 'after task'"

greet-pepper:
  run:
    - "echo 'Hello Pepper'"

greet-loki:
  run:
    - "echo 'Hello Loki'"

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

*Additional flags:*

* `--force`: Runs the given command regardless whether the files under `files:` have changed
* `--no-cache` : Goke caches the given configuration to speed up execution and avoid parsing the configuration on every run. Clear the cache if you are changing your configuration

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

## License
GNU General Public License v3.0
