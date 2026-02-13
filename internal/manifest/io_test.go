package manifest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestSave tests manifest saving
func TestSave(t *testing.T) {
	tmpDir := t.TempDir()

	m := New()
	m.AddFile("opencode", "~/.config/opencode", "config.json")
	m.AddFile("opencode", "~/.config/opencode", "agents/review.md")
	m.AddFile("vscode", "~/.config/Code", "settings.json")

	err := m.Save(tmpDir)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Check that the file was created
	manifestPath := filepath.Join(tmpDir, "dotsync", ManifestFileName)
	if _, err := os.Stat(manifestPath); err != nil {
		t.Fatalf("manifest file not created: %v", err)
	}

	// Read and verify the content
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to read manifest: %v", err)
	}

	var loaded Manifest
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to parse manifest: %v", err)
	}

	if loaded.Version != CurrentVersion {
		t.Errorf("Version = %d, want %d", loaded.Version, CurrentVersion)
	}

	if len(loaded.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(loaded.Entries))
	}

	// Verify opencode entry
	opencode, ok := loaded.Entries["opencode"]
	if !ok {
		t.Fatal("opencode entry not found")
	}
	if opencode.Root != "~/.config/opencode" {
		t.Errorf("opencode Root = %q, want %q", opencode.Root, "~/.config/opencode")
	}
	if len(opencode.Files) != 2 {
		t.Errorf("opencode expected 2 files, got %d", len(opencode.Files))
	}
}

// TestLoad tests manifest loading
func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a manifest file manually
	dotsyncDir := filepath.Join(tmpDir, "dotsync")
	if err := os.MkdirAll(dotsyncDir, 0755); err != nil {
		t.Fatalf("failed to create dotsync dir: %v", err)
	}

	manifestData := `{
  "version": 1,
  "entries": {
    "opencode": {
      "root": "~/.config/opencode",
      "files": ["config.json", "agents/review.md"]
    },
    "vscode": {
      "root": "~/.config/Code",
      "files": ["settings.json"]
    }
  }
}`

	manifestPath := filepath.Join(dotsyncDir, ManifestFileName)
	if err := os.WriteFile(manifestPath, []byte(manifestData), 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	// Load the manifest
	m, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if m.Version != 1 {
		t.Errorf("Version = %d, want 1", m.Version)
	}

	if len(m.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(m.Entries))
	}

	// Verify opencode entry
	opencode := m.GetEntry("opencode")
	if opencode == nil {
		t.Fatal("opencode entry not found")
	}
	if opencode.Root != "~/.config/opencode" {
		t.Errorf("opencode Root = %q, want %q", opencode.Root, "~/.config/opencode")
	}
	if len(opencode.Files) != 2 {
		t.Errorf("opencode expected 2 files, got %d", len(opencode.Files))
	}

	// Verify vscode entry
	vscode := m.GetEntry("vscode")
	if vscode == nil {
		t.Fatal("vscode entry not found")
	}
	if vscode.Root != "~/.config/Code" {
		t.Errorf("vscode Root = %q, want %q", vscode.Root, "~/.config/Code")
	}
	if len(vscode.Files) != 1 {
		t.Errorf("vscode expected 1 file, got %d", len(vscode.Files))
	}
}

// TestLoad_NotFound tests loading when manifest doesn't exist
func TestLoad_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := Load(tmpDir)
	if err == nil {
		t.Fatal("Load() should fail when manifest doesn't exist")
	}

	// Check error message contains "manifest not found"
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

// TestLoad_VersionTooNew tests loading a manifest with a newer version
func TestLoad_VersionTooNew(t *testing.T) {
	tmpDir := t.TempDir()

	dotsyncDir := filepath.Join(tmpDir, "dotsync")
	if err := os.MkdirAll(dotsyncDir, 0755); err != nil {
		t.Fatalf("failed to create dotsync dir: %v", err)
	}

	// Create manifest with version 99
	manifestData := `{
  "version": 99,
  "entries": {}
}`

	manifestPath := filepath.Join(dotsyncDir, ManifestFileName)
	if err := os.WriteFile(manifestPath, []byte(manifestData), 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	_, err := Load(tmpDir)
	if err == nil {
		t.Fatal("Load() should fail for newer version")
	}

	// Check that it's an ErrVersionTooNew
	_, ok := err.(ErrVersionTooNew)
	if !ok {
		t.Errorf("expected ErrVersionTooNew, got %T", err)
	}
}

// TestLoad_EmptyManifest tests loading an empty manifest
func TestLoad_EmptyManifest(t *testing.T) {
	tmpDir := t.TempDir()

	dotsyncDir := filepath.Join(tmpDir, "dotsync")
	if err := os.MkdirAll(dotsyncDir, 0755); err != nil {
		t.Fatalf("failed to create dotsync dir: %v", err)
	}

	manifestData := `{
  "version": 1,
  "entries": {}
}`

	manifestPath := filepath.Join(dotsyncDir, ManifestFileName)
	if err := os.WriteFile(manifestPath, []byte(manifestData), 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	m, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if m.Version != 1 {
		t.Errorf("Version = %d, want 1", m.Version)
	}

	if len(m.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(m.Entries))
	}

	if m.Entries == nil {
		t.Error("Entries map should be initialized, not nil")
	}
}

// TestLoad_InvalidJSON tests loading invalid JSON
func TestLoad_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	dotsyncDir := filepath.Join(tmpDir, "dotsync")
	if err := os.MkdirAll(dotsyncDir, 0755); err != nil {
		t.Fatalf("failed to create dotsync dir: %v", err)
	}

	manifestPath := filepath.Join(dotsyncDir, ManifestFileName)
	if err := os.WriteFile(manifestPath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	_, err := Load(tmpDir)
	if err == nil {
		t.Fatal("Load() should fail for invalid JSON")
	}
}

// TestSaveLoad_RoundTrip tests saving and loading preserves data
func TestSaveLoad_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()

	// Create and save a manifest
	original := New()
	original.AddFile("opencode", "~/.config/opencode", "config.json")
	original.AddFile("opencode", "~/.config/opencode", "agents/review.md")
	original.AddFile("vscode", "~/.config/Code", "settings.json")
	original.AddFile("zsh", "~", ".zshrc")

	if err := original.Save(tmpDir); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Load it back
	loaded, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Compare
	if loaded.Version != original.Version {
		t.Errorf("Version = %d, want %d", loaded.Version, original.Version)
	}

	if len(loaded.Entries) != len(original.Entries) {
		t.Fatalf("expected %d entries, got %d", len(original.Entries), len(loaded.Entries))
	}

	for name, origEntry := range original.Entries {
		loadedEntry, ok := loaded.Entries[name]
		if !ok {
			t.Errorf("entry %q not found in loaded manifest", name)
			continue
		}

		if loadedEntry.Root != origEntry.Root {
			t.Errorf("entry %q: Root = %q, want %q", name, loadedEntry.Root, origEntry.Root)
		}

		if len(loadedEntry.Files) != len(origEntry.Files) {
			t.Errorf("entry %q: expected %d files, got %d", name, len(origEntry.Files), len(loadedEntry.Files))
			continue
		}

		for i, origFile := range origEntry.Files {
			if loadedEntry.Files[i] != origFile {
				t.Errorf("entry %q: Files[%d] = %q, want %q", name, i, loadedEntry.Files[i], origFile)
			}
		}
	}
}

// TestExists tests manifest existence checking
func TestExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Should not exist initially
	if Exists(tmpDir) {
		t.Error("Exists() returned true for non-existent manifest")
	}

	// Create a manifest
	m := New()
	if err := m.Save(tmpDir); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Should exist now
	if !Exists(tmpDir) {
		t.Error("Exists() returned false for existing manifest")
	}
}

// TestManifestPath tests manifest path generation
func TestManifestPath(t *testing.T) {
	storagePath := "/path/to/storage"
	expected := filepath.Join(storagePath, "dotsync", ManifestFileName)

	got := ManifestPath(storagePath)
	if got != expected {
		t.Errorf("ManifestPath() = %q, want %q", got, expected)
	}
}

// TestErrVersionTooNew tests version error type
func TestErrVersionTooNew(t *testing.T) {
	err := ErrVersionTooNew{Version: 5}
	expected := "manifest version 5 not supported. Please upgrade dotsync"

	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}
