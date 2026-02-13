# dotsync

A lightweight CLI tool for syncing config files across machines using cloud storage providers.

## Overview

**dotsync** helps you keep your developer config files (dotfiles, AI tool configs, etc.) synchronized across multiple machines without the complexity of version control or specialized sync services. It works by moving your config files to cloud storage and creating symlinks at their original locations, letting your existing cloud provider (Google Drive, Dropbox, or iCloud) handle the synchronization.

## Features

- **Multi-provider support** - Works with Google Drive, Dropbox, and iCloud Drive
- **Auto-detection** - Automatically finds your cloud storage locations
- **Symlink management** - Seamlessly manages symlinks between local and cloud storage
- **Entry-based tracking** - Groups related files together (e.g., all OpenCode configs)
- **Cross-machine sync** - Set up once, sync everywhere
- **Safe operations** - Automatic backups before destructive operations
- **Status tracking** - View sync status of all tracked files

## Installation

### Prerequisites

- Go 1.21 or later
- One of the supported cloud storage providers (Google Drive, Dropbox, or iCloud Drive) installed and syncing

### Install from source

```bash
git clone https://github.com/wtfzambo/dotsync.git
cd dotsync
task build
```

The binary will be available at `./bin/dotsync`.

### Install via go install

```bash
go install github.com/wtfzambo/dotsync/cmd/dotsync@latest
```

### Verify installation

```bash
dotsync --help
```

## Quick Start

### 1. Initialize dotsync

Choose your cloud storage provider and initialize:

```bash
# For Google Drive
dotsync init gdrive

# For Dropbox
dotsync init dropbox

# For iCloud Drive (macOS only)
dotsync init icloud

# Or specify a custom path
dotsync init --path ~/my-cloud-folder
```

dotsync will auto-detect your cloud storage location and create a `dotsync` folder inside it to store your synced files.

### 2. Add files to sync

Add config files or directories you want to sync:

```bash
# Add a config file
dotsync add ~/.config/opencode/config.json

# Add a dotfile
dotsync add ~/.zshrc

# Add with a custom entry name
dotsync add ~/.aws/credentials --name aws-config
```

Each file is moved to cloud storage and replaced with a symlink. Related files are grouped into "entries" (e.g., all files under `~/.config/opencode/` become the "opencode" entry).

### 3. Check sync status

View all tracked files and their sync status:

```bash
# List all entries
dotsync list

# Show detailed file information
dotsync list --details
```

### 4. Set up on a new machine

On another machine with the same cloud storage account:

```bash
# Initialize with the same provider
dotsync init gdrive

# Create symlinks for all tracked files
dotsync link

# Or link a specific entry
dotsync link opencode
```

dotsync will create symlinks pointing to the cloud-synced files. If local files exist, you'll be prompted to back them up, skip, or abort.

## Supported Cloud Providers

### Google Drive

- **macOS:** `~/Library/CloudStorage/GoogleDrive-*/My Drive`
- **Linux:** `~/Google Drive` or `~/google-drive`
- Auto-detection: Yes

### Dropbox

- **macOS:** `~/Library/CloudStorage/Dropbox` or `~/Dropbox`
- **Linux:** `~/Dropbox`
- Auto-detection: Yes

### iCloud Drive

- **macOS:** `~/Library/Mobile Documents/com~apple~CloudDocs`
- **Linux:** Not available
- Auto-detection: Yes (macOS only)

If auto-detection fails, you can specify the path manually using `dotsync init --path <your-path>`.

## Commands Reference

| Command | Description | Examples |
|---------|-------------|----------|
| `init <provider>` | Initialize dotsync with a cloud storage provider | `dotsync init gdrive`<br>`dotsync init --path ~/my-cloud` |
| `add <path>` | Add a file to be synced | `dotsync add ~/.zshrc`<br>`dotsync add ~/.config/test/config.json` |
| `list` | List all tracked entries and their status | `dotsync list`<br>`dotsync list --details` |
| `link [entry]` | Create symlinks for tracked files | `dotsync link`<br>`dotsync link opencode`<br>`dotsync link --backup` |
| `unlink [entry]` | Remove symlinks and restore files locally | `dotsync unlink`<br>`dotsync unlink opencode` |

### Command Details

#### `dotsync init`

Initializes dotsync with a cloud storage provider.

**Flags:**
- `-p, --path <path>` - Explicitly specify the storage path (skips auto-detection)

**Example:**
```bash
dotsync init gdrive
```

#### `dotsync add`

Adds a file to be tracked and synced. The file is moved to cloud storage and replaced with a symlink.

**Flags:**
- `-n, --name <name>` - Specify a custom entry name (otherwise inferred from path)

**Example:**
```bash
dotsync add ~/.config/opencode/config.json
dotsync add ~/.aws/credentials --name aws-config
```

#### `dotsync list`

Lists all tracked entries and their sync status on this machine.

**Flags:**
- `-d, --details` - Show detailed file list for each entry

**Example:**
```bash
dotsync list --details
```

#### `dotsync link`

Creates symlinks for tracked files. Use this on a new machine to set up symlinks pointing to cloud-synced files.

**Flags:**
- `-b, --backup` - Automatically backup existing files without prompting

**Example:**
```bash
dotsync link               # Link all entries
dotsync link opencode      # Link only the "opencode" entry
dotsync link --backup      # Auto-backup conflicts
```

#### `dotsync unlink`

Removes symlinks and copies files from cloud storage back to their original locations. The files remain tracked and can be re-linked later.

**Example:**
```bash
dotsync unlink             # Unlink all entries
dotsync unlink opencode    # Unlink only the "opencode" entry
```

## How It Works

dotsync uses a simple approach to sync files across machines:

1. **File Movement:** When you add a file, dotsync moves it from your local filesystem to a `dotsync` folder in your cloud storage.

2. **Symlink Creation:** A symlink is created at the original file location, pointing to the cloud storage location.

3. **Cloud Sync:** Your cloud storage provider (Google Drive, Dropbox, iCloud) handles the actual synchronization across machines.

4. **Manifest Tracking:** A manifest file (`.dotsync.json`) in cloud storage tracks all files and their locations.

5. **Multi-Machine Setup:** On another machine, `dotsync link` reads the manifest and creates the appropriate symlinks.

### File Structure

```
<cloud-storage>/
└── dotsync/
    ├── .dotsync.json          # Manifest file
    ├── opencode/              # Entry name
    │   └── config/
    │       └── config.json    # Actual file
    └── zsh/                   # Another entry
        └── .zshrc
```

### Local Configuration

dotsync stores its local configuration at `~/.config/dotsync/config.json`. This file contains:
- The cloud storage path
- Local settings (if any)

## Important Notes

### macOS plist files

**macOS 14+ does NOT support symlinks for plist files** in `~/Library/Preferences/`. dotsync will reject these files with an error message. For syncing plist files, use [Mackup](https://github.com/lra/mackup) with [Copy mode](https://github.com/lra/mackup?tab=readme-ov-file#copy-mode) instead.

### Files outside home directory

dotsync will warn you if you try to add files outside your home directory. Symlinks may not work correctly if the absolute paths differ across machines.

### File must exist

You can only add files that currently exist on your filesystem. dotsync cannot add files that don't exist yet.

### Cloud storage must be available

All commands require your cloud storage to be mounted and accessible. If you see "storage unavailable" errors, check that your cloud storage is running and synced.

## Development

### Build the binary

```bash
task build        # Optimized build
task dev          # Fast build without optimizations
```

### Run tests

```bash
task test         # Run all tests
task test:cover   # Generate coverage report
```

### Other tasks

```bash
task clean        # Remove build artifacts
task tidy         # Tidy Go modules
task lint         # Run linters (requires golangci-lint)
task install      # Install to GOPATH/bin
```

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- Go standard library
