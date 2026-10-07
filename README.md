# bonjour

<div align="center">

<img src="./assets/bonjour logo.png" alt="bonjour's logo" width="400"/>

<!-- Workflow Status Badge -->
[![Tests](https://github.com/inestrivino/bonjour/actions/workflows/testcoverage.yaml/badge.svg)](https://github.com/inestrivino/bonjour/actions/workflows/testcoverage.yaml) <!-- Coverage Badge --> ![Coverage](https://github.com/inestrivino/bonjour/raw/badges/assets/coveragebadge.svg) <!-- Tag badge --> [![Tag](https://img.shields.io/github/v/tag/inestrivino/bonjour?color=blue&label=version)](https://github.com/inestrivino/bonjour/releases/latest) <!-- License badge --> [![License](https://img.shields.io/badge/License-GPLv3-blue)](LICENSE)

</div>

> bonjour is a CLI dashboard application that helps you start your day the right way. Receive a greeting, a motivational quote, weather information for your area, and a list of upcoming events!

## Table of contents

- [bonjour](#bonjour)
  - [Table of contents](#table-of-contents)
  - [Demo and usage](#demo-and-usage)
  - [Installation](#installation)
    - [Go installation](#go-installation)
    - [Binaries installation](#binaries-installation)
      - [Linux and macOS](#linux-and-macos)
      - [Windows](#windows)
    - [Docker](#docker)
  - [Contributing](#contributing)
  - [Tools used](#tools-used)
  - [License and authorship](#license-and-authorship)
  - [Related documentation](#related-documentation)

## Demo and usage

<img src="./assets/demo.gif" alt="bonjour demo">

Make sure to check the [Usage Documentation](./docs/usage.md) for a detailed guide on all the flags, commands and personalization elements for introducing bonjour into your daily workflow.

## Installation

There are 3 different ways to get bonjour in your machine:

1. [Go installation](#go-installation)
2. [Binaries installation](#binaries-installation)
3. [Docker](#docker)

If you are having issues during installation, please refer to the [troubleshooting documentation](./docs/troubleshooting.md). If your problem is not described there, open an issue explaining the problem in detail (check the "reporting bugs" section in [Contributing](./CONTRIBUTING.MD)).

### Go installation

If you have the Go toolchain installed, you can install bonjour directly with:

```sh
go install github.com/inestrivino/bonjour/cmd/bonjour@latest
```

### Binaries installation

#### Linux and macOS

For Linux and macOS systems, a complementary [installation script](./scripts/install.sh) is included with the project.

1. Clone the project or get the source code from the [releases page](https://github.com/inestrivino/bonjour/releases/latest) and unzip it.
2. From the root of the project, execute the following commands (you may need to use `sudo`):

```bash
# Give the script execution privileges
chmod +x ./scripts/install.sh
# Execute the script
./scripts/install.sh
```

3. Now confirm that installation was successful:

```bash
bonjour --version
```

Alternatively, you can execute the script without downloading it.

1. In a terminal, execute the following commands:

```bash
curl -sSL [https://raw.githubusercontent.com/](https://raw.githubusercontent.com/)inestrivino/bonjour/main/scripts/install.sh | bash
```

2. Now confirm that installation was successful:

```bash
bonjour --version
```

There is also a complementary [uninstallation script](./scripts/uninstall.sh), that you can run just like the installation script (exchange `install.sh` for `uninstall.sh` in the commands).

If you wish to manually install the binaries yourself, download the binary corresponding to your OS and architecture from the [releases page](https://github.com/inestrivino/bonjour/releases/latest), then follow the manual installation steps for your OS.

#### Windows

For Windows systems, the binary archives must be manually installed.

1. Download the latest Windows archive from the [releases page](https://github.com/inestrivino/bonjour/releases/latest).
2. Extract `bonjour.exe` into a folder of your choice (for example: `C:\Program Files\bonjour\`).
3. Add that folder path to your System Environment Variables (`Path`) so you can run `bonjour` from any terminal:
   - Open **Start** → Search for **"Environment Variables"** → Click **Edit the system environment variables**.
   - Click **Environment Variables...** at the bottom right.
   - Under **System variables**, select `Path` and click **Edit...**.
   - Click **New** and paste your folder path (`C:\Program Files\bonjour\`).
   - Click **OK** to save changes.
4. Open a new PowerShell or Command Prompt window and confirm that installation was successful:

```powershell
bonjour --version
```

### Docker

If you prefer not to install anything, there is a Docker image available with each new release.
For information on how to try it out, run it persistently, and more, check the [Docker documentation](./docs/docker.md).

## Contributing

If you want to contribute to the project, please read [Contributing](./CONTRIBUTING.MD).
If your contribution is related to software development, please also read the [Development documentation](./docs/development.md).

## Tools used

This project makes use of the [Open-Meteo API](https://open-meteo.com/) for weather and geocoding services, and the [Zen Quotes API](https://zenquotes.io/) for quotes.

Automatic binaries and docker images releases are handled with [Goreleaser v2](https://goreleaser.com/).

The UI and forms styling was achieved with [huh](github.com/charmbracelet/huh) and [lipgloss](github.com/charmbracelet/lipgloss).

The functionality to accept multiple commands and flags was implemented using [Cobra](github.com/spf13/cobra).

## License and authorship

This project is open-source software licensed under the [GNU General Public License v3.0 (GPLv3)](./LICENSE).

> [!IMPORTANT]
> **Important Note on Commercial Use**
> 
> While the source code of this application is freely available under the GPLv3, **this software cannot be legally embedded into commercial products, services, or internal business workflows** as is. This restriction is due to its integration with the **Open-Meteo API**, which restricts its free, key-less tier strictly to **non-commercial use**. Any commercial deployment requires a paid commercial license directly from Open-Meteo, or changing the source code to use a different API. More info on [Open-Meteo's About Page](https://open-meteo.com/en/about).

## Related documentation

* [Configuration Documentation](./docs/configuration.md): Learn where and how bonjour stores the user's configuration.
* [Troubleshooting Documentation](./docs/troubleshooting.md): See common troubleshooting fixes for installing and running bonjour.
* [Usage](./docs/usage.md): Learn all commands and flags you can use in bonjour and how.
