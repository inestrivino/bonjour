# Development Guide

> Everything you need to set up your local environment, build, test, and contribute to `bonjour`.

## Prerequisites

Before building `bonjour` locally, make sure you have the following installed:

* **Go:** Version `1.22+`
* **Git:** For source control
* **Goreleaser v2 (Optional):** For building release binaries and docker images of the project locally
* **Docker (Optional):** For testing containerized builds

Make sure to read `CONTRIBUTING.md` before doing anything else.

## Project structure

Here is an overview of the project with explanations:

```txt
├── .github/                  # GitHub configuration and automation
│   └── workflows/
│       ├── coverage.yaml     # Test passing and code coverage tracking workflow
│       └── release.yaml      # Automated release pipeline (GoReleaser)
├── cmd/                      # Application entry points
│   ├── main.go               # Main executable logic
│   └── main_test.go          # Tests for entry point initialization
├── docs/                     # Project documentation 
├── examples/                 # Example files
├── internal/                 # Private application code
│   ├── config/               # Application configuration parsing and handling
│   │   ├── config.go
│   │   └── config_test.go
│   ├── modules/              # Core feature modules
│   │   ├── events/           # Events domain logic
│   │   ├── quotes/           # Quotes domain logic
│   │   └── weather/          # Weather domain logic
│   └── ui/                   # Terminal UI logic
├── scripts/                  # Helper scripts for installation, and testing
├── .gitignore                # Files and patterns ignored by Git
├── .goreleaser.yaml          # GoReleaser release configuration
├── CONTRIBUTING.md           # Guidelines for contributing to the project
├── Dockerfile                # Docker build instructions
├── go.mod                    # Go module definitions and core dependencies
├── go.sum                    # Go module dependency checksums
├── LICENSE                   # License terms
└── README.md                 
```

## Development workflow

1. **Clone the repository:**

  ```bash
  git clone https://github.com/inestrivino/bonjour.git
  cd bonjour
  ```

2. **Download the dependencies:**

  ```bash
  go mod download
  ```

3. **Create a local branch for development:**
  ```bash
  git checkout -b branch-name
  ```

4. **Run it with go:**

  ```bash
  go run ./cmd/bonjour/main.go
  ```

5. If all works well, **start editing the code**: Create new features or fix bugs. Don't forget to create the corresponding tests for every feature you create.

6. Ensure the code is correctly formatted to follow Go standards:

```bash
go fmt ./...
```

7. Ensure all tests pass and the coverage is over the treshold (>90%):

```bash
go run ./scripts/testcoverage.go
```

The result will be displayed in the terminal's output. This command also creates a complementary `coverage.html` at the root of the project, open it to see which lines are still uncovered.

8. **Submit your work**: Commit your changes, push your branch, and open a Pull Request against the main branch. Wait for a code review.

## Branch naming rules

All branches dedicated to new features should be called `[feature]-dev`. All branches related to bug fixes should be called `[bug]-fix`.

## Common development commands

* Building binary locally:

```bash
go build -o bin/bonjour ./cmd/main.go
./bin/bonjour
```

* (**GoReleaser necessary**) Creating the installation binaries for your OS and architecture:

```bash
goreleaser build --snapshot --single-target --clean -p 1
```

* (**GoReleaser and Docker necessary**) Create a local Docker image from the installation binaries built in the previous point:

```bash
docker build --build-arg TARGETPLATFORM=$(find dist -type d -name "bonjour_*" | head -n 1) -t bonjour:local .
docker run --rm -it bonjour:local
```

* (**GoReleaser necessary**) Snapshot release testing: 

```bash
goreleaser release --snapshot --clean -p 1
```
  