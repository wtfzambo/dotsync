package pathutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/wtfzambo/dotsync/internal/manifest"
)

// TestValidateForAdd_FileNotFound tests validation when file doesn't exist
func TestValidateForAdd_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistent := filepath.Join(tmpDir, "nonexistent.txt")

	err := ValidateForAdd(nonExistent)
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}

	verr, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if verr.IsWarn {
		t.Error("expected fatal error, got warning")
	}
}

// TestValidateForAdd_Symlink tests validation rejects symlinks
func TestValidateForAdd_Symlink(t *testing.T) {
	tmpDir := t.TempDir()
	realFile := filepath.Join(tmpDir, "real.txt")
	symlinkFile := filepath.Join(tmpDir, "link.txt")

	// Create a real file
	if err := os.WriteFile(realFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create a symlink
	if err := os.Symlink(realFile, symlinkFile); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	err := ValidateForAdd(symlinkFile)
	if err == nil {
		t.Fatal("expected error for symlink")
	}

	verr, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if verr.IsWarn {
		t.Error("expected fatal error, got warning")
	}
}

// TestValidateForAdd_Directory tests validation rejects directories
func TestValidateForAdd_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "testdir")

	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	err := ValidateForAdd(testDir)
	if err == nil {
		t.Fatal("expected error for directory")
	}

	verr, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if verr.IsWarn {
		t.Error("expected fatal error, got warning")
	}
}

// TestValidateForAdd_PlistFile tests macOS plist rejection
func TestValidateForAdd_PlistFile(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping macOS-specific test")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	// We can't create files in real Library/Preferences, so we test with a temp dir
	tmpDir := t.TempDir()
	prefsDir := filepath.Join(tmpDir, "Library", "Preferences")
	if err := os.MkdirAll(prefsDir, 0755); err != nil {
		t.Fatalf("failed to create preferences dir: %v", err)
	}

	plistFile := filepath.Join(prefsDir, "com.example.app.plist")
	if err := os.WriteFile(plistFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create plist file: %v", err)
	}

	// Note: This test only works if tmpDir is under home
	// For a real test, we'd need to temporarily modify the home detection
	// So we'll just verify the isPlistFile logic works
	if !filepath.HasPrefix(plistFile, home) {
		t.Skip("temp dir not under home, skipping")
	}

	err = ValidateForAdd(plistFile)
	if err == nil {
		t.Fatal("expected error for plist file")
	}
}

// TestValidateForAdd_OutsideHome tests warning for files outside home
func TestValidateForAdd_OutsideHome(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "config.txt")

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err := ValidateForAdd(testFile)
	if err == nil {
		// If tmpDir happens to be under home, skip this test
		home, _ := os.UserHomeDir()
		if filepath.HasPrefix(tmpDir, home) {
			t.Skip("temp dir is under home, skipping")
		}
		t.Fatal("expected warning for file outside home")
	}

	verr, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if !verr.IsWarn {
		t.Error("expected warning, got fatal error")
	}
}

// TestValidateForAdd_ValidFile tests validation passes for valid files
func TestValidateForAdd_ValidFile(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tmpDir := filepath.Join(home, ".cache", "dotsync-test")
	defer os.RemoveAll(tmpDir)

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	testFile := filepath.Join(tmpDir, "valid.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err = ValidateForAdd(testFile)
	if err != nil {
		t.Errorf("expected no error for valid file, got: %v", err)
	}
}

// TestCheckEntryConflict tests entry conflict detection
func TestCheckEntryConflict(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	m := manifest.New()
	m.AddFile("opencode", filepath.Join("~", ".config", "opencode"), "config.json")
	m.AddFile("vscode", filepath.Join("~", ".config", "Code"), "settings.json")

	tests := []struct {
		name         string
		path         string
		explicitName string
		wantConflict string
		descr        string
	}{
		{
			name:         "file under existing root",
			path:         filepath.Join(home, ".config", "opencode", "another.json"),
			explicitName: "",
			wantConflict: "",
			descr:        "should not conflict - will be added to opencode entry",
		},
		{
			name:         "file under existing root with different name",
			path:         filepath.Join(home, ".config", "opencode", "another.json"),
			explicitName: "different",
			wantConflict: "opencode",
			descr:        "should conflict - explicit name differs from existing",
		},
		{
			name:         "non-conflicting file",
			path:         filepath.Join(home, ".config", "cursor", "settings.json"),
			explicitName: "",
			wantConflict: "",
			descr:        "should not conflict - new entry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckEntryConflict(tt.path, tt.explicitName, m)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantConflict {
				t.Errorf("CheckEntryConflict() = %q, want %q: %s", got, tt.wantConflict, tt.descr)
			}
		})
	}
}

// TestIsAlreadyTracked tests tracking detection
func TestIsAlreadyTracked(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	m := manifest.New()
	m.AddFile("opencode", filepath.Join("~", ".config", "opencode"), "config.json")
	m.AddFile("opencode", filepath.Join("~", ".config", "opencode"), filepath.Join("agents", "review.md"))
	m.AddFile("zsh", "~", ".zshrc")

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "tracked file in .config",
			path: filepath.Join(home, ".config", "opencode", "config.json"),
			want: "opencode",
		},
		{
			name: "tracked nested file",
			path: filepath.Join(home, ".config", "opencode", "agents", "review.md"),
			want: "opencode",
		},
		{
			name: "tracked dotfile",
			path: filepath.Join(home, ".zshrc"),
			want: "zsh",
		},
		{
			name: "not tracked",
			path: filepath.Join(home, ".config", "cursor", "settings.json"),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAlreadyTracked(tt.path, m)
			if got != tt.want {
				t.Errorf("IsAlreadyTracked() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestValidationError tests ValidationError type
func TestValidationError(t *testing.T) {
	tests := []struct {
		name   string
		err    ValidationError
		want   string
		isWarn bool
	}{
		{
			name: "fatal error",
			err: ValidationError{
				Path:    "/some/path",
				Message: "file not found",
				IsWarn:  false,
			},
			want:   "file not found",
			isWarn: false,
		},
		{
			name: "warning",
			err: ValidationError{
				Path:    "/some/path",
				Message: "file outside home",
				IsWarn:  true,
			},
			want:   "file outside home",
			isWarn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
			if tt.err.IsWarn != tt.isWarn {
				t.Errorf("IsWarn = %v, want %v", tt.err.IsWarn, tt.isWarn)
			}
		})
	}
}

// TestCheckEntryConflict_OverlappingRoots_DetectsConflict tests bug #4:
// Adding a file to an overlapping root should detect conflict.
//
// EXPECTED: Conflict detected - different inferred root vs existing root
// ACTUAL (bug): No conflict, file gets added to wrong entry
func TestCheckEntryConflict_OverlappingRoots_DetectsConflict(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	// 1. Create manifest with BROAD entry: root="~/.config"
	// This represents: user first added some file from ~/.config directory
	m := manifest.New()
	m.AddFile("config", filepath.Join("~", ".config"), "somefile")

	// Verify the entry was added correctly
	entry := m.GetEntry("config")
	if entry == nil {
		t.Fatal("failed to create 'config' entry")
	}
	t.Logf("Created entry 'config' with root: %s", entry.Root)

	// 2. Try to add file: "~/.config/app/file.json" (more specific root)
	// This should infer entry "app" with root "~/.config/app"
	testPath := filepath.Join(home, ".config", "app", "file.json")
	t.Logf("Testing file path: %s", testPath)

	// Create the file so InferFromPath can work with it
	if err := os.MkdirAll(filepath.Dir(testPath), 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	if err := os.WriteFile(testPath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Check what entry name would be inferred
	inferred := InferFromPath(testPath)
	if inferred == nil {
		t.Fatal("failed to infer entry from path")
	}
	t.Logf("Inferred entry name: %s, root: %s", inferred.Name, inferred.Root)

	// The inferred entry should be "app" with root "~/.config/app"
	if inferred.Name != "app" {
		t.Errorf("inferred name = %q, want %q", inferred.Name, "app")
	}

	// 3. EXPECTED: Conflict detected - inferred root (~/.config/app) is MORE SPECIFIC than existing root (~/.config)
	// The file's parent directory (~/.config/app) is a SUBDIRECTORY of the existing entry's root (~/.config)
	// This creates overlapping roots which should require explicit --name

	conflict, err := CheckEntryConflict(testPath, "", m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("CheckEntryConflict returned: %q", conflict)

	// The bug: No conflict is detected, so file would be added to "config" entry
	// even though it naturally belongs to a more specific entry "app"
	if conflict == "" {
		t.Errorf("BUG DETECTED: No conflict detected for overlapping roots")
		t.Errorf("  File path: %s", testPath)
		t.Errorf("  Inferred root: %s (for entry 'app')", inferred.Root)
		t.Errorf("  Existing entry 'config' has root: %s", entry.Root)
		t.Errorf("  EXPECTED: Should detect conflict - inferred root is a SUBDIRECTORY of existing root")
		t.Errorf("  ACTUAL: No conflict returned, file would be added to wrong entry")
		t.Logf("CONSEQUENCE: File would be absorbed into broad 'config' entry instead of 'app'")
	} else {
		t.Logf("Conflict correctly detected with entry: %s", conflict)
	}
}
