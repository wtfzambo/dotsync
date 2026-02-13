package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const ManifestFileName = ".dotsync.json"

// ErrVersionTooNew is returned when the manifest version is newer than supported.
type ErrVersionTooNew struct {
	Version int
}

func (e ErrVersionTooNew) Error() string {
	return fmt.Sprintf("manifest version %d not supported. Please upgrade dotsync", e.Version)
}

// Load reads a manifest from the given dotsync storage directory.
// Returns ErrVersionTooNew if the manifest version is not supported.
func Load(storagePath string) (*Manifest, error) {
	manifestPath := filepath.Join(storagePath, "dotsync", ManifestFileName)

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("manifest not found at %s", manifestPath)
		}
		return nil, fmt.Errorf("reading manifest: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}

	// Version check
	if m.Version > CurrentVersion {
		return nil, ErrVersionTooNew{Version: m.Version}
	}

	// Ensure entries map is initialized
	if m.Entries == nil {
		m.Entries = make(map[string]Entry)
	}

	return &m, nil
}

// Save writes the manifest to the given dotsync storage directory.
func (m *Manifest) Save(storagePath string) error {
	dotsyncDir := filepath.Join(storagePath, "dotsync")
	manifestPath := filepath.Join(dotsyncDir, ManifestFileName)

	// Ensure the dotsync directory exists
	if err := os.MkdirAll(dotsyncDir, 0755); err != nil {
		return fmt.Errorf("creating dotsync directory: %w", err)
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding manifest: %w", err)
	}

	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}

	return nil
}

// ManifestPath returns the full path to the manifest file.
func ManifestPath(storagePath string) string {
	return filepath.Join(storagePath, "dotsync", ManifestFileName)
}

// Exists checks if a manifest exists at the given storage path.
func Exists(storagePath string) bool {
	_, err := os.Stat(ManifestPath(storagePath))
	return err == nil
}
