// Package backup handles temporary file backups during dotsync operations.
package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// BackupDir returns the path to the backup directory.
// Default: ~/.cache/dotsync/backups/
func BackupDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, ".cache", "dotsync", "backups"), nil
}

// EnsureBackupDir creates the backup directory if it doesn't exist.
func EnsureBackupDir() (string, error) {
	dir, err := BackupDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("creating backup directory: %w", err)
	}

	return dir, nil
}

// Backup represents a backup of a file.
type Backup struct {
	OriginalPath string
	BackupPath   string
}

// Create creates a backup of a file.
// Returns a Backup that can be used to restore or cleanup.
func Create(originalPath string) (*Backup, error) {
	dir, err := EnsureBackupDir()
	if err != nil {
		return nil, err
	}

	// Generate backup filename: timestamp-originalfilename
	timestamp := time.Now().Format("20060102-150405")
	filename := filepath.Base(originalPath)
	backupFilename := fmt.Sprintf("%s-%s", timestamp, filename)
	backupPath := filepath.Join(dir, backupFilename)

	// Copy the file
	if err := copyFile(originalPath, backupPath); err != nil {
		return nil, fmt.Errorf("creating backup: %w", err)
	}

	return &Backup{
		OriginalPath: originalPath,
		BackupPath:   backupPath,
	}, nil
}

// Restore restores the backup to the original location.
func (b *Backup) Restore() error {
	// Ensure parent directory exists
	parentDir := filepath.Dir(b.OriginalPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("creating parent directory: %w", err)
	}

	// Copy backup to original location
	if err := copyFile(b.BackupPath, b.OriginalPath); err != nil {
		return fmt.Errorf("restoring from backup: %w", err)
	}

	// Remove backup file
	os.Remove(b.BackupPath)

	return nil
}

// Cleanup removes the backup file.
func (b *Backup) Cleanup() error {
	return os.Remove(b.BackupPath)
}

// copyFile copies a file from src to dst, preserving permissions.
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

	destFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
