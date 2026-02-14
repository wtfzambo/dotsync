package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/wtfzambo/dotsync/internal/manifest"
	"github.com/wtfzambo/dotsync/internal/symlink"
)

func TestAddLinkUnlinkCycle(t *testing.T) {
	tmpHome := t.TempDir()
	tmpStorage := t.TempDir()

	configDir := filepath.Join(tmpHome, ".config", "testapp")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	testFile := filepath.Join(configDir, "config.json")
	testContent := []byte(`{"test": "value"}`)
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	m := manifest.New()
	dotsyncDir := filepath.Join(tmpStorage, "dotsync")
	if err := os.MkdirAll(dotsyncDir, 0755); err != nil {
		t.Fatalf("failed to create dotsync dir: %v", err)
	}

	t.Run("add_file", func(t *testing.T) {
		entryName := "testapp"
		relPath := "config.json"
		destPath := filepath.Join(tmpStorage, "dotsync", entryName, relPath)

		if err := symlink.MoveFile(testFile, destPath); err != nil {
			t.Fatalf("failed to move file: %v", err)
		}

		if err := symlink.Create(testFile, destPath); err != nil {
			t.Fatalf("failed to create symlink: %v", err)
		}

		m.AddFile(entryName, filepath.Join("~", ".config", "testapp"), relPath)
		if err := m.Save(tmpStorage); err != nil {
			t.Fatalf("failed to save manifest: %v", err)
		}

		status, target, err := symlink.Check(testFile, destPath)
		if err != nil {
			t.Fatalf("failed to check symlink: %v", err)
		}
		if status != symlink.StatusLinked {
			t.Errorf("status = %v, want %v (target: %s)", status, symlink.StatusLinked, target)
		}

		if _, err := os.Stat(destPath); err != nil {
			t.Errorf("file not found in storage: %v", err)
		}

		content, err := os.ReadFile(destPath)
		if err != nil {
			t.Fatalf("failed to read file from storage: %v", err)
		}
		if string(content) != string(testContent) {
			t.Errorf("content mismatch: got %q, want %q", content, testContent)
		}
	})

	t.Run("unlink_file", func(t *testing.T) {
		entryName := "testapp"
		relPath := "config.json"
		destPath := filepath.Join(tmpStorage, "dotsync", entryName, relPath)

		status, _, err := symlink.Check(testFile, destPath)
		if err != nil {
			t.Fatalf("failed to check symlink: %v", err)
		}
		if status != symlink.StatusLinked {
			t.Errorf("expected linked status before unlink, got %v", status)
		}

		if err := symlink.Remove(testFile); err != nil {
			t.Fatalf("failed to remove symlink: %v", err)
		}

		if err := symlink.CopyFile(destPath, testFile); err != nil {
			t.Fatalf("failed to copy file back: %v", err)
		}

		info, err := os.Lstat(testFile)
		if err != nil {
			t.Fatalf("file doesn't exist after unlink: %v", err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			t.Error("file is still a symlink after unlink")
		}

		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if string(content) != string(testContent) {
			t.Errorf("content mismatch after unlink: got %q, want %q", content, testContent)
		}
	})

	t.Run("link_file_again", func(t *testing.T) {
		entryName := "testapp"
		relPath := "config.json"
		destPath := filepath.Join(tmpStorage, "dotsync", entryName, relPath)

		if err := os.Remove(testFile); err != nil {
			t.Fatalf("failed to remove file: %v", err)
		}

		if err := symlink.Create(testFile, destPath); err != nil {
			t.Fatalf("failed to create symlink: %v", err)
		}

		status, _, err := symlink.Check(testFile, destPath)
		if err != nil {
			t.Fatalf("failed to check symlink: %v", err)
		}
		if status != symlink.StatusLinked {
			t.Errorf("status = %v, want %v", status, symlink.StatusLinked)
		}

		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("failed to read through symlink: %v", err)
		}
		if string(content) != string(testContent) {
			t.Errorf("content mismatch: got %q, want %q", content, testContent)
		}
	})
}

func TestAddFromReadOnlyDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping read-only directory test on Windows")
	}

	tmpHome := t.TempDir()
	tmpStorage := t.TempDir()

	readonlyDir := filepath.Join(tmpHome, "readonly")
	if err := os.MkdirAll(readonlyDir, 0755); err != nil {
		t.Fatalf("failed to create readonly dir: %v", err)
	}

	testFile := filepath.Join(readonlyDir, "config.txt")
	testContent := []byte("test content")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if err := os.Chmod(readonlyDir, 0555); err != nil {
		t.Fatalf("failed to chmod directory: %v", err)
	}
	defer os.Chmod(readonlyDir, 0755)

	destPath := filepath.Join(tmpStorage, "dotsync", "testapp", "config.txt")

	err := symlink.MoveFile(testFile, destPath)

	originalExists := false
	if _, err := os.Stat(testFile); err == nil {
		originalExists = true
	}

	storageExists := false
	if _, err := os.Stat(destPath); err == nil {
		storageExists = true
	}

	if originalExists && storageExists {
		t.Errorf("file exists in both locations - operation was not atomic")
	}

	if !originalExists && storageExists {
		t.Log("correct: file moved to storage")
	}

	if originalExists && !storageExists {
		if err == nil {
			t.Error("operation succeeded but file only in original - inconsistent state")
		} else {
			t.Log("correct: file remains in original location after failed operation")
		}
	}
}
