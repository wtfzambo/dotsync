# Specification: storage-detection

## ADDED Requirements

### Requirement: User can initialize storage with provider selection

The system SHALL accept a cloud provider argument during `init` and locate the storage path automatically.

#### Scenario: Initialize with Google Drive

- **WHEN** user runs `dotsync init gdrive`
- **THEN** system searches for Google Drive mount in known locations
- **THEN** if found, system creates `<gdrive-path>/dotsync/` directory and `.dotsync.json` manifest
- **THEN** system saves storage path to local config at `~/.config/dotsync/config.json`

#### Scenario: Initialize with Dropbox

- **WHEN** user runs `dotsync init dropbox`
- **THEN** system searches for Dropbox mount in known locations
- **THEN** if found, system creates `<dropbox-path>/dotsync/` directory and `.dotsync.json` manifest

#### Scenario: Initialize with iCloud

- **WHEN** user runs `dotsync init icloud`
- **THEN** system searches for iCloud Drive mount in known locations
- **THEN** if found, system creates `<icloud-path>/dotsync/` directory and `.dotsync.json` manifest

### Requirement: System searches known paths for cloud storage

The system SHALL search platform-specific known paths when locating cloud storage mounts.

#### Scenario: Google Drive detection on macOS

- **WHEN** system searches for Google Drive on macOS
- **THEN** system checks `~/Library/CloudStorage/GoogleDrive-*/My Drive/` (glob pattern)
- **THEN** system checks `~/Google Drive/` as fallback

#### Scenario: Dropbox detection on macOS

- **WHEN** system searches for Dropbox on macOS
- **THEN** system checks `~/Library/CloudStorage/Dropbox/`
- **THEN** system checks `~/Dropbox/` as fallback

#### Scenario: iCloud detection on macOS

- **WHEN** system searches for iCloud on macOS
- **THEN** system checks `~/Library/Mobile Documents/com~apple~CloudDocs/`

#### Scenario: Cloud storage detection on Linux

- **WHEN** system searches for cloud storage on Linux
- **THEN** system checks `~/Dropbox/` for Dropbox
- **THEN** system checks `~/.dropbox-dist/` as fallback for Dropbox
- **THEN** system checks `~/Google Drive/` for Google Drive (Insync, google-drive-ocamlfuse)
- **THEN** system checks `~/google-drive/` as fallback for Google Drive

#### Scenario: Cloud storage detection on Windows (future - PATHS TBD!!!)

- **WHEN** system searches for cloud storage on Windows
- **THEN** system checks `%USERPROFILE%\Google Drive\` for Google Drive
- **THEN** system checks `%USERPROFILE%\Dropbox\` for Dropbox
- **THEN** system checks `%USERPROFILE%\iCloudDrive\` for iCloud
- **THEN** system checks `%USERPROFILE%\OneDrive\` for OneDrive

### Requirement: User can specify explicit storage path

The system SHALL allow users to provide an explicit path when auto-detection fails or is not desired.

#### Scenario: Prompt for path when not found

- **WHEN** user runs `dotsync init gdrive`
- **THEN** system cannot find Google Drive in known locations
- **THEN** system prompts: "Google Drive not found at known locations. Enter path (or 'q' to quit):"
- **THEN** user enters a valid path
- **THEN** system uses that path for storage

#### Scenario: Initialize with explicit path flag

- **WHEN** user runs `dotsync init --path ~/my-cloud-folder`
- **THEN** system uses the provided path directly without provider detection
- **THEN** system creates `~/my-cloud-folder/dotsync/` directory and manifest

#### Scenario: Reject invalid path

- **WHEN** user provides a path that does not exist
- **THEN** system displays error: "Path does not exist: <path>"
- **THEN** system exits with non-zero status

### Requirement: System validates storage path is writable

The system SHALL verify write permissions before completing initialization.

#### Scenario: Storage path is not writable

- **WHEN** user initializes with a path that exists but is not writable
- **THEN** system displays error: "Cannot write to storage path: <path>"
- **THEN** system exits with non-zero status

#### Scenario: Storage path becomes writable after mount

- **WHEN** user initializes and path is writable
- **THEN** system creates the dotsync directory and manifest successfully

### Requirement: System stores local configuration

The system SHALL persist the configured storage path in a local (non-synced) config file.

#### Scenario: Save storage path to local config

- **WHEN** initialization completes successfully
- **THEN** system creates `~/.config/dotsync/config.json` with the storage path
- **THEN** subsequent commands read storage path from this config

#### Scenario: Config file already exists

- **WHEN** user runs `init` and `~/.config/dotsync/config.json` already exists
- **THEN** system prompts: "dotsync is already initialized. Reinitialize? [y/N]"
- **THEN** if user confirms, system overwrites the config

### Requirement: System validates storage availability before operations

The system SHALL check that configured storage is accessible before any operation that requires it.

#### Scenario: Storage unavailable during add

- **WHEN** user runs `dotsync add <file>`
- **THEN** configured storage path does not exist (cloud folder not mounted)
- **THEN** system displays error: "Cloud storage not available at <path>. Is it mounted?"
- **THEN** system exits with non-zero status

#### Scenario: No local config exists

- **WHEN** user runs any command except `init`
- **THEN** local config `~/.config/dotsync/config.json` does not exist
- **THEN** system displays error: "dotsync not initialized. Run 'dotsync init' first."
