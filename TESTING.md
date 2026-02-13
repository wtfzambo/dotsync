# Cross-Platform Testing Guide

This document describes how to test dotsync across different platforms (macOS, Linux) to ensure compatibility.

## Platform-Specific Paths

### Google Drive

| Platform | Path | Auto-detection |
|----------|------|----------------|
| macOS | `~/Library/CloudStorage/GoogleDrive-*/My Drive` | Yes (glob pattern) |
| macOS (fallback) | `~/Google Drive` | Yes |
| Linux | `~/Google Drive` | Yes |
| Linux (fallback) | `~/google-drive` | Yes |

### Dropbox

| Platform | Path | Auto-detection |
|----------|------|----------------|
| macOS | `~/Library/CloudStorage/Dropbox` | Yes |
| macOS (fallback) | `~/Dropbox` | Yes |
| Linux | `~/Dropbox` | Yes |

### iCloud Drive

| Platform | Path | Auto-detection |
|----------|------|----------------|
| macOS | `~/Library/Mobile Documents/com~apple~CloudDocs` | Yes |
| Linux | Not available | N/A |

## Test Matrix

The following tests should be run on each platform to ensure full compatibility.

### Build Tests

| Test | macOS | Linux | Notes |
|------|-------|-------|-------|
| `task build` succeeds | ✓ | ✓ | Verify binary is created in `bin/dotsync` |
| `task dev` succeeds | ✓ | ✓ | Fast build without optimizations |
| Binary runs | ✓ | ✓ | `./bin/dotsync --help` |
| All tests pass | ✓ | ✓ | `task test` |

### Init Command Tests

| Test | macOS | Linux | Notes |
|------|-------|-------|-------|
| Google Drive auto-detection | ✓ | ✓ | `dotsync init gdrive` |
| Dropbox auto-detection | ✓ | ✓ | `dotsync init dropbox` |
| iCloud auto-detection | ✓ | ✗ | `dotsync init icloud` (macOS only) |
| Manual path initialization | ✓ | ✓ | `dotsync init --path ~/test-path` |
| Invalid provider error | ✓ | ✓ | `dotsync init invalid` should error |
| Reinitialize prompt | ✓ | ✓ | Run `init` twice, verify prompt |

### Add Command Tests

#### XDG Config Paths

| Test | macOS | Linux | Notes |
|------|-------|-------|-------|-------|
| Add file in `~/.config/` | ✓ | ✓ | `dotsync add ~/.config/test/config.json` |
| Entry name inference | ✓ | ✓ | Should infer entry name as "test" |

#### Dotfiles

| Test | macOS | Linux | Notes |
|------|-------|-------|-------|
| Add `~/.zshrc` | ✓ | ✓ | Common shell config |
| Add `~/.bashrc` | ✓ | ✓ | Common shell config |
| Add `~/.vimrc` | ✓ | ✓ | Common editor config |
| Entry name inference | ✓ | ✓ | Should infer or prompt |

#### Hidden Directories

| Test | macOS | Linux | Notes |
|------|-------|-------|-------|
| Add `~/.ssh/config` | ✓ | ✓ | Sensitive config |
| Add `~/.aws/credentials` | ✓ | ✓ | AWS credentials |
| Custom entry name | ✓ | ✓ | Use `--name` flag |

#### macOS-Specific Paths

| Test | macOS | Linux | Notes |
|------|-------|-------|-------|
| Add Application Support file | ✓ | N/A | `~/Library/Application Support/app/config.json` |
| Reject plist files | ✓ | N/A | `~/Library/Preferences/com.app.plist` should error |
| Plist error message clear | ✓ | N/A | Should mention Mackup Copy mode |

#### Edge Cases

| Test | macOS | Linux | Notes |
|------|-------|-------|-------|
| Add non-existent file | ✓ | ✓ | Should error with clear message |
| Add file outside home | ✓ | ✓ | Should warn, prompt for confirmation |
| Add already-tracked file | ✓ | ✓ | Should skip with message |
| Add symlink | ✓ | ✓ | Should error |

### Link Command Tests

| Test | macOS | Linux | Notes |
|------|-------|-------|-------|
| Link all entries | ✓ | ✓ | `dotsync link` |
| Link specific entry | ✓ | ✓ | `dotsync link entryname` |
| Link with existing file | ✓ | ✓ | Prompt for backup/skip/abort |
| Link with `--backup` flag | ✓ | ✓ | Auto-backup conflicts |
| Link broken symlink | ✓ | ✓ | Should replace broken symlink |
| Link incorrect symlink | ✓ | ✓ | Should prompt to replace |

### Unlink Command Tests

| Test | macOS | Linux | Notes |
|------|-------|-------|-------|
| Unlink all entries | ✓ | ✓ | `dotsync unlink` |
| Unlink specific entry | ✓ | ✓ | `dotsync unlink entryname` |
| Verify files are regular files | ✓ | ✓ | Check with `ls -l` |
| Verify cloud files intact | ✓ | ✓ | Cloud storage still has files |
| Unlink when not linked | ✓ | ✓ | Should skip gracefully |
| Unlink broken symlink | ✓ | ✓ | Should remove symlink |

### List Command Tests

| Test | macOS | Linux | Notes |
|------|-------|-------|-------|
| List empty manifest | ✓ | ✓ | Should show helpful message |
| List entries overview | ✓ | ✓ | `dotsync list` |
| List with details | ✓ | ✓ | `dotsync list --details` |
| Status icons correct | ✓ | ✓ | [ok], [not lnk], [broken], etc. |
| Status summary correct | ✓ | ✓ | "all linked", "2 linked, 1 not linked", etc. |

### Full Workflow Tests

#### Single Machine Workflow

| Test | macOS | Linux | Notes |
|------|-------|-------|-------|
| Init → Add → List | ✓ | ✓ | Complete workflow |
| Add → Unlink → Link | ✓ | ✓ | Verify round-trip |
| Add → Verify symlink created | ✓ | ✓ | Check with `ls -l` |
| Verify file moved to cloud | ✓ | ✓ | Check cloud storage path |

#### Multi-Machine Workflow

This requires access to two machines (or VMs) with the same cloud storage account.

| Test | Machines | Notes |
|------|----------|-------|
| Add on Machine A | A: macOS, B: macOS | |
| Wait for cloud sync | Both | |
| Init on Machine B | A: macOS, B: macOS | Same provider |
| Link on Machine B | A: macOS, B: macOS | Verify files appear |
| Modify file on Machine B | Both | Edit synced file |
| Verify change syncs to A | Both | Cloud provider syncs |
| Cross-platform: macOS to Linux | A: macOS, B: Linux | Test path compatibility |
| Cross-platform: Linux to macOS | A: Linux, B: macOS | Test path compatibility |

## Known Platform Differences

### macOS

- **Application Support:** macOS uses `~/Library/Application Support/` for app configs. This path is valid for syncing.
- **Preferences (plist files):** macOS 14+ does NOT support symlinks for plist files in `~/Library/Preferences/`. These files are rejected by dotsync with a helpful error message.
- **iCloud Drive:** Only available on macOS. The path is `~/Library/Mobile Documents/com~apple~CloudDocs`.

### Linux

- **XDG Base Directory:** Linux commonly uses `~/.config/` for application configs (XDG Base Directory specification).
- **Google Drive:** Linux users typically mount Google Drive at `~/Google Drive` or `~/google-drive` using tools like `google-drive-ocamlfuse` or `rclone`.
- **iCloud Drive:** Not available natively on Linux. Attempting to use iCloud on Linux should error gracefully.

### Path Compatibility

Files synced across platforms must have compatible paths:

- **Safe:** Files in `~/.config/`, `~/.*` (dotfiles in home)
- **Risky:** Absolute paths outside home directory (e.g., `/usr/local/bin/script`)
- **Platform-specific:** macOS `~/Library/` paths won't work on Linux

## Automated CI Suggestions

To automate testing across platforms, consider using GitHub Actions:

```yaml
name: CI

on: [push, pull_request]

jobs:
  test:
    strategy:
      matrix:
        os: [macos-latest, ubuntu-latest]
        go: ['1.21']
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go }}
      - name: Build
        run: |
          go build -o bin/dotsync ./cmd/dotsync
      - name: Test
        run: go test -v ./...
      - name: Verify help commands
        run: |
          ./bin/dotsync --help
          ./bin/dotsync init --help
          ./bin/dotsync add --help
          ./bin/dotsync link --help
          ./bin/dotsync unlink --help
          ./bin/dotsync list --help
```

### Test Strategy

1. **Unit tests:** Run on all platforms via `go test ./...`
2. **Integration tests:** Test commands with real filesystem operations
3. **Platform-specific tests:** Conditionally run based on `runtime.GOOS`
4. **Manual testing:** Multi-machine sync requires manual verification

## Test Execution Checklist

Before releasing a new version:

- [ ] All unit tests pass on macOS
- [ ] All unit tests pass on Linux
- [ ] Build succeeds on both platforms
- [ ] Help text is comprehensive and accurate
- [ ] Init auto-detection works for all providers (platform-specific)
- [ ] Full add/link/unlink cycle works
- [ ] Multi-machine sync verified (at least one pair)
- [ ] Error messages are clear and actionable
- [ ] Edge cases handled gracefully (non-existent files, symlinks, etc.)

## Reporting Issues

If you find platform-specific bugs:

1. Note the exact platform (OS, version)
2. Include cloud storage provider and path
3. Provide full error messages
4. Include steps to reproduce
5. Check if issue is platform-specific or general

## Future Platform Support

### Windows

Windows support is planned but not yet implemented. Key differences:

- Different path separators (`\` vs `/`)
- Different cloud storage mount points
- Symlink support requires admin privileges (or Developer Mode in Windows 10+)
- Different home directory structure

When implementing Windows support:

- Test on Windows 10 and Windows 11
- Test with and without admin privileges
- Test with Developer Mode enabled/disabled
- Verify all three cloud providers work
- Check symlink permissions and creation
