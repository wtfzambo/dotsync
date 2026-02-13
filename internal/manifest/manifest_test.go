package manifest

import (
	"testing"
)

// TestNew tests manifest creation
func TestNew(t *testing.T) {
	m := New()
	if m == nil {
		t.Fatal("New() returned nil")
	}
	if m.Version != CurrentVersion {
		t.Errorf("Version = %d, want %d", m.Version, CurrentVersion)
	}
	if m.Entries == nil {
		t.Error("Entries map is nil")
	}
	if len(m.Entries) != 0 {
		t.Errorf("expected empty entries, got %d", len(m.Entries))
	}
}

// TestAddFile_NewEntry tests adding a file to a new entry
func TestAddFile_NewEntry(t *testing.T) {
	m := New()
	added := m.AddFile("opencode", "~/.config/opencode", "config.json")

	if !added {
		t.Error("AddFile() returned false, expected true")
	}

	entry := m.GetEntry("opencode")
	if entry == nil {
		t.Fatal("entry not found")
	}

	if entry.Root != "~/.config/opencode" {
		t.Errorf("Root = %q, want %q", entry.Root, "~/.config/opencode")
	}

	if len(entry.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(entry.Files))
	}

	if entry.Files[0] != "config.json" {
		t.Errorf("Files[0] = %q, want %q", entry.Files[0], "config.json")
	}
}

// TestAddFile_ExistingEntry tests adding a file to an existing entry
func TestAddFile_ExistingEntry(t *testing.T) {
	m := New()
	m.AddFile("opencode", "~/.config/opencode", "config.json")
	added := m.AddFile("opencode", "~/.config/opencode", "agents/review.md")

	if !added {
		t.Error("AddFile() returned false, expected true")
	}

	entry := m.GetEntry("opencode")
	if entry == nil {
		t.Fatal("entry not found")
	}

	if len(entry.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(entry.Files))
	}

	expectedFiles := []string{"config.json", "agents/review.md"}
	for i, expected := range expectedFiles {
		if entry.Files[i] != expected {
			t.Errorf("Files[%d] = %q, want %q", i, entry.Files[i], expected)
		}
	}
}

// TestAddFile_Duplicate tests that duplicate files are not added
func TestAddFile_Duplicate(t *testing.T) {
	m := New()
	m.AddFile("opencode", "~/.config/opencode", "config.json")
	added := m.AddFile("opencode", "~/.config/opencode", "config.json")

	if added {
		t.Error("AddFile() returned true for duplicate, expected false")
	}

	entry := m.GetEntry("opencode")
	if entry == nil {
		t.Fatal("entry not found")
	}

	if len(entry.Files) != 1 {
		t.Errorf("expected 1 file after duplicate add, got %d", len(entry.Files))
	}
}

// TestAddFile_MultipleEntries tests adding files to multiple entries
func TestAddFile_MultipleEntries(t *testing.T) {
	m := New()
	m.AddFile("opencode", "~/.config/opencode", "config.json")
	m.AddFile("vscode", "~/.config/Code", "settings.json")
	m.AddFile("zsh", "~", ".zshrc")

	if len(m.Entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(m.Entries))
	}

	tests := []struct {
		name     string
		wantRoot string
		wantFile string
	}{
		{"opencode", "~/.config/opencode", "config.json"},
		{"vscode", "~/.config/Code", "settings.json"},
		{"zsh", "~", ".zshrc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := m.GetEntry(tt.name)
			if entry == nil {
				t.Fatalf("entry %q not found", tt.name)
			}
			if entry.Root != tt.wantRoot {
				t.Errorf("Root = %q, want %q", entry.Root, tt.wantRoot)
			}
			if len(entry.Files) != 1 {
				t.Fatalf("expected 1 file, got %d", len(entry.Files))
			}
			if entry.Files[0] != tt.wantFile {
				t.Errorf("Files[0] = %q, want %q", entry.Files[0], tt.wantFile)
			}
		})
	}
}

// TestHasEntry tests entry existence checking
func TestHasEntry(t *testing.T) {
	m := New()
	m.AddFile("opencode", "~/.config/opencode", "config.json")

	if !m.HasEntry("opencode") {
		t.Error("HasEntry(opencode) returned false, expected true")
	}

	if m.HasEntry("nonexistent") {
		t.Error("HasEntry(nonexistent) returned true, expected false")
	}
}

// TestGetEntry tests entry retrieval
func TestGetEntry(t *testing.T) {
	m := New()
	m.AddFile("opencode", "~/.config/opencode", "config.json")

	entry := m.GetEntry("opencode")
	if entry == nil {
		t.Fatal("GetEntry(opencode) returned nil")
	}

	if entry.Root != "~/.config/opencode" {
		t.Errorf("Root = %q, want %q", entry.Root, "~/.config/opencode")
	}

	nonExistent := m.GetEntry("nonexistent")
	if nonExistent != nil {
		t.Error("GetEntry(nonexistent) returned non-nil")
	}
}

// TestGetEntry_ModifyReturned tests that modifying returned entry doesn't affect original
func TestGetEntry_ModifyReturned(t *testing.T) {
	m := New()
	m.AddFile("opencode", "~/.config/opencode", "config.json")

	entry := m.GetEntry("opencode")
	if entry == nil {
		t.Fatal("GetEntry returned nil")
	}

	// Try to modify the returned entry
	entry.Root = "different"
	entry.Files = []string{"modified"}

	// Check that the original is unchanged
	original := m.GetEntry("opencode")
	if original.Root == "different" {
		t.Error("modifying returned entry affected the original")
	}
	if len(original.Files) != 1 || original.Files[0] != "config.json" {
		t.Error("modifying returned entry files affected the original")
	}
}

// TestCurrentVersion tests that version constant is set
func TestCurrentVersion(t *testing.T) {
	if CurrentVersion <= 0 {
		t.Errorf("CurrentVersion = %d, expected > 0", CurrentVersion)
	}
}

// TestManifest_EmptyEntries tests manifest with no entries
func TestManifest_EmptyEntries(t *testing.T) {
	m := New()

	if m.GetEntry("anything") != nil {
		t.Error("GetEntry on empty manifest returned non-nil")
	}

	if m.HasEntry("anything") {
		t.Error("HasEntry on empty manifest returned true")
	}
}
