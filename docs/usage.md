# Usage Guide

> Learn how to effectively use `bonjour`. From special commands to flags and environment variables.

## Basic Usage

Run the default dashboard by executing `bonjour` without any arguments:

```bash
bonjour
```

To view general help or options at any time, run:

```bash
bonjour --help
```

You can check your current `bonjour` version with:

```bash
bonjour --version
```

## Command Line Flags

Flags allow you to customize the layout and content shown when launching `bonjour`.

| Flag | Short | Description |
| --- | --- | --- |
| `--help` | `-h` | Display help information for `bonjour` |
| `--mini` |  | Execute and display a compact dashboard view |
| `--noevents` |  | Suppress/hide the events module |
| `--noquotes` |  | Suppress/hide the daily quotes module |
| `--noweather` |  | Suppress/hide the weather module |
| `--version` | `-v` | Display the current installed version of `bonjour` |

### Flag Examples

* **Display a compact view without weather:**

```bash
bonjour --mini --noweather
```

* **Disable both events and quotes:**

```bash
bonjour --noevents --noquotes
```

## Commands

### `config`

Opens the interactive configuration wizard to edit your settings.

```bash
bonjour config
```

### `events`

Opens the interactive events wizard to manage your daily schedule (add, edit, or delete events).

```bash
bonjour events
```

### `completion`

Generates auto-completion scripts for your preferred shell (e.g., `bash`, `zsh`, `fish`, `powershell`).

```bash
# Example: Generate completion for Zsh
bonjour completion zsh
```

### `help`

Provides detailed help for any specific command.

```bash
bonjour [command] --help
```

## Environment Variables

### NO_COLOR

`bonjour` honors the standard [NO_COLOR](https://no-color.org/) environment variable specification. When set to any non-empty string, ANSI color codes and text styling will be disabled across all output.

```bash
# Run once without color output
NO_COLOR=1 bonjour

# Disable color output for the current terminal session
export NO_COLOR=1
bonjour
```
