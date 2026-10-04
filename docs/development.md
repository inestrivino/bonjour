# Development Guide

> Everything you need to set up your local environment, build, test, and contribute to `bonjour`.

## Prerequisites

Before building `bonjour` locally, make sure you have the following installed:

* **Go:** Version `1.22+`
* **Git:** For source control
* **Goreleaser (Optional):** For building release binaries and docker images of the project locally
* **Docker (Optional):** For testing containerized builds

Make sure to read `CONTRIBUTING.md` before doing anything else.

## Getting Started

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
  go run ./cmd/main.go
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

8. **Submit your work**: Commit your changes, push your branch, and open a Pull Request against the main branch.

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
  