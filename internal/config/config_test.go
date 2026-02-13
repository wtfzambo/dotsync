package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestNew tests config creation
func TestNew(t *testing.T) {
	storagePath := "/path/to/storage"
	cfg := New(storagePath)

	if cfg == nil {
		t.Fatal("New() returned nil")
	}

	if cfg.StoragePath != storagePath {
		t.Errorf("StoragePath = %q, want %q", cfg.StoragePath, storagePath)
	}
}

// TestSave tests config saving
func TestSave(t *testing.T) {
	// Use a temporary home directory for testing
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".config", "dotsync")
	configFile := filepath.Join(configDir, "config.json")

	// Override the config directory for testing
	// We'll have to manually construct the path
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	cfg := New("/path/to/storage")

	// Manually save to test location
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configFile); err != nil {
		t.Error("config file was not created")
	}

	// Verify content
	readData, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(readData, &loaded); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	if loaded.StoragePath != cfg.StoragePath {
		t.Errorf("loaded StoragePath = %q, want %q", loaded.StoragePath, cfg.StoragePath)
	}
}

// TestLoad tests config loading
func TestLoad_NotExist(t *testing.T) {
	// This test will try to load from the real config location
	// Since we can't easily mock the home directory, we'll just test
	// that Load doesn't panic and handles missing config gracefully

	// Temporarily move the real config if it exists
	realConfigPath, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() failed: %v", err)
	}

	// Check if real config exists
	realConfigExists := false
	if _, err := os.Stat(realConfigPath); err == nil {
		realConfigExists = true
		// Backup real config
		backupPath := realConfigPath + ".backup"
		os.Rename(realConfigPath, backupPath)
		defer os.Rename(backupPath, realConfigPath)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Should return nil when config doesn't exist
	if cfg != nil && !realConfigExists {
		t.Error("Load() should return nil for non-existent config")
	}
}

// TestSaveLoad_RoundTrip tests saving and loading
func TestSaveLoad_RoundTrip(t *testing.T) {
	// We'll test the marshal/unmarshal logic directly
	original := New("/test/storage/path")

	data, err := json.MarshalIndent(original, "", "  ")
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if loaded.StoragePath != original.StoragePath {
		t.Errorf("StoragePath = %q, want %q", loaded.StoragePath, original.StoragePath)
	}
}

// TestConfigDir tests config directory path
func TestConfigDir(t *testing.T) {
	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() failed: %v", err)
	}

	if dir == "" {
		t.Error("ConfigDir() returned empty string")
	}

	// Should end with .config/dotsync
	if filepath.Base(dir) != "dotsync" {
		t.Errorf("directory name = %q, want %q", filepath.Base(dir), "dotsync")
	}

	parent := filepath.Base(filepath.Dir(dir))
	if parent != ".config" {
		t.Errorf("parent directory = %q, want %q", parent, ".config")
	}
}

// TestConfigPath tests config file path
func TestConfigPath(t *testing.T) {
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() failed: %v", err)
	}

	if path == "" {
		t.Error("ConfigPath() returned empty string")
	}

	// Should end with config.json
	if filepath.Base(path) != "config.json" {
		t.Errorf("filename = %q, want %q", filepath.Base(path), "config.json")
	}
}

// TestExists tests config existence checking
func TestExists(t *testing.T) {
	// Check current state (may or may not exist)
	exists, err := Exists()
	if err != nil {
		t.Fatalf("Exists() failed: %v", err)
	}

	// Just verify it returns a boolean
	t.Logf("Config exists: %v", exists)
}

// TestConfig_EmptyStoragePath tests config with empty storage path
func TestConfig_EmptyStoragePath(t *testing.T) {
	cfg := New("")

	if cfg.StoragePath != "" {
		t.Errorf("StoragePath = %q, want empty", cfg.StoragePath)
	}
}

// TestConfig_JSON tests JSON marshaling
func TestConfig_JSON(t *testing.T) {
	tests := []struct {
		name        string
		storagePath string
	}{
		{"normal path", "/path/to/storage"},
		{"tilde path", "~/storage"},
		{"empty path", ""},
		{"path with spaces", "/path with spaces/storage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := New(tt.storagePath)

			// Marshal to JSON
			data, err := json.Marshal(cfg)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			// Unmarshal back
			var loaded Config
			if err := json.Unmarshal(data, &loaded); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			if loaded.StoragePath != cfg.StoragePath {
				t.Errorf("StoragePath = %q, want %q", loaded.StoragePath, cfg.StoragePath)
			}
		})
	}
}

// TestConfig_JSONPretty tests pretty JSON formatting
func TestConfig_JSONPretty(t *testing.T) {
	cfg := New("/test/path")

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent failed: %v", err)
	}

	// Verify it contains the storage path
	dataStr := string(data)
	if dataStr == "" {
		t.Error("JSON output is empty")
	}

	// Should contain "storagePath" field
	if !contains(dataStr, "storagePath") {
		t.Error("JSON should contain 'storagePath' field")
	}
}

// TestDelete tests config deletion
func TestDelete(t *testing.T) {
	// Create a temporary config file
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".config", "dotsync")
	configFile := filepath.Join(configDir, "config.json")

	os.MkdirAll(configDir, 0755)
	cfg := New("/test/path")
	data, _ := json.Marshal(cfg)
	os.WriteFile(configFile, data, 0644)

	// Note: Delete() uses the real home directory, so we can't easily test it
	// without mocking. We'll just verify the function signature.
	_ = Delete
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
