# Exploration Recap

## Problem Statement

The current `add` command accepts both files and directories. While convenient for some use cases, directory tracking introduces significant complexity: recursive syncing of all files within, preserving entire directory structures, handling nested symlinks, and ensuring proper management of all contained files. Most developers only need to sync specific config files, not entire directories. The MVP originally allowed directories for flexibility, but this adds implementation burden and potential user confusion.

## Options Considered

### Option 1: Keep allowing directories (status quo)

**Description**: Continue accepting both files and directories in the `add` command.

**Pros**:
- Maximum flexibility for users
- No breaking changes

**Cons**:
- Complex implementation: recursive file handling, nested symlinks
- Harder to test all edge cases
- Potential confusion about what's synced
- Directory structures may differ across machines

### Option 2: Reject directories (files only)

**Description**: Modify `add` command to reject directory paths with a clear error message.

**Pros**:
- Simpler implementation
- More predictable behavior
- Easier to test and maintain
- Aligns with most common use case (individual dotfiles)

**Cons**:
- Breaking change for users currently using directory tracking
- Users must add files individually

## Decision & Rationale

We chose **Option 2 (files only)** because:
- The implementation is significantly simpler - just add a directory check in validation
- Most users sync individual config files anyway
- The change is easy to document and communicate
- Can revisit directory support later if demand exists

## Codebase Findings

- **Add command**: `cmd/dotsync/cmd/add.go` - handles file/directory addition
- **Validation logic**: `internal/pathutil/validate.go` - `ValidateForAdd()` function (lines 24-65)
  - Currently validates: file exists, not a symlink, not macOS plist, warning for outside-home
  - **Location to add directory check**: After symlink check (line 45), before plist check
- **Tests**: `internal/pathutil/validate_test.go` - existing validation tests to extend

## Open Questions

- None - the implementation path is straightforward

## References

- `openspec/specs/file-tracking/spec.md` - existing spec for file tracking capability
- Prior art: chezmoi and mackup both focus on individual file/dotfile tracking
