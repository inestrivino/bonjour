# bonjour

- /build
  - /ci
  - /package
- /cmd
  - main.go
  - main_test.go
- /docs
- /examples
- /internal
  - /config
    - config.go
    - config_test.go
  - modules
    - /events
    - /quotes
    - /weather
  - /ui
- /scripts
- .gitignore
- go.mod
- go.sum
- LICENSE
- README.md

If you run the script as `NO_COLOR=1 go run cmd/main.go` you will be able to display the result without color or style.
Run `go run scripts/testCoverage.go` to receive a quality html report on code coverage (which lines are covered and which are not, etc).
