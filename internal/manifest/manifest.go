// Package manifest handles reading and writing the dotsync manifest file.
// The manifest tracks all entries (apps/tools) and their associated files.
package manifest

// CurrentVersion is the current manifest schema version.
const CurrentVersion = 1

// Manifest represents the dotsync manifest file stored in cloud storage.
// Location: <cloud-folder>/dotsync/.dotsync.json
type Manifest struct {
	// Version is the schema version for forward compatibility
	Version int `json:"version"`

	// Entries maps entry names to their configuration
	// Key is the entry name (e.g., "opencode", "zsh", "cursor")
	Entries map[string]Entry `json:"entries"`
}

// Entry represents a tracked application/tool configuration.
type Entry struct {
	// Root is the original parent directory (uses ~ for home)
	// e.g., "~/.config/opencode" or "~"
	Root string `json:"root"`

	// Files are relative paths from Root
	// e.g., ["config.json", "agents/review.md"]
	Files []string `json:"files"`
}

// New creates a new empty manifest with the current version.
func New() *Manifest {
	return &Manifest{
		Version: CurrentVersion,
		Entries: make(map[string]Entry),
	}
}

// AddFile adds a file to an entry. Creates the entry if it doesn't exist.
// Returns true if the file was added, false if it was already tracked.
func (m *Manifest) AddFile(name, root, relPath string) bool {
	entry, exists := m.Entries[name]
	if !exists {
		entry = Entry{
			Root:  root,
			Files: []string{},
		}
	}

	// Check if file is already tracked
	for _, f := range entry.Files {
		if f == relPath {
			return false
		}
	}

	entry.Files = append(entry.Files, relPath)
	m.Entries[name] = entry
	return true
}

// HasEntry returns true if an entry with the given name exists.
func (m *Manifest) HasEntry(name string) bool {
	_, exists := m.Entries[name]
	return exists
}

// GetEntry returns an entry by name, or nil if not found.
func (m *Manifest) GetEntry(name string) *Entry {
	entry, exists := m.Entries[name]
	if !exists {
		return nil
	}
	return &entry
}

// IsFileTracked returns true if the given file is already tracked in any entry.
// Returns the entry name if found, empty string otherwise.
func (m *Manifest) IsFileTracked(absPath string) string {
	// This requires path expansion to check properly
	// Will be implemented with path utilities
	return ""
}
