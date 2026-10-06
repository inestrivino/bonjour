# Contributing Guide

> Thank you for taking the time to contribute to `bonjour`! Contributions like yours can make this project great.

## Ground rules

Please keep interactions respectful, constructive, and inclusive across issues, discussions, and pull requests.

Responsibilities of contributors:

* Ensure cross-platform compatibility for every change that's accepted. Windows, Mac & Linux.
* Create issues for any major changes and enhancements that you wish to make. Discuss things transparently and get community feedback.
* Keep feature versions as small as possible, preferably one new feature per version.

## How can I contribute?

1. Reporting bugs: Before creating a bug report, please check existing issues to avoid duplicates. When opening an issue include:
    * Clear title that summarizes the unexpected behaviour.
    * Explain the steps to follow to reproduce the issue.
    * Indicate your OS, Go version and `bonjour` version.
    * What you expected to happened vs what actually happened.
2. Suggesting features: If you have an idea for a new feature or a major refactor. Follow these steps:
    * Open a **Feature Request** issue first to discuss the scope before writing code.
    * Once approved by maintainers, start your development branch. More information in the section below.
3. Review pull requests, add documentation, or just share the project with your friends!

## Development & PR Workflow

For local setup, building, and testing commands, please refer to our **[Development Guide](docs/development.md)**.

### Pull Request Checklist

Before submitting your Pull Request, ensure:

- [ ] **Tests & Coverage:** All unit tests pass, and code coverage meets or exceeds our **90% threshold**.
- [ ] **Formatting:** Code passes standard formatting checks.
- [ ] **Documentation:** Relevant changes are updated in `/docs` (e.g., `docs/configuration.md`, `docs/usage.md`).
- [ ] **Commit Messages:** Written clearly in the imperative mood (e.g., `feat: add custom theme flag` or `fix: handle missing config path`).

## Coding Standards

* **Follow Idiomatic Go:** Follow official [Effective Go](https://golang.org/doc/effective_go) practices.
* **Errors:** Handle all returned errors explicitly; do not swallow unexpected errors.
* **Tests:** Write unit tests alongside every feature or bug fix.

## Code review process

As this is a small project, code from pull requests has to be reviewed by the repository's owner. You can expect to hear from me in, at maximum, one month's time.
After feedback has been given, a response is expected within a month. After that time the pull request may be closed due to not showing activity.
