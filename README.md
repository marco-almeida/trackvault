# TrackVault

TrackVault is a CLI tool to back up and restore Spotify library data.

It supports backing up liked songs and playlists (including private ones) and restoring them into a new account using portable backup files.

This tool requires your own Spotify Developer App\
In dev mode, only whitelisted accounts work\
This is a Spotify limitation, not a TrackVault bug

## Features

- Spotify OAuth authentication
- Backup liked songs and playlists (private included)
- Restore backups into a new Spotify account
- Platform-agnostic backup format
- Cross-platform (Windows, macOS, Linux)

## Installation

### Download a binary

Download a prebuilt binary from  
<https://github.com/marco-almeida/trackvault/releases>

On macOS/Linux:

```sh
chmod +x trackvault
```

### Install with Go

```sh
go install github.com/marco-almeida/trackvault@latest
```

## Usage

### Authenticate

```sh
trackvault login --provider spotify
```

### Backup

```sh
trackvault backup --provider spotify -o .
```

### Restore

```sh
trackvault restore --source ./backups/2025-03-01T14-32-09Z
```

If output directory isnt specified, the backups are stored by default in the OS data directory.

### Development

```sh
go run main.go
```

### Disclaimer

Not affiliated with Spotify. Use at your own risk.
