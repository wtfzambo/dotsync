package pathutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/wtfzambo/dotsync/internal/manifest"
)

// ValidationError represents an error during file validation.
type ValidationError struct {
	Path    string
	Message string
	IsWarn  bool // If true, this is a warning, not a fatal error
}

func (e ValidationError) Error() string {
	return e.Message
}

// ValidateForAdd checks if a file can be added to dotsync.
// Returns an error if validation fails, or a warning ValidationError if there's a non-fatal issue.
func ValidateForAdd(absPath string) error {
	// Check if file exists
	info, err := os.Lstat(absPath) // Use Lstat to detect symlinks
	if os.IsNotExist(err) {
		return ValidationError{
			Path:    absPath,
			Message: fmt.Sprintf("file not found: %s", absPath),
		}
	}
	if err != nil {
		return fmt.Errorf("checking file: %w", err)
	}

	// Check if it's a symlink
	if info.Mode()&os.ModeSymlink != 0 {
		return ValidationError{
			Path:    absPath,
			Message: "cannot track symlinks. If this is already synced elsewhere, unlink it first",
		}
	}

	// Check if it's a directory
	if info.IsDir() {
		return ValidationError{
			Path:    absPath,
			Message: "cannot add directories. Use 'dotsync add <file>' to add individual files.",
		}
	}

	// Check for macOS plist files
	if runtime.GOOS == "darwin" && isPlistFile(absPath) {
		return ValidationError{
			Path:    absPath,
			Message: "macOS 14+ does NOT support symlinks for plist files. Use Mackup Copy mode (https://github.com/lra/mackup?tab=readme-ov-file#copy-mode) for these files",
		}
	}

	// Check if outside home directory (warning, not error)
	if !IsUnderHome(absPath) {
		return ValidationError{
			Path:    absPath,
			Message: "file is outside home directory. Symlinks may not work as expected if paths differ across machines",
			IsWarn:  true,
		}
	}

	return nil
}

// isPlistFile checks if a path is a macOS plist file in ~/Library/Preferences/
func isPlistFile(absPath string) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	prefsDir := filepath.Join(home, "Library", "Preferences")
	if !strings.HasPrefix(absPath, prefsDir) {
		return false
	}

	return strings.HasSuffix(strings.ToLower(absPath), ".plist")
}

// CheckEntryConflict checks if adding a file would conflict with an existing entry.
// Returns the conflicting entry name if there's a conflict, empty string otherwise.
func CheckEntryConflict(absPath string, explicitName string, m *manifest.Manifest) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	for name, entry := range m.Entries {
		// Expand the entry root
		entryRoot := ExpandHome(entry.Root)

		// Check if the file is under this entry's root
		if strings.HasPrefix(absPath, entryRoot+string(filepath.Separator)) || absPath == entryRoot {
			// File is under this entry's root
			if explicitName != "" && explicitName != name {
				// User specified a different name, but file is under existing entry
				return name, nil
			}
			// File belongs to this entry (no conflict, will be added to it)
			return "", nil
		}

		// Check if any of the entry's files match this path
		for _, f := range entry.Files {
			fullPath := filepath.Join(entryRoot, f)
			if fullPath == absPath {
				// File is already tracked
				return name, nil
			}
		}
	}

	// Also check if a new root would conflict with existing entries
	inferred := InferFromPath(absPath)
	if inferred != nil && explicitName == "" {
		inferredRoot := ExpandHome(inferred.Root)
		for name, entry := range m.Entries {
			entryRoot := ExpandHome(entry.Root)
			// Check if roots would overlap
			if strings.HasPrefix(inferredRoot, entryRoot+string(filepath.Separator)) ||
				strings.HasPrefix(entryRoot, inferredRoot+string(filepath.Separator)) {
				return name, nil
			}
		}
	}

	_ = home // silence unused warning
	return "", nil
}

// IsAlreadyTracked checks if a file is already tracked in the manifest.
// Returns the entry name if tracked, empty string otherwise.
func IsAlreadyTracked(absPath string, m *manifest.Manifest) string {
	for name, entry := range m.Entries {
		entryRoot := ExpandHome(entry.Root)
		for _, f := range entry.Files {
			fullPath := filepath.Join(entryRoot, f)
			if fullPath == absPath {
				return name
			}
		}
	}
	return ""
}
