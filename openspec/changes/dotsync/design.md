# Design: dotsync

## Context

dotsync is a new CLI tool that syncs developer config files across machines by leveraging cloud storage providers. The core insight is that we don't need to solve sync—Google Drive, Dropbox, and iCloud already do this well. We just need symlink management and a manifest.

Target users are developers with multiple machines who want their AI tool configs (OpenCode, Cursor, etc.) and dotfiles to stay in sync without manual git workflows.

Constraints:
- Weekend project scope—keep it simple
- Go single binary, no dependencies
- macOS first, then Linux, eventually Windows
- macOS 14+ has symlink limitations for `~/Library/Preferences/` (acceptable for our use case)

## Goals / Non-Goals

### Goals

- Enable tracking config files by moving them to cloud storage and creating symlinks
- Support multiple files per tool (e.g., `~/.config/opencode/config.json` and `~/.config/opencode/agents/review.md`)
- Preserve directory structure relative to app root
- Detect common cloud storage locations automatically
- Provide clear feedback when symlinks break or cloud storage is unavailable

### Non-Goals

- Custom sync implementation (no file watching, no git)
- Version history (cloud provider handles this if at all)
- Encryption (user can add this to cloud provider)
- Pre-defined app configs like mackup (user specifies paths explicitly)
- Windows support in MVP
- GUI or menu bar integration

## Data Structures

### Manifest (`.dotsync.json`)

Location: `<cloud-folder>/dotsync/.dotsync.json`

```json
{
  "version": 1,
  "entries": {
    "opencode": {
      "root": "~/.config/opencode",
      "files": [
        "config.json",
        "agents/review.md"
      ]
    },
    "zsh": {
      "root": "~",
      "files": [".zshrc"]
    },
    "cursor": {
      "root": "~/.cursor",
      "files": [
        "settings.json",
        "keybindings.json"
      ]
    }
  }
}
```

Key design choices:
- `root` stores the original parent directory (expanded `~` during operations)
- `files` are relative paths from root, preserving directory structure
- Name (e.g., "opencode") is the grouping key, inferred or user-provided

### Storage Structure

```
<cloud-folder>/dotsync/
├── .dotsync.json           # manifest
├── opencode/
│   ├── config.json
│   └── agents/
│       └── review.md
├── zsh/
│   └── .zshrc
└── cursor/
    ├── settings.json
    └── keybindings.json
```

### Local Config (`~/.config/dotsync/config.json`)

Stores machine-specific settings (not synced):

```json
{
  "storagePath": "~/Library/CloudStorage/GoogleDrive-user@gmail.com/My Drive"
}
```

## Decisions

### Decision: Cloud storage detection strategy

**Choice**: Ask user which Cloud storage they're using during `init` command.

**Behavior**:
`dotsync init <cloud-storage>`
Cloud storage can be one of 3 options: Google drive, dropbox, icloud
Then try to find the storage mount in common locations, and ask user where it is if it can't be found. Options

**Alternatives considered**:
- Require explicit path always (worse UX for common cases)
- Only support one provider (limits adoption)
- Use environment variables (non-obvious for users)

**Rationale**: Most users have one cloud provider. Auto-detection reduces friction. Prompting when ambiguous prevents wrong choice.

### Decision: Name inference from path

**Choice**: Infer name from the closest recognizable directory in the path

**Inference rules**:
1. For `~/.config/<name>/*` → use `<name>` (XDG_CONFIG_HOME only, not DATA/STATE/CACHE)
2. For `~/.<name>` (dotfiles like `.zshrc`) → use `<name>` without dot, root is `~`
3. For `~/.<name>/*` (hidden dirs like `.aws/config`) → use `<name>` without dot, root is `~/.<name>`
4. For `~/Library/Application Support/<name>/*` → use `<name>`
5. Otherwise → prompt user

**Examples**:
- `~/.config/opencode/agents/review.md` → name: `opencode`, root: `~/.config/opencode`, file: `agents/review.md`
- `~/.zshrc` → name: `zsh`, root: `~`, file: `.zshrc`
- `~/.aws/config` → name: `aws`, root: `~/.aws`, file: `config`
- `~/.vscode/settings.json` → name: `vscode`, root: `~/.vscode`, file: `settings.json`
- `~/some/random/path.txt` → prompt user for name

**Alternatives considered**:
- Always require explicit name (more typing)
- Use full path as name (ugly, verbose)
- Use filename (too generic, collisions)

**Rationale**: Developers have strong mental models—"my opencode config", "my zsh config". Inferring aligns with how they think about their files.

### Decision: Handling existing files on `link` command

**Choice**: Prompt user with three options: backup, skip, abort

**Behavior**:
```
File already exists: ~/.config/opencode/config.json
  [b] Backup to ~/.config/opencode/config.json.dotsync-backup
  [s] Skip this file
  [a] Abort entire operation
Choice:
```

**Alternatives considered**:
- Always overwrite (data loss risk)
- Always backup (clutters filesystem)
- Always abort (frustrating for partial setups)
- `--force` flag only (easy to forget, potential data loss)

**Rationale**: Config files are valuable. User should decide per-file. Backup is non-destructive default.

### Decision: Symlink direction

**Choice**: Original location becomes symlink pointing to cloud storage

```
~/.config/opencode/config.json → ~/Google Drive/dotsync/opencode/config.json
```

**Alternatives considered**:
- Cloud storage contains symlinks to original (breaks sync—symlinks don't sync file content)
- Copy-on-change (requires file watching, defeats simplicity goal)

**Rationale**: This is how mackup works. Cloud storage has the real file, which syncs. Original location is just a pointer.

### Decision: What `add` does when file already tracked

**Choice**: If same name, add to existing entry. If different name, error.

**Behavior**:
- `add ~/.config/opencode/config.json` → creates entry "opencode"
- `add ~/.config/opencode/agents/review.md` → adds to existing "opencode" entry
- `add ~/.config/opencode/agents/review.md --name foo` → error: "File is under root for existing entry 'opencode'"

**Alternatives considered**:
- Allow overlapping entries (confusing, which symlink wins?)
- Silent no-op (user doesn't know if it worked)
- Update entry silently (unexpected behavior)

**Rationale**: One file can only belong to one entry. Inferring grouping by path keeps things organized.

## Risks / Trade-offs

- [Cloud storage not installed/configured] → Detect and show clear error with instructions per provider
- [Symlink breaks if cloud folder moves] → `dotsync status` command shows broken symlinks, user can re-run `link`
- [macOS 14+ plist symlink issue] → Reject adding `~/Library/Preferences/*.plist` paths, suggest mackup copy mode
- [Conflicting edits on two machines] → Delegate to cloud provider's conflict resolution (typically creates duplicate file)
- [User deletes file from cloud storage] → Symlink becomes dangling; `status` shows this, user can `unlink` and re-add
- [Circular symlinks or symlink-to-symlink] → Detect and refuse to track files that are already symlinks
- [Move to cloud fails mid-operation] → Create temporary backup in `~/.cache/dotsync/backups/` before move, restore on failure

## Edge Cases

### Adding a file that's already a symlink

Refuse with clear error: "Cannot track symlinks. If this is already synced elsewhere, unlink it first."

### Adding a file outside home directory

Allow but warn: "File is outside home directory. Symlinks may not work as expected if paths differ across machines."

### Running `link` when cloud storage is offline

Fail gracefully: Check if storage path exists before any operations. Show: "Cloud storage not available at <path>. Is it mounted?"

### Running `add` on a new machine (no manifest yet)

Fail with guidance: "No dotsync storage found. Run `dotsync init` first, or ensure cloud storage is synced."

### File path differs across machines (Linux vs macOS)

MVP: Don't handle. Document that paths must be identical across machines. Future: Per-platform path mappings.

## Platform-Specific Concerns

### Cloud Storage Paths (as code constants)

The system maintains a structured list of known cloud storage paths per platform:

**macOS**:
| Provider | Primary Path | Fallback |
|----------|--------------|----------|
| Google Drive | `~/Library/CloudStorage/GoogleDrive-*/My Drive/` | `~/Google Drive/` |
| Dropbox | `~/Library/CloudStorage/Dropbox/` | `~/Dropbox/` |
| iCloud | `~/Library/Mobile Documents/com~apple~CloudDocs/` | - |

**Linux**:
| Provider | Primary Path | Fallback |
|----------|--------------|----------|
| Dropbox | `~/Dropbox/` | `~/.dropbox-dist/` |
| Google Drive | `~/Google Drive/` | `~/google-drive/` |

**Windows** (future - PATHS TBD!!!):
| Provider | Path |
|----------|------|
| Google Drive | `%USERPROFILE%\Google Drive\` |
| Dropbox | `%USERPROFILE%\Dropbox\` |
| iCloud | `%USERPROFILE%\iCloudDrive\` |
| OneDrive | `%USERPROFILE%\OneDrive\` |

### macOS

- Cloud storage paths use `~/Library/CloudStorage/` format (newer) or `~/Dropbox/` etc. (older)
- `~/Library/Preferences/*.plist` symlinks broken on macOS 14+ — reject these paths, suggest mackup copy mode
- Consider notarization for distribution (future)

### Linux

- Cloud storage clients vary (rclone, Insync, google-drive-ocamlfuse, native Dropbox)
- XDG paths (`~/.config/`) work well with symlinks

### Windows (future)

- Cloud storage usually in `C:\Users\<user>\OneDrive\` or similar
- Symlinks require admin or developer mode
- Path separators differ (handle in path normalization)

## Migration Plan

N/A - New project, no existing users to migrate.

## Open Questions

- [ ] Should `list` differentiate between "tracked and linked on this machine" vs "tracked but not linked"?
- [ ] Should we support glob patterns for `add` (e.g., `add ~/.config/opencode/*.json`)?
- [ ] How to handle the case where the same file is added from two different machines with different paths?
- [ ] Should `init` create the storage folder if it doesn't exist, or require it to exist?
