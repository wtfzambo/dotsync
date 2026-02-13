# Specification: file-tracking

## ADDED Requirements

### Requirement: User can add files to tracking

The system SHALL allow users to add files to the manifest and move them to cloud storage.

#### Scenario: Add file with inferred name

- **WHEN** user runs `dotsync add ~/.config/opencode/config.json`
- **THEN** system infers name as "opencode" from path
- **THEN** system moves file to `<storage>/dotsync/opencode/config.json`
- **THEN** system creates symlink at original location pointing to cloud storage
- **THEN** system updates manifest with entry for "opencode"

#### Scenario: Add file with explicit name

- **WHEN** user runs `dotsync add ~/.zshrc --name shell`
- **THEN** system uses "shell" as the entry name
- **THEN** system moves file to `<storage>/dotsync/shell/.zshrc`
- **THEN** system updates manifest with entry for "shell"

#### Scenario: Add multiple files to same entry

- **WHEN** entry "opencode" already exists with root `~/.config/opencode`
- **THEN** user runs `dotsync add ~/.config/opencode/agents/review.md`
- **THEN** system adds `agents/review.md` to the existing "opencode" entry's files list
- **THEN** system moves file to `<storage>/dotsync/opencode/agents/review.md`

### Requirement: System infers entry name from path patterns

The system SHALL infer the entry name using known path patterns.

#### Scenario: Infer from XDG config path

- **WHEN** user adds a file matching `~/.config/<name>/*`
- **THEN** system infers entry name as `<name>`
- **THEN** system sets root as `~/.config/<name>`

#### Scenario: Infer from dotfile in home

- **WHEN** user adds a file matching `~/.<name>` (e.g., `~/.zshrc`, `~/.gitconfig`)
- **THEN** system infers entry name as `<name>` (without leading dot)
- **THEN** system sets root as `~`

#### Scenario: Infer from hidden directory in home

- **WHEN** user adds a file matching `~/.<name>/*` (e.g., `~/.aws/config`, `~/.vscode/settings.json`, `~/.ssh/config`)
- **THEN** system infers entry name as `<name>` (without leading dot)
- **THEN** system sets root as `~/.<name>`

#### Scenario: Infer from Application Support path

- **WHEN** user adds a file matching `~/Library/Application Support/<name>/*`
- **THEN** system infers entry name as `<name>`
- **THEN** system sets root as `~/Library/Application Support/<name>`

#### Scenario: Prompt when pattern not recognized

- **WHEN** user adds a file that doesn't match known patterns (e.g., `~/Documents/notes.txt`)
- **THEN** system prompts: "Cannot infer name from path. Enter a name for this entry:"
- **THEN** user provides name
- **THEN** system uses provided name and sets root to file's parent directory

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
- **THEN** system displays warning: "macOS 14+ does NOT support symlinks for plist files. Use `Mackup` Copy mode (<https://github.com/lra/mackup?tab=readme-ov-file#copy-mode>) for these files."
- **THEN** system exits with zero-status (no-op).

### Requirement: System prevents entry conflicts

The system SHALL prevent conflicting entry configurations.

#### Scenario: File already tracked in same entry

- **WHEN** user runs `dotsync add <path>` for a file already in the manifest
- **THEN** system displays: "File already tracked in entry '<name>'"
- **THEN** system exits with zero status (no-op)

#### Scenario: Path conflicts with different entry

- **WHEN** user runs `dotsync add ~/.config/opencode/agents/review.md --name foo`
- **THEN** file path falls under root of existing entry "opencode"
- **THEN** system displays error: "Path conflicts with existing entry 'opencode'. Use that entry or choose a different name."
- **THEN** system exits with non-zero status

### Requirement: User can list tracked entries

The system SHALL display all tracked entries and their status.

#### Scenario: List all entries

- **WHEN** user runs `dotsync list`
- **THEN** system displays each entry name with its root and file count
- **THEN** system indicates link status for each entry on this machine

#### Scenario: List with details flag

- **WHEN** user runs `dotsync list --details`
- **THEN** system displays each entry with all tracked files
- **THEN** for each file, system shows: original path, cloud path, link status (linked/not linked/broken)

#### Scenario: Empty manifest

- **WHEN** user runs `dotsync list` with no tracked entries
- **THEN** system displays: "No entries tracked. Use 'dotsync add <file>' to start tracking."

### Requirement: Manifest format is versioned

The system SHALL maintain a versioned manifest format for forward compatibility.

#### Scenario: Create new manifest

- **WHEN** system creates a new manifest during init
- **THEN** manifest includes `"version": 1`
- **THEN** manifest includes empty `"entries": {}`

#### Scenario: Read manifest with matching version

- **WHEN** system reads manifest with `"version": 1`
- **THEN** system parses entries normally

#### Scenario: Read manifest with newer version

- **WHEN** system reads manifest with version > 1
- **THEN** system displays error: "Manifest version <N> not supported. Please upgrade dotsync."
- **THEN** system exits with non-zero status
