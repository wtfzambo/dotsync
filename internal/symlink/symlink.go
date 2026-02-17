// Package symlink handles symlink creation and management.
package symlink

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// Create creates a symlink at linkPath pointing to targetPath.
// Creates parent directories if needed.
func Create(linkPath, targetPath string) error {
	// Ensure parent directory exists
	parentDir := filepath.Dir(linkPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("creating parent directory: %w", err)
	}

	// Create the symlink
	if err := os.Symlink(targetPath, linkPath); err != nil {
		if runtime.GOOS == "windows" {
			return fmt.Errorf("creating symlink: %w\n\nWindows 10/11 requires Developer Mode or Administrator privileges to create symlinks.\nPlease enable Developer Mode in Settings > Privacy & Security > Developer Mode,\nor run this command as Administrator", err)
		}
		return fmt.Errorf("creating symlink: %w", err)
	}

	return nil
}

// Remove removes a symlink at the given path.
// Returns an error if the path is not a symlink.
func Remove(linkPath string) error {
	info, err := os.Lstat(linkPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Already gone
		}
		return fmt.Errorf("checking path: %w", err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("not a symlink: %s", linkPath)
	}

	return os.Remove(linkPath)
}

// IsSymlink checks if the given path is a symlink.
func IsSymlink(path string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.Mode()&os.ModeSymlink != 0, nil
}

// ReadTarget returns the target of a symlink.
func ReadTarget(linkPath string) (string, error) {
	return os.Readlink(linkPath)
}

// Status represents the status of a symlink.
type Status int

const (
	StatusNotExist  Status = iota // Path doesn't exist
	StatusLinked                  // Symlink exists and points to correct target
	StatusBroken                  // Symlink exists but target is missing
	StatusIncorrect               // Symlink exists but points to wrong target
	StatusNotLinked               // Regular file exists (not a symlink)
)

// String returns a human-readable status.
func (s Status) String() string {
	switch s {
	case StatusNotExist:
		return "not exist"
	case StatusLinked:
		return "linked"
	case StatusBroken:
		return "broken"
	case StatusIncorrect:
		return "incorrect"
	case StatusNotLinked:
		return "not linked"
	default:
		return "unknown"
	}
}

// Check checks the status of a path that should be a symlink to expectedTarget.
func Check(linkPath, expectedTarget string) (Status, string, error) {
	info, err := os.Lstat(linkPath)
	if err != nil {
		if os.IsNotExist(err) {
			return StatusNotExist, "", nil
		}
		return 0, "", err
	}

	// Check if it's a symlink
	if info.Mode()&os.ModeSymlink == 0 {
		return StatusNotLinked, "", nil
	}

	// Read the target
	actualTarget, err := os.Readlink(linkPath)
	if err != nil {
		return 0, "", err
	}

	// Check if target exists
	if _, err := os.Stat(linkPath); err != nil {
		if os.IsNotExist(err) {
			return StatusBroken, actualTarget, nil
		}
		return 0, actualTarget, err
	}

	// Check if target matches
	if actualTarget != expectedTarget {
		// Also try resolving to absolute paths
		absExpected, _ := filepath.Abs(expectedTarget)
		absActual, _ := filepath.Abs(actualTarget)
		if absExpected != absActual {
			return StatusIncorrect, actualTarget, nil
		}
	}

	return StatusLinked, actualTarget, nil
}

// MoveFile moves a file from src to dst, creating parent directories if needed.
func MoveFile(src, dst string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("creating destination directory: %w", err)
	}

	// Try rename first (fastest, works on same filesystem)
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// Fall back to copy + delete (for cross-filesystem moves)
	if err := copyFile(src, dst); err != nil {
		return err
	}

	if err := os.Remove(src); err != nil {
		os.Remove(dst)
		return fmt.Errorf("removing original file: %w", err)
	}

	return nil
}

// CopyFile copies a file from src to dst, preserving permissions.
func CopyFile(src, dst string) error {
	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	destFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
