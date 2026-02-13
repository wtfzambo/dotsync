# Tasks: add-files-only

## 1. Implementation

- [x] 1.1 Add directory check in `internal/pathutil/validate.go` in `ValidateForAdd` function (check if `info.IsDir()` is true after symlink check)
- [x] 1.2 Update CLI help text in `cmd/dotsync/cmd/add.go` to clarify files-only policy

## 2. Testing

- [x] 2.1 Add unit test in `internal/pathutil/validate_test.go` for directory rejection
- [x] 2.2 Run existing tests to ensure no regressions

## 3. Manual Validation

- [x] 3.1 Test adding a directory with `dotsync add <directory-path>` and verify error message
- [x] 3.2 Test adding a file still works correctly
