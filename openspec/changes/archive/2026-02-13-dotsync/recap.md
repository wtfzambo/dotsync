# Exploration Recap

## Problem Statement

Developers using multiple machines need to sync configuration files for various tools (AI coding assistants, editors, shell configs). Current solutions are either too manual (dotfiles git repos requiring commit/push/pull) or work but become tedious at scale (Google Drive + symlinks managed by hand).

The goal is a CLI tool that makes syncing config files as frictionless as possible—ideally "set it and forget it."

## Options Considered
    
### Option 1: Git-backed with Manual Sync (chezmoi/yadm style)

**Description**: Store configs in a git repo, user runs sync commands manually or via shell hooks.

**Pros**:
- Full version history
- Works offline
- User controls when sync happens

**Cons**:
- Requires remembering to sync
- If you modify a file and forget to push, changes don't propagate
- Shell hooks only trigger on new terminals, not on file save

### Option 2: Git-backed with File Watcher Daemon

**Description**: Background daemon watches tracked files, auto-commits and pushes on change.

**Pros**:
- True "set and forget" experience
- Changes sync automatically

**Cons**:
- Complex to implement (process management, startup, battery impact)
- Not weekend-scoped
- Platform-specific daemon management

### Option 3: Git-backed with OS-level File Watchers (launchd/systemd)

**Description**: Use OS-native file watching (launchd on macOS, systemd on Linux) to trigger sync on file change.

**Pros**:
- No custom daemon to manage
- OS handles reliability and restarts
- Reasonable to implement

**Cons**:
- Platform-specific implementation
- Windows support unclear (Task Scheduler?)
- Still requires git push/pull mechanics

### Option 4: Cloud Storage Backend (mackup style)

**Description**: Store configs in a cloud-synced folder (Google Drive, Dropbox, iCloud). Symlink files to their original locations. Cloud provider handles all sync.

**Pros**:
- Zero sync logic to implement—cloud provider does it
- Real-time sync across machines
- Proven model (mackup has 15k stars)
- Simple to implement

**Cons**:
- No version history (unless cloud provider offers it)
- Requires cloud storage subscription
- Dependent on cloud provider's sync reliability

## Decision & Rationale

**Going with Option 4: Cloud Storage Backend**

Deciding factors:
1. **Solves the hard problem**: Real-time sync is delegated to tools that already do it well
2. **Weekend-scoped**: No daemon, no git sync logic, just symlink management and a manifest
3. **Proven model**: mackup validates this approach works
4. **User already has the infrastructure**: Most developers already use Google Drive/Dropbox

Version history was deemed non-essential for MVP. Cloud providers offer some versioning, and users can always add git to the cloud folder later if they want.

## Codebase Findings

This is a new project—no existing codebase. However, research into prior art revealed:

**mackup** (15k stars):
- Uses cloud storage + symlinks
- Has both "copy mode" and "link mode"
- Link mode broken on macOS 14+ for `~/Library/Preferences/` due to Apple security changes
- Supports 500+ applications with pre-defined configs

**chezmoi** (18k stars):
- Git-backed, manual sync
- Very feature-rich (templates, encryption, machine-specific configs)
- No auto-sync

**gopass** (6.7k stars):
- Git-backed password manager
- Auto-push on mutation, manual pull
- Has `core.autopush` and `core.autosync` (interval-based) settings

**Key technical finding**: macOS 14+ breaks symlinks for `~/Library/Preferences/*.plist` files (CFPreferences API rejects writes through symlinks). This affects GUI app preferences but NOT:
- `~/.config/*` (XDG config directory)
- `~/.*` (dotfiles in home)
- Most developer tool configs

For dotsync's target users (developers syncing AI tool configs), this limitation is acceptable.

## Open Questions

- [ ] Google Drive path detection: Is it `~/Google Drive/` or `~/Library/CloudStorage/GoogleDrive-*/My Drive/`? Need to handle both.
- [ ] What happens if user runs `dotsync add` for a file that's already tracked? Update the entry? Error?
- [ ] Should `dotsync list` show files that exist in storage but aren't linked on this machine?
- [ ] How to handle the case where cloud folder doesn't exist yet on a new machine?

## References

**Tools researched**:
- [mackup](https://github.com/lra/mackup) - Cloud storage + symlinks approach
- [chezmoi](https://github.com/twpayne/chezmoi) - Git-backed dotfile manager
- [gopass](https://github.com/gopasspw/gopass) - Git-backed password manager with auto-sync
- [yadm](https://yadm.io/) - Git wrapper for dotfiles

**macOS symlink issue**:
- [mackup issue #2035](https://github.com/lra/mackup/issues/2035) - macOS Sonoma breaks symlinked preferences
- [mackup PR #2085](https://github.com/lra/mackup/pull/2085) - Copy mode workaround

**gopass sync config**:
- `core.autopush` (default: true) - Push after every mutation
- `core.autosync` (default: true, interval: 3 days) - Periodic fetch+push
