# Specification: symlink-management

## ADDED Requirements

### Requirement: System creates local backup before moving files

The system SHALL create a temporary local backup before moving files to cloud storage.

#### Scenario: Create backup before move

- **WHEN** system moves a file to cloud storage during `add`
- **THEN** system first copies the file to `~/.cache/dotsync/backups/<timestamp>-<filename>`
- **THEN** system moves the original file to cloud storage
- **THEN** system creates symlink at original location
- **THEN** system deletes the backup after successful operation

#### Scenario: Restore from backup on failure

- **WHEN** move or symlink creation fails
- **THEN** system restores file from `~/.cache/dotsync/backups/` to original location
- **THEN** system deletes the backup
- **THEN** system displays error explaining what went wrong

#### Scenario: Handle read-only parent directory

- **WHEN** system attempts to move a file from a read-only parent directory
- **THEN** system detects directory is not writable before attempting move
- **THEN** system displays error: "Cannot delete file from read-only directory: <path>"
- **THEN** system exits with non-zero status
- **THEN** file remains in original location only (no partial state)

#### Scenario: Backup directory creation

- **WHEN** `~/.cache/dotsync/backups/` does not exist
- **THEN** system creates the directory before backup

### Requirement: System creates symlinks when adding files

The system SHALL create symlinks at original locations pointing to cloud storage after moving files.

#### Scenario: Create symlink after move

- **WHEN** system moves a file to cloud storage during `add`
- **THEN** system creates a symlink at the original path
- **THEN** symlink points to the file in cloud storage
- **THEN** file access through original path works transparently

#### Scenario: Create parent directories for symlink

- **WHEN** symlink target path has parent directories that don't exist
- **THEN** system creates necessary parent directories
- **THEN** system creates symlink in the created directory structure

### Requirement: User can link entries on new machines

The system SHALL create symlinks for tracked entries from the manifest.

#### Scenario: Link all files in an entry

- **WHEN** user runs `dotsync link opencode`
- **THEN** system reads "opencode" entry from manifest
- **THEN** for each file in entry, system creates symlink at `<root>/<file>` pointing to `<storage>/dotsync/opencode/<file>`

#### Scenario: Link all entries

- **WHEN** user runs `dotsync link` (no entry specified)
- **THEN** system links all entries in the manifest

#### Scenario: Entry not found

- **WHEN** user runs `dotsync link nonexistent`
- **THEN** system displays error: "Entry 'nonexistent' not found in manifest"
- **THEN** system exits with non-zero status

### Requirement: System handles existing files during link

The system SHALL prompt user when a file already exists at the symlink target location.

#### Scenario: Prompt for existing file

- **WHEN** system attempts to create symlink at a path that already contains a file
- **THEN** system prompts:

  ```
  File already exists: <path>
    [b] Backup to <path>.dotsync-backup
    [s] Skip this file
    [a] Abort entire operation
  Choice:
  ```

#### Scenario: User chooses backup

- **WHEN** user selects backup option
- **THEN** system renames existing file to `<path>.dotsync-backup`
- **THEN** system creates symlink at original path

#### Scenario: User chooses skip

- **WHEN** user selects skip option
- **THEN** system leaves existing file in place
- **THEN** system continues with next file
- **THEN** system reports skipped file at end of operation

#### Scenario: User chooses abort

- **WHEN** user selects abort option
- **THEN** system stops immediately
- **THEN** system does not undo any symlinks already created
- **THEN** system displays: "Aborted. Some files may have been linked."

#### Scenario: Force flag skips prompts

- **WHEN** user runs `dotsync link --backup`
- **THEN** system automatically backs up all conflicting files without prompting

### Requirement: User can unlink entries

The system SHALL restore files from cloud storage to original locations, removing symlinks.

#### Scenario: Unlink specific entry

- **WHEN** user runs `dotsync unlink opencode`
- **THEN** for each file in "opencode" entry:
- **THEN** system copies file from cloud storage to original location (replacing symlink)
- **THEN** original location now contains the actual file, not a symlink

#### Scenario: Unlink all entries

- **WHEN** user runs `dotsync unlink` (no entry specified)
- **THEN** system unlinks all entries in the manifest

#### Scenario: Unlink preserves cloud copy

- **WHEN** system unlinks a file
- **THEN** file remains in cloud storage
- **THEN** manifest entry remains unchanged
- **THEN** user can re-link later with `dotsync link`

#### Scenario: Unlink with broken symlink

- **WHEN** symlink exists but cloud file is missing
- **THEN** system removes the broken symlink
- **THEN** system displays warning: "Cloud file missing for <path>. Symlink removed but file could not be restored."

### Requirement: System validates symlink integrity

The system SHALL detect broken or incorrect symlinks.

#### Scenario: Detect broken symlink

- **WHEN** symlink exists but target file is missing (cloud storage offline or file deleted)
- **THEN** system reports symlink as "broken" in list output

#### Scenario: Detect file instead of symlink

- **WHEN** original location contains a regular file instead of symlink
- **THEN** system reports as "not linked" in list output
- **THEN** this indicates the file was never linked or symlink was replaced

#### Scenario: Detect symlink to wrong target

- **WHEN** symlink exists but points to different location than expected
- **THEN** system reports as "incorrect" in list output
- **THEN** system shows expected vs actual target

### Requirement: Symlinks are created atomically

The system SHALL ensure symlink creation is atomic to prevent partial states.

#### Scenario: Move and link atomically

- **WHEN** system adds a file (move + symlink)
- **THEN** system first moves file to cloud storage
- **THEN** system verifies file exists at destination
- **THEN** only then system creates symlink
- **THEN** if symlink creation fails, system moves file back and reports error

#### Scenario: Recover from failed symlink

- **WHEN** symlink creation fails after file was moved
- **THEN** system moves file back to original location
- **THEN** system removes entry from manifest
- **THEN** system displays error: "Failed to create symlink. File restored to original location."
