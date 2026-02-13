# Design: add-files-only

## Context

The `add` command currently accepts both files and directories. This change restricts `add` to accept only files, rejecting directories with a clear error. This simplifies the implementation and reduces potential user confusion about what's being synced.

## Goals / Non-Goals

### Goals

- Reject directory paths in the `add` command with a clear error message
- Exit with non-zero status when directory is provided
- Update CLI help text to clarify files-only policy

### Non-Goals

- Add support for recursive directory syncing
- Add bulk file addition (wildcards)
- Modify existing tracked entries

## Decisions

### Decision: Add directory check in ValidateForAdd

**Choice**: Add directory check in `internal/pathutil/validate.go` alongside existing validation logic

**Alternatives considered**:
- Check in `add.go` command handler (less reusable, violates locality of behavior)
- Check at multiple entry points (duplication)

**Rationale**: The `ValidateForAdd` function is the central place for all path validation. Adding the check there ensures consistency and reusability.

### Decision: Use ValidationError for directory rejection

**Choice**: Return a fatal ValidationError (not a warning) when directory is provided

**Alternatives considered**:
- Return warning and prompt (inconsistent with other fatal validations like symlinks)
- Return plain error (less consistent with existing pattern)

**Rationale**: Matches existing pattern for symlink rejection. Clear and consistent UX.

## Risks / Trade-offs

- None significant - this is a simple validation addition with minimal risk

## Migration Plan

1. Implement directory check in `ValidateForAdd`
2. Add unit test for directory rejection
3. Update CLI help text in `add.go`
4. Test manually with directory path

Rollback: Revert code change, no data migration needed.

## Open Questions

- None
