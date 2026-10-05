# Troubleshooting

> Common problems during installation or usage of the app, along with possible solutions.

## Installation problems

### Permission Denied

If you receive a message similar to this:

```bash
Permission Denied (bash: /usr/local/bin/bonjour: Permission denied)
```

Then ensure bonjour's binary has execution permissions in your computer:

```bash
sudo chmod +x /usr/local/bin/bonjour
```

### Command Not Found

1. If, while **installing from binaries**, you are receiving an error that says:

    ```bash
    command not found: bonjour
    ```

    It is probably happening due to `/usr/local/bin` not being in your shell's `$PATH`. You can add it via `~/.bashrc` or `~/.zshrc`:

    ```bash
    export PATH="/usr/local/bin:$PATH"
    ```

2. If you receive this error after having performed the **Go installation**, then make sure that `$GOPATH/bin` is in the system `$PATH`. Configure your `$PATH` to fix it:

    macOS:

    ```bash
    echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
    source ~/.zshrc
    ```

    Linux:

    ```bash
    echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.bashrc
    source ~/.bashrc
    ```

    Windows (PowerShell):

    ```PowerShell
    [Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:USERPROFILE\go\bin", "User")
    ```

## Execution problems

### macOS Gatekeeper Warning

If you are a macOS user, you may receive the following error/warning:

```bash
bonjour cannot be opened because it is from an unidentified developer
```

This happens because pre-compiled binaries from GitHub Releases aren't signed with an Apple Developer certificate.
macOS users can allow execution in one of two ways:

1. Run once in terminal to remove the quarantine attribute:
  
    ```bash
    xattr -d com.apple.quarantine /usr/local/bin/bonjour
    ```

2. Or go to `System Settings → Privacy & Security` and click `"Allow Anyway"`.
