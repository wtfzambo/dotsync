# Proposal: dotsync

## Why

Developers using multiple machines need to sync configuration files (AI assistants, editors, shell configs) without manual git commit/push/pull cycles. Current solutions are either too manual or require complex daemon setups. Cloud storage providers already solve the hard sync problem—we just need symlink management.

## What Changes

- Add `dotsync init` command to configure cloud storage path and create manifest file
- Add `dotsync add <file>` command to move a file to cloud storage and create symlink
- Add `dotsync link` command to restore symlinks on a new machine from manifest
- Add `dotsync list` command to show tracked files and their link status
- Add `dotsync unlink <file>` command to restore a file from cloud storage to its original location

## Capabilities

### New Capabilities

- `storage-detection`: Detect and validate cloud storage paths (Google Drive, Dropbox, iCloud)
- `file-tracking`: Manage manifest of tracked files with source/destination paths
- `symlink-management`: Create, verify, and restore symlinks between original locations and cloud storage

### Modified Capabilities

(none - new project)

## Impact

- **CLI**: New `dotsync` command with subcommands (init, add, link, list, unlink)
- **File System**: Creates symlinks in user's home directory pointing to cloud storage
- **Dependencies**: Standard library only for MVP (fs, path operations)
- **Storage**: Creates `~/<cloud-folder>/dotsync/` directory with manifest.json and tracked files
