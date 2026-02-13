package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wtfzambo/dotsync/internal/backup"
	"github.com/wtfzambo/dotsync/internal/config"
	"github.com/wtfzambo/dotsync/internal/manifest"
	"github.com/wtfzambo/dotsync/internal/pathutil"
	"github.com/wtfzambo/dotsync/internal/symlink"
)

// TestAddLinkUnlinkCycle tests the full workflow: add file → link → unlink
func TestAddLinkUnlinkCycle(t *testing.T) {
	// Setup: create temporary directories
	tmpHome := t.TempDir()
	tmpStorage := t.TempDir()

	// Create a test file in "home"
	configDir := filepath.Join(tmpHome, ".config", "testapp")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	testFile := filepath.Join(configDir, "config.json")
	testContent := []byte(`{"test": "value"}`)
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create manifest
	m := manifest.New()
	dotsyncDir := filepath.Join(tmpStorage, "dotsync")
	if err := os.MkdirAll(dotsyncDir, 0755); err != nil {
		t.Fatalf("failed to create dotsync dir: %v", err)
	}

	// PHASE 1: Add file (simulate add command)
	t.Run("add_file", func(t *testing.T) {
		// Infer entry details
		entryName := "testapp"
		relPath := "config.json"

		// Destination in storage
		destPath := filepath.Join(tmpStorage, "dotsync", entryName, relPath)

		// Move file to storage
		if err := symlink.MoveFile(testFile, destPath); err != nil {
			t.Fatalf("failed to move file: %v", err)
		}

		// Create symlink
		if err := symlink.Create(testFile, destPath); err != nil {
			t.Fatalf("failed to create symlink: %v", err)
		}

		// Update manifest
		m.AddFile(entryName, filepath.Join("~", ".config", "testapp"), relPath)
		if err := m.Save(tmpStorage); err != nil {
			t.Fatalf("failed to save manifest: %v", err)
		}

		// Verify symlink exists and points correctly
		status, target, err := symlink.Check(testFile, destPath)
		if err != nil {
			t.Fatalf("failed to check symlink: %v", err)
		}
		if status != symlink.StatusLinked {
			t.Errorf("status = %v, want %v (target: %s)", status, symlink.StatusLinked, target)
		}

		// Verify file exists in storage
		if _, err := os.Stat(destPath); err != nil {
			t.Errorf("file not found in storage: %v", err)
		}

		// Verify content is preserved
		content, err := os.ReadFile(destPath)
		if err != nil {
			t.Fatalf("failed to read file from storage: %v", err)
		}
		if string(content) != string(testContent) {
			t.Errorf("content mismatch: got %q, want %q", content, testContent)
		}
	})

	// PHASE 2: Unlink (simulate unlink command)
	t.Run("unlink_file", func(t *testing.T) {
		entryName := "testapp"
		relPath := "config.json"
		destPath := filepath.Join(tmpStorage, "dotsync", entryName, relPath)

		// Check current status
		status, _, err := symlink.Check(testFile, destPath)
		if err != nil {
			t.Fatalf("failed to check symlink: %v", err)
		}
		if status != symlink.StatusLinked {
			t.Errorf("expected linked status before unlink, got %v", status)
		}

		// Remove symlink
		if err := symlink.Remove(testFile); err != nil {
			t.Fatalf("failed to remove symlink: %v", err)
		}

		// Copy file back from storage
		if err := symlink.CopyFile(destPath, testFile); err != nil {
			t.Fatalf("failed to copy file back: %v", err)
		}

		// Verify file is regular file now
		info, err := os.Lstat(testFile)
		if err != nil {
			t.Fatalf("file doesn't exist after unlink: %v", err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			t.Error("file is still a symlink after unlink")
		}

		// Verify content
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if string(content) != string(testContent) {
			t.Errorf("content mismatch after unlink: got %q, want %q", content, testContent)
		}
	})

	// PHASE 3: Link again (simulate link command on new machine)
	t.Run("link_file_again", func(t *testing.T) {
		entryName := "testapp"
		relPath := "config.json"
		destPath := filepath.Join(tmpStorage, "dotsync", entryName, relPath)

		// Remove local file (simulate new machine)
		if err := os.Remove(testFile); err != nil {
			t.Fatalf("failed to remove file: %v", err)
		}

		// Create symlink (file is in storage)
		if err := symlink.Create(testFile, destPath); err != nil {
			t.Fatalf("failed to create symlink: %v", err)
		}

		// Verify symlink
		status, _, err := symlink.Check(testFile, destPath)
		if err != nil {
			t.Fatalf("failed to check symlink: %v", err)
		}
		if status != symlink.StatusLinked {
			t.Errorf("status = %v, want %v", status, symlink.StatusLinked)
		}

		// Verify can read through symlink
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("failed to read through symlink: %v", err)
		}
		if string(content) != string(testContent) {
			t.Errorf("content mismatch: got %q, want %q", content, testContent)
		}
	})
}

// TestConflictScenarios tests various conflict scenarios
func TestConflictScenarios(t *testing.T) {
	tmpHome := t.TempDir()
	tmpStorage := t.TempDir()

	// Create manifest with proper roots
	m := manifest.New()
	m.AddFile("app1", filepath.Join(tmpHome, ".config", "app1"), "config.json")
	m.AddFile("app2", filepath.Join(tmpHome, ".config", "app2"), "settings.json")

	t.Run("file_already_tracked", func(t *testing.T) {
		// Create the actual file so IsAlreadyTracked can find it
		absPath := filepath.Join(tmpHome, ".config", "app1", "config.json")
		os.MkdirAll(filepath.Dir(absPath), 0755)
		os.WriteFile(absPath, []byte("test"), 0644)

		tracked := pathutil.IsAlreadyTracked(absPath, m)
		if tracked != "app1" {
			t.Errorf("tracked by %q, want %q", tracked, "app1")
		}
	})

	t.Run("file_exists_at_target", func(t *testing.T) {
		// Create destination file
		destPath := filepath.Join(tmpStorage, "dotsync", "app3", "config.json")
		os.MkdirAll(filepath.Dir(destPath), 0755)
		os.WriteFile(destPath, []byte("existing"), 0644)

		// Check if destination exists
		if _, err := os.Stat(destPath); err != nil {
			t.Error("destination should exist")
		}
	})

	t.Run("entry_conflict", func(t *testing.T) {
		// Create files to test conflict
		absPath := filepath.Join(tmpHome, ".config", "app1", "other.json")
		os.MkdirAll(filepath.Dir(absPath), 0755)
		os.WriteFile(absPath, []byte("test"), 0644)

		conflict, err := pathutil.CheckEntryConflict(absPath, "app2", m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if conflict != "app1" {
			t.Errorf("conflict = %q, want %q (file is under app1 root)", conflict, "app1")
		}
	})
}

// TestBrokenSymlinkHandling tests handling of broken symlinks
func TestBrokenSymlinkHandling(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a broken symlink
	linkPath := filepath.Join(tmpDir, "broken.txt")
	targetPath := filepath.Join(tmpDir, "nonexistent.txt")

	if err := os.Symlink(targetPath, linkPath); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Check status
	status, actualTarget, err := symlink.Check(linkPath, targetPath)
	if err != nil {
		t.Fatalf("Check() failed: %v", err)
	}

	if status != symlink.StatusBroken {
		t.Errorf("status = %v, want %v", status, symlink.StatusBroken)
	}

	if actualTarget != targetPath {
		t.Errorf("actualTarget = %q, want %q", actualTarget, targetPath)
	}

	// Remove broken symlink
	if err := symlink.Remove(linkPath); err != nil {
		t.Errorf("failed to remove broken symlink: %v", err)
	}

	// Verify it's gone
	if _, err := os.Lstat(linkPath); !os.IsNotExist(err) {
		t.Error("broken symlink should be removed")
	}
}

// TestRollbackOnFailure tests rollback behavior when operations fail
func TestRollbackOnFailure(t *testing.T) {
	tmpDir := t.TempDir()
	originalFile := filepath.Join(tmpDir, "original.txt")
	content := []byte("important data")

	// Create original file
	if err := os.WriteFile(originalFile, content, 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	t.Run("backup_and_restore", func(t *testing.T) {
		// Create backup
		bk, err := backup.Create(originalFile)
		if err != nil {
			t.Fatalf("backup creation failed: %v", err)
		}

		// Simulate an operation that modifies the file
		if err := os.WriteFile(originalFile, []byte("modified"), 0644); err != nil {
			t.Fatalf("failed to modify file: %v", err)
		}

		// Simulate failure - restore backup
		if err := bk.Restore(); err != nil {
			t.Fatalf("restore failed: %v", err)
		}

		// Verify content is restored
		restoredContent, err := os.ReadFile(originalFile)
		if err != nil {
			t.Fatalf("failed to read restored file: %v", err)
		}

		if string(restoredContent) != string(content) {
			t.Errorf("content = %q, want %q", restoredContent, content)
		}
	})
}

// TestManifestUpdates tests manifest update scenarios
func TestManifestUpdates(t *testing.T) {
	tmpStorage := t.TempDir()

	m := manifest.New()

	t.Run("add_multiple_files_to_entry", func(t *testing.T) {
		m.AddFile("app", "~/.config/app", "config.json")
		m.AddFile("app", "~/.config/app", "settings.json")
		m.AddFile("app", "~/.config/app", "themes/dark.json")

		entry := m.GetEntry("app")
		if entry == nil {
			t.Fatal("entry not found")
		}

		if len(entry.Files) != 3 {
			t.Errorf("expected 3 files, got %d", len(entry.Files))
		}

		expected := []string{"config.json", "settings.json", "themes/dark.json"}
		for i, exp := range expected {
			if entry.Files[i] != exp {
				t.Errorf("file[%d] = %q, want %q", i, entry.Files[i], exp)
			}
		}
	})

	t.Run("save_and_load_manifest", func(t *testing.T) {
		// Save
		if err := m.Save(tmpStorage); err != nil {
			t.Fatalf("save failed: %v", err)
		}

		// Load
		loaded, err := manifest.Load(tmpStorage)
		if err != nil {
			t.Fatalf("load failed: %v", err)
		}

		// Compare
		if len(loaded.Entries) != len(m.Entries) {
			t.Errorf("loaded %d entries, want %d", len(loaded.Entries), len(m.Entries))
		}

		for name, origEntry := range m.Entries {
			loadedEntry := loaded.GetEntry(name)
			if loadedEntry == nil {
				t.Errorf("entry %q not loaded", name)
				continue
			}

			if loadedEntry.Root != origEntry.Root {
				t.Errorf("entry %q: root = %q, want %q", name, loadedEntry.Root, origEntry.Root)
			}

			if len(loadedEntry.Files) != len(origEntry.Files) {
				t.Errorf("entry %q: %d files, want %d", name, len(loadedEntry.Files), len(origEntry.Files))
			}
		}
	})
}

// TestConfigAndManifestInteraction tests config and manifest working together
func TestConfigAndManifestInteraction(t *testing.T) {
	tmpStorage := t.TempDir()

	// Create config
	cfg := config.New(tmpStorage)

	if cfg.StoragePath != tmpStorage {
		t.Errorf("StoragePath = %q, want %q", cfg.StoragePath, tmpStorage)
	}

	// Create manifest in storage
	m := manifest.New()
	m.AddFile("test", "~/.config/test", "config.json")

	if err := m.Save(tmpStorage); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	// Load manifest using config's storage path
	loaded, err := manifest.Load(cfg.StoragePath)
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}

	if len(loaded.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(loaded.Entries))
	}

	if !loaded.HasEntry("test") {
		t.Error("test entry not found")
	}
}
