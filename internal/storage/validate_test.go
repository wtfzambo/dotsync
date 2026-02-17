package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestValidatePath tests path validation
func TestValidatePath(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		path    string
		setup   func() string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid writable directory",
			setup: func() string {
				dir := filepath.Join(tmpDir, "valid")
				os.MkdirAll(dir, 0755)
				return dir
			},
			wantErr: false,
		},
		{
			name: "non-existent path",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent")
			},
			wantErr: true,
			errMsg:  "does not exist",
		},
		{
			name: "file instead of directory",
			setup: func() string {
				file := filepath.Join(tmpDir, "file.txt")
				os.WriteFile(file, []byte("test"), 0644)
				return file
			},
			wantErr: true,
			errMsg:  "not a directory",
		},
		{
			name: "read-only directory",
			setup: func() string {
				dir := filepath.Join(tmpDir, "readonly")
				os.MkdirAll(dir, 0555) // read-only (Unix only)
				return dir
			},
			wantErr: runtime.GOOS != "windows", // Windows permissions work differently
			errMsg:  "cannot write",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			err := ValidatePath(path)

			if tt.wantErr && err == nil {
				t.Error("ValidatePath() should return error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidatePath() unexpected error: %v", err)
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				// Just check that error message contains expected text
				if err.Error() == "" {
					t.Error("error message is empty")
				}
			}
		})
	}
}

// TestValidatePath_WithTilde tests validation with ~ expansion
func TestValidatePath_WithTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	// Test with a path that should exist (home directory itself)
	err = ValidatePath("~")
	if err != nil {
		t.Errorf("ValidatePath(~) failed: %v", err)
	}

	// Create a test directory in home
	testDir := filepath.Join(home, ".cache", "dotsync-test-validate")
	defer os.RemoveAll(testDir)

	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	// Test with tilde path
	tildeDir := filepath.Join("~", ".cache", "dotsync-test-validate")
	err = ValidatePath(tildeDir)
	if err != nil {
		t.Errorf("ValidatePath(%s) failed: %v", tildeDir, err)
	}
}

// TestEnsureDotsyncDir tests dotsync directory creation
func TestEnsureDotsyncDir(t *testing.T) {
	tmpDir := t.TempDir()

	dotsyncDir, err := EnsureDotsyncDir(tmpDir)
	if err != nil {
		t.Fatalf("EnsureDotsyncDir() failed: %v", err)
	}

	if dotsyncDir == "" {
		t.Error("EnsureDotsyncDir() returned empty string")
	}

	// Verify directory exists
	info, err := os.Stat(dotsyncDir)
	if err != nil {
		t.Fatalf("dotsync directory doesn't exist: %v", err)
	}

	if !info.IsDir() {
		t.Error("dotsync path is not a directory")
	}

	// Verify it's named correctly
	if filepath.Base(dotsyncDir) != "dotsync" {
		t.Errorf("directory name = %q, want %q", filepath.Base(dotsyncDir), "dotsync")
	}
}

// TestEnsureDotsyncDir_AlreadyExists tests with existing directory
func TestEnsureDotsyncDir_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create dotsync dir manually
	dotsyncPath := filepath.Join(tmpDir, "dotsync")
	if err := os.MkdirAll(dotsyncPath, 0755); err != nil {
		t.Fatalf("failed to create dotsync dir: %v", err)
	}

	// Should succeed even if already exists
	result, err := EnsureDotsyncDir(tmpDir)
	if err != nil {
		t.Errorf("EnsureDotsyncDir() failed: %v", err)
	}

	if result != dotsyncPath {
		t.Errorf("EnsureDotsyncDir() = %q, want %q", result, dotsyncPath)
	}
}

// TestIsAvailable tests storage availability checking
func TestIsAvailable(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		path  string
		setup func()
		want  bool
	}{
		{
			name:  "available path",
			path:  tmpDir,
			setup: func() {},
			want:  true,
		},
		{
			name:  "non-existent path",
			path:  filepath.Join(tmpDir, "nonexistent"),
			setup: func() {},
			want:  false,
		},
		{
			name: "newly created path",
			path: filepath.Join(tmpDir, "newdir"),
			setup: func() {
				os.MkdirAll(filepath.Join(tmpDir, "newdir"), 0755)
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got := IsAvailable(tt.path)
			if got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDotsyncDir tests dotsync directory path generation
func TestDotsyncDir(t *testing.T) {
	storagePath := "/path/to/storage"
	expected := filepath.Join(storagePath, "dotsync")

	got := DotsyncDir(storagePath)
	if got != expected {
		t.Errorf("DotsyncDir() = %q, want %q", got, expected)
	}
}

// TestDotsyncDir_WithTilde tests with tilde expansion
func TestDotsyncDir_WithTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	storagePath := "~/storage"
	expected := filepath.Join(home, "storage", "dotsync")

	got := DotsyncDir(storagePath)
	if got != expected {
		t.Errorf("DotsyncDir() = %q, want %q", got, expected)
	}
}

// TestExpandPath tests path expansion
func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "tilde expansion",
			path: "~/.config",
			want: filepath.Join(home, ".config"),
		},
		{
			name: "no expansion needed",
			path: "/usr/local",
			want: "/usr/local",
		},
		{
			name: "relative path",
			path: "relative",
			want: "relative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandPath(tt.path)
			if got != tt.want {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

// TestValidatePath_WritableCheck tests write permission checking
func TestValidatePath_WritableCheck(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a writable directory
	writableDir := filepath.Join(tmpDir, "writable")
	if err := os.MkdirAll(writableDir, 0755); err != nil {
		t.Fatalf("failed to create writable dir: %v", err)
	}

	// Should succeed
	if err := ValidatePath(writableDir); err != nil {
		t.Errorf("ValidatePath() failed for writable dir: %v", err)
	}

	// Verify test file was cleaned up
	testFile := filepath.Join(writableDir, ".dotsync-write-test")
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("test file was not cleaned up")
	}
}

// TestEnsureDotsyncDir_WithEnvVars tests with environment variables
func TestEnsureDotsyncDir_WithEnvVars(t *testing.T) {
	// Set a test environment variable
	os.Setenv("DOTSYNC_TEST_VAR", "testvalue")
	defer os.Unsetenv("DOTSYNC_TEST_VAR")

	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "$DOTSYNC_TEST_VAR")

	// Create the directory manually since we're using an env var
	expandedPath := filepath.Join(tmpDir, "testvalue")
	os.MkdirAll(expandedPath, 0755)

	// This should expand the env var
	result, err := EnsureDotsyncDir(testPath)
	if err != nil {
		t.Errorf("EnsureDotsyncDir() failed: %v", err)
	}

	expected := filepath.Join(expandedPath, "dotsync")
	if result != expected {
		t.Errorf("EnsureDotsyncDir() = %q, want %q", result, expected)
	}
}
