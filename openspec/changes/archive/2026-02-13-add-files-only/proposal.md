# Proposal: add-files-only

## Why

The current `add` command accepts both files and directories, but directory tracking introduces significant complexity: recursive syncing, preserving entire directory structures, handling nested symlinks, and ensuring all files within are properly managed. Most developers only need to sync specific config files, not entire directories. Preventing directory additions simplifies the implementation and reduces potential user confusion.

## What Changes

- Add validation to reject directory paths in the `add` command
- Display clear error message when user attempts to add a directory
- Exit with non-zero status when directory is provided

## Capabilities

### New Capabilities

- None (this is a constraint on existing behavior)

### Modified Capabilities

- `file-tracking`: Add requirement to reject directory paths in the `add` command

## Impact

- Command: `add` command validation logic
- Error handling: New validation branch for directory detection
- Documentation: Update CLI help text to clarify files-only policy
