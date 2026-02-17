package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wtfzambo/dotsync/internal/pathutil"
)

// ValidatePath checks if a path exists and is writable.
func ValidatePath(path string) error {
	// Expand home directory
	expanded := pathutil.ExpandHome(path)
	expanded = os.ExpandEnv(expanded)

	// Check if path exists
	info, err := os.Stat(expanded)
	if os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}
	if err != nil {
		return fmt.Errorf("checking path: %w", err)
	}

	// Check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	// Check if writable by attempting to create a temp file
	testFile := filepath.Join(expanded, ".dotsync-write-test")
	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("cannot write to storage path: %s", path)
	}
	f.Close()
	os.Remove(testFile)

	return nil
}

// EnsureDotsyncDir ensures the dotsync directory exists within the storage path.
// Returns the full path to the dotsync directory.
func EnsureDotsyncDir(storagePath string) (string, error) {
	expanded := pathutil.ExpandHome(storagePath)
	expanded = os.ExpandEnv(expanded)

	dotsyncDir := filepath.Join(expanded, "dotsync")
	if err := os.MkdirAll(dotsyncDir, 0755); err != nil {
		return "", fmt.Errorf("creating dotsync directory: %w", err)
	}

	return dotsyncDir, nil
}

// IsAvailable checks if the storage path is currently accessible.
func IsAvailable(storagePath string) bool {
	expanded := pathutil.ExpandHome(storagePath)
	expanded = os.ExpandEnv(expanded)

	_, err := os.Stat(expanded)
	return err == nil
}

// DotsyncDir returns the full path to the dotsync directory within storage.
func DotsyncDir(storagePath string) string {
	expanded := pathutil.ExpandHome(storagePath)
	expanded = os.ExpandEnv(expanded)
	return filepath.Join(expanded, "dotsync")
}

// ExpandPath expands ~ and environment variables in a path.
func ExpandPath(path string) string {
	expanded := pathutil.ExpandHome(path)
	return os.ExpandEnv(expanded)
}
