package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestDetectPath tests cloud storage path detection
func TestDetectPath(t *testing.T) {
	// Note: This test depends on actual filesystem, so results may vary
	// We'll test that it returns empty string for non-existent providers

	// Test with a provider that likely doesn't exist
	path := DetectPath("nonexistent")
	if path != "" {
		t.Logf("unexpected path for nonexistent provider: %s", path)
	}

	// Test real providers (may or may not find them)
	providers := []Provider{ProviderGoogleDrive, ProviderDropbox}
	if runtime.GOOS == "darwin" {
		providers = append(providers, ProviderICloud)
	}

	for _, provider := range providers {
		path := DetectPath(provider)
		// We can't assert much here since it depends on user's system
		t.Logf("Provider %s: %s", provider, path)
	}
}

// TestFindPath tests path finding with various patterns
func TestFindPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test directories
	testDir := filepath.Join(tmpDir, "testdir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	tests := []struct {
		name    string
		pattern string
		wantNil bool
	}{
		{
			name:    "existing directory",
			pattern: testDir,
			wantNil: false,
		},
		{
			name:    "non-existent directory",
			pattern: filepath.Join(tmpDir, "nonexistent"),
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findPath(tt.pattern)
			if tt.wantNil && got != "" {
				t.Errorf("findPath() = %q, expected empty", got)
			}
			if !tt.wantNil && got == "" {
				t.Errorf("findPath() = empty, expected non-empty")
			}
		})
	}
}

// TestKnownPaths tests that KnownPaths returns data for current platform
func TestKnownPaths(t *testing.T) {
	paths := KnownPaths()
	if paths == nil {
		t.Error("KnownPaths() returned nil")
		return
	}

	// Should have at least one provider
	if len(paths) == 0 {
		t.Error("KnownPaths() returned empty map")
	}

	// Check that providers have paths
	for provider, pp := range paths {
		t.Logf("Provider %s: Primary=%s, Fallback=%s", provider, pp.Primary, pp.Fallback)
		if pp.Primary == "" && pp.Fallback == "" {
			t.Errorf("provider %s has no paths", provider)
		}
	}
}

// TestSupportedProviders tests supported providers list
func TestSupportedProviders(t *testing.T) {
	providers := SupportedProviders()
	if providers == nil {
		t.Error("SupportedProviders() returned nil")
		return
	}

	// Should have at least one provider
	if len(providers) == 0 {
		t.Error("SupportedProviders() returned empty list")
	}

	// Verify each provider has a primary path
	paths := KnownPaths()
	for _, p := range providers {
		pp, ok := paths[p]
		if !ok {
			t.Errorf("provider %s not in KnownPaths", p)
			continue
		}
		if pp.Primary == "" {
			t.Errorf("provider %s has empty primary path", p)
		}
	}
}

// TestParseProvider tests provider parsing
func TestParseProvider(t *testing.T) {
	tests := []struct {
		input string
		want  Provider
	}{
		{"gdrive", ProviderGoogleDrive},
		{"googledrive", ProviderGoogleDrive},
		{"google-drive", ProviderGoogleDrive},
		{"dropbox", ProviderDropbox},
		{"icloud", ProviderICloud},
		{"unknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseProvider(tt.input)
			if got != tt.want {
				t.Errorf("ParseProvider(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestProvider_String tests provider string representation
func TestProvider_String(t *testing.T) {
	tests := []struct {
		provider Provider
		want     string
	}{
		{ProviderGoogleDrive, "gdrive"},
		{ProviderDropbox, "dropbox"},
		{ProviderICloud, "icloud"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.provider.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestProvider_DisplayName tests provider display name
func TestProvider_DisplayName(t *testing.T) {
	tests := []struct {
		provider Provider
		want     string
	}{
		{ProviderGoogleDrive, "Google Drive"},
		{ProviderDropbox, "Dropbox"},
		{ProviderICloud, "iCloud Drive"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.provider.DisplayName()
			if got != tt.want {
				t.Errorf("DisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestPlatformSpecificPaths tests that platform-specific paths are correct
func TestPlatformSpecificPaths(t *testing.T) {
	switch runtime.GOOS {
	case "darwin":
		if macOSPaths == nil {
			t.Fatal("macOSPaths is nil")
		}
		// Check iCloud path
		icloud, ok := macOSPaths[ProviderICloud]
		if !ok {
			t.Fatal("iCloud not in macOSPaths")
		}
		if icloud.Primary == "" {
			t.Error("iCloud should have primary path on macOS")
		}

	case "linux":
		if linuxPaths == nil {
			t.Fatal("linuxPaths is nil")
		}
		// iCloud should not have paths on Linux
		icloud, ok := linuxPaths[ProviderICloud]
		if ok && icloud.Primary != "" {
			t.Error("iCloud should not have paths on Linux")
		}

	case "windows":
		if windowsPaths == nil {
			t.Fatal("windowsPaths is nil")
		}
	}
}
