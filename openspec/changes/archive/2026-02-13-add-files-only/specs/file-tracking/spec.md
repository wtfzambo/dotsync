# Specification: file-tracking

## MODIFIED Requirements

### Requirement: System validates files before tracking

The system SHALL validate that files can be tracked.

#### Scenario: Reject symlinks

- **WHEN** user runs `dotsync add <path>` where path is a symlink
- **THEN** system displays error: "Cannot track symlinks. If this is already synced elsewhere, unlink it first."
- **THEN** system exits with non-zero status

#### Scenario: Reject non-existent files

- **WHEN** user runs `dotsync add <path>` where path does not exist
- **THEN** system displays error: "File not found: <path>"
- **THEN** system exits with non-zero status

#### Scenario: Reject directories

- **WHEN** user runs `dotsync add <path>` where path is a directory
- **THEN** system displays error: "Cannot add directories. Use 'dotsync add <file>' to add individual files."
- **THEN** system exits with non-zero status

#### Scenario: Warn for paths outside home directory

- **WHEN** user adds a file outside home directory (e.g., `/etc/hosts`)
- **THEN** system displays warning: "File is outside home directory. Symlinks may not work as expected if paths differ across machines."
- **THEN** system asks user for confirmation

#### Scenario: Warn for macOS plist files

- **WHEN** user adds a file matching `~/Library/Preferences/*.plist` on macOS
- **THEN** system displays warning: "macOS 14+ does NOT support symlinks for plist files. Use `Mackup` Copy mode (https://github.com/lra/mackup?tab=readme-ov-file#copy-mode) for these files."
- **THEN** system exits with zero-status (no-op).
