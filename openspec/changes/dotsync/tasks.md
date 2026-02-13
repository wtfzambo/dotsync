# Tasks: dotsync

## 1. Project Setup

- [x] 1.1 Initialize Go module (`go mod init github.com/<user>/dotsync`)
- [x] 1.2 Set up project structure (cmd/, internal/, pkg/)
- [x] 1.3 Add CLI framework (cobra or similar)
- [x] 1.4 Configure basic build and test scripts

## 2. Core Data Types

- [x] 2.1 Define manifest types (Manifest, Entry structs with JSON tags)
- [x] 2.2 Define local config types (Config struct)
- [x] 2.3 Implement manifest read/write with version checking
- [x] 2.4 Implement local config read/write (~/.config/dotsync/config.json)

## 3. Storage Detection (storage-detection spec)

- [x] 3.1 Define cloud storage path constants per platform (macOS, Linux)
- [x] 3.2 Implement provider detection functions (Google Drive, Dropbox, iCloud)
- [x] 3.3 Implement path validation (exists, writable)
- [x] 3.4 Implement `init` command with provider argument
- [x] 3.5 Add prompt for manual path when auto-detection fails
- [x] 3.6 Add reinitialize confirmation when config exists

## 4. Path Inference (file-tracking spec)

- [x] 4.1 Implement path pattern matching for XDG config (~/.config/<name>/*)
- [x] 4.2 Implement pattern matching for dotfiles in home (~/.<name>)
- [x] 4.3 Implement pattern matching for hidden directories (~/.<name>/*)
- [x] 4.4 Implement pattern matching for Application Support (macOS)
- [x] 4.5 Implement prompt fallback when no pattern matches
- [x] 4.6 Add entry conflict detection (path under existing entry's root)

## 5. File Validation (file-tracking spec)

- [x] 5.1 Implement file existence check
- [x] 5.2 Implement symlink detection (reject symlinks)
- [x] 5.3 Implement outside-home-directory warning with confirmation
- [x] 5.4 Implement macOS plist path rejection with mackup suggestion
- [x] 5.5 Implement "already tracked" detection (no-op case)

## 6. Backup System (symlink-management spec)

- [x] 6.1 Implement backup directory creation (~/.cache/dotsync/backups/)
- [x] 6.2 Implement backup file creation with timestamp
- [x] 6.3 Implement restore-from-backup on failure
- [x] 6.4 Implement backup cleanup on success

## 7. Add Command (file-tracking + symlink-management specs)

- [x] 7.1 Implement `add` command argument parsing (path, --name flag)
- [x] 7.2 Wire up validation pipeline (exists, not symlink, not plist, etc.)
- [x] 7.3 Implement file move to cloud storage with directory creation
- [x] 7.4 Implement symlink creation at original location
- [x] 7.5 Implement manifest update (new entry or add to existing)
- [x] 7.6 Implement atomic operation with rollback on failure

## 8. Link Command (symlink-management spec)

- [x] 8.1 Implement `link` command argument parsing (optional entry name)
- [x] 8.2 Implement entry lookup from manifest
- [x] 8.3 Implement existing-file conflict detection
- [x] 8.4 Implement interactive prompt (backup/skip/abort)
- [x] 8.5 Implement --backup flag for non-interactive mode
- [x] 8.6 Implement symlink creation with parent directory creation
- [x] 8.7 Implement summary output (linked, skipped files)

## 9. Unlink Command (symlink-management spec)

- [x] 9.1 Implement `unlink` command argument parsing (optional entry name)
- [x] 9.2 Implement file copy from cloud to original location
- [x] 9.3 Implement symlink removal
- [x] 9.4 Handle broken symlink case (remove symlink, warn about missing file)
- [x] 9.5 Preserve manifest entries (don't remove on unlink)

## 10. List Command (file-tracking spec)

- [x] 10.1 Implement `list` command basic output (entries with file counts)
- [x] 10.2 Implement symlink status checking (linked/not linked/broken/incorrect)
- [x] 10.3 Implement --details flag for per-file output
- [x] 10.4 Implement empty manifest message

## 11. Storage Availability Checks

- [x] 11.1 Implement pre-operation storage availability check
- [x] 11.2 Implement "not initialized" error for missing config
- [x] 11.3 Implement "storage unavailable" error for missing cloud folder

## 12. Testing

- [x] 12.1 Add unit tests for path inference logic
- [x] 12.2 Add unit tests for manifest read/write
- [x] 12.3 Add unit tests for validation functions
- [x] 12.4 Add integration tests for init command (with mock filesystem)
- [x] 12.5 Add integration tests for add/link/unlink cycle
- [x] 12.6 Add integration tests for edge cases (conflicts, broken symlinks)

## 13. Documentation & Polish

- [x] 13.1 Write README with installation and usage instructions
- [x] 13.2 Add --help text for all commands
- [x] 13.3 Add error messages review (clear, actionable)
- [x] 13.4 Cross-platform testing (macOS + Linux)
