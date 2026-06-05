# Installation

Marvin is distributed as a single CLI binary. It does not require AWS
credentials or a running service.

## Install From A GitHub Release

For tagged releases, download the binary for your operating system and CPU from
the GitHub Releases page:

https://github.com/kaleab-kali/marvin/releases

Release asset names use this pattern:

```text
marvin-linux-amd64
marvin-linux-arm64
marvin-darwin-amd64
marvin-darwin-arm64
marvin-windows-amd64.exe
marvin-windows-arm64.exe
```

Each binary is published with a matching `.sha256` checksum file.

### Verify Checksums On Linux Or macOS

From the directory containing the downloaded binary and checksum:

```sh
sha256sum -c marvin-linux-amd64.sha256
```

Replace the file name with the checksum for your platform.

### Verify Checksums On Windows

In PowerShell:

```powershell
Get-FileHash .\marvin-windows-amd64.exe -Algorithm SHA256
Get-Content .\marvin-windows-amd64.exe.sha256
```

The hash from `Get-FileHash` should match the hash in the `.sha256` file.

### Put Marvin On Your PATH

Move the binary to a directory on your `PATH`.

Linux or macOS example:

```sh
chmod +x marvin-linux-amd64
mkdir -p ~/.local/bin
mv marvin-linux-amd64 ~/.local/bin/marvin
marvin version
```

Windows PowerShell example:

```powershell
New-Item -ItemType Directory -Force "$HOME\bin"
Move-Item .\marvin-windows-amd64.exe "$HOME\bin\marvin.exe"
marvin version
```

Make sure the target directory is listed in your `PATH`.

## Build From Source

Requirements:

- Go 1.22 or newer.
- Git.

Clone and build:

```sh
git clone https://github.com/kaleab-kali/marvin.git
cd marvin
go test ./...
go build -o bin/marvin ./cmd/marvin
```

On Windows, the build output can use `.exe`:

```powershell
go build -o bin\marvin.exe .\cmd\marvin
```

Source builds are useful for contributors. Tagged release binaries are the
recommended path for users who want version-injected builds.

## Package Managers

Homebrew, Scoop, Winget, and container images are not published yet. Until those
are available, use GitHub Releases or build from source.
