package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestAdd_WithEmptyName_Rejects tests bug #3a:
// Using --name with an empty string should reject the operation.
//
// EXPECTED: Error "empty name not allowed"
// ACTUAL (bug): Falls through to prompt or accepts it
func TestAdd_WithEmptyName_Rejects(t *testing.T) {
	// 1. Create temp storage + config
	tmpHome := t.TempDir()
	tmpStorage := t.TempDir()

	// Build the binary
	binPath := filepath.Join(t.TempDir(), "dotsync")
	buildCmd := exec.Command("go", "build", "-o", binPath, "../cmd/dotsync")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	runDotsync := func(args ...string) (string, error) {
		cmd := exec.Command(binPath, args...)
		cmd.Env = append(os.Environ(),
			"HOME="+tmpHome,
			"XDG_CONFIG_HOME="+filepath.Join(tmpHome, ".config"),
		)
		output, err := cmd.CombinedOutput()
		return string(output), err
	}

	// Initialize dotsync
	if output, err := runDotsync("init", "--path", tmpStorage); err != nil {
		t.Fatalf("init failed: %v\n%s", err, output)
	}

	// 2. Create a test file
	testDir := filepath.Join(tmpHome, ".config", "testapp")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	testFile := filepath.Join(testDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// 3. Run: dotsync add file.txt --name ""
	output, err := runDotsync("add", testFile, "--name", "")

	t.Logf("Add output: %s", output)
	t.Logf("Error: %v", err)

	// 4. EXPECTED: Error "empty name not allowed"
	if err == nil {
		t.Errorf("BUG DETECTED: Command succeeded with empty name")
		t.Errorf("EXPECTED: Should reject empty name with error")
		t.Errorf("ACTUAL: Command succeeded or fell through to prompt")
	}

	// 5. ACTUAL (bug): Falls through to prompt or accepts it
	if err != nil {
		// Check if error message is appropriate
		if !strings.Contains(output, "empty") && !strings.Contains(output, "name") && !strings.Contains(strings.ToLower(output), "required") {
			t.Logf("Command failed but with unexpected error message")
			t.Logf("EXPECTED error about empty name, got: %s", output)
		}
	}

	// Verify no file was added to storage with empty or invalid name
	storagePath := filepath.Join(tmpStorage, "dotsync")
	if entries, err := os.ReadDir(storagePath); err == nil {
		for _, entry := range entries {
			if entry.Name() == "" || entry.Name() == "." {
				t.Errorf("BUG: File was added with empty/invalid entry name: %s", entry.Name())
			}
		}
	}
}

// TestAdd_WithPathCharsInName_Rejects tests bug #3b:
// Using --name with path separators should reject the operation.
//
// EXPECTED: Error "name cannot contain /"
// ACTUAL (bug): Accepts it, creates nested dirs
func TestAdd_WithPathCharsInName_Rejects(t *testing.T) {
	// 1. Create temp storage + config
	tmpHome := t.TempDir()
	tmpStorage := t.TempDir()

	// Build the binary
	binPath := filepath.Join(t.TempDir(), "dotsync")
	buildCmd := exec.Command("go", "build", "-o", binPath, "../cmd/dotsync")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	runDotsync := func(args ...string) (string, error) {
		cmd := exec.Command(binPath, args...)
		cmd.Env = append(os.Environ(),
			"HOME="+tmpHome,
			"XDG_CONFIG_HOME="+filepath.Join(tmpHome, ".config"),
		)
		output, err := cmd.CombinedOutput()
		return string(output), err
	}

	// Initialize dotsync
	if output, err := runDotsync("init", "--path", tmpStorage); err != nil {
		t.Fatalf("init failed: %v\n%s", err, output)
	}

	// 2. Create a test file
	testDir := filepath.Join(tmpHome, ".config", "testapp")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	testFile := filepath.Join(testDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// 3. Run: dotsync add file.txt --name "foo/bar"
	output, err := runDotsync("add", testFile, "--name", "foo/bar")

	t.Logf("Add output: %s", output)
	t.Logf("Error: %v", err)

	// 4. EXPECTED: Error "name cannot contain /"
	if err == nil {
		t.Errorf("BUG DETECTED: Command succeeded with path separator in name")
		t.Errorf("EXPECTED: Should reject name containing '/'")
		t.Errorf("ACTUAL: Command succeeded")

		// Check if nested directories were created (bug behavior)
		nestedPath := filepath.Join(tmpStorage, "dotsync", "foo", "bar")
		if info, statErr := os.Stat(nestedPath); statErr == nil && info.IsDir() {
			t.Errorf("BUG CONFIRMED: Nested directories created at %s", nestedPath)
			t.Errorf("This creates an invalid manifest structure")
		}
	}

	// 5. ACTUAL (bug): Accepts it, creates nested dirs
	if err != nil {
		// Check if error message is appropriate
		if !strings.Contains(output, "cannot contain") && !strings.Contains(output, "invalid") {
			t.Logf("Command failed but with unexpected error message")
			t.Logf("EXPECTED error about path separator, got: %s", output)
		}
	}

	// Verify the manifest structure
	// Entry names should be simple identifiers, not paths
	storagePath := filepath.Join(tmpStorage, "dotsync")
	if entries, err := os.ReadDir(storagePath); err == nil {
		for _, entry := range entries {
			if strings.Contains(entry.Name(), "/") || strings.Contains(entry.Name(), "\\") {
				t.Errorf("BUG: Entry name contains path separator: %s", entry.Name())
			}
			// Check for "foo" directory that shouldn't exist at top level
			if entry.Name() == "foo" && entry.IsDir() {
				t.Errorf("BUG CONFIRMED: Created 'foo' directory instead of rejecting 'foo/bar' name")

				// Check if 'bar' subdirectory exists
				fooPath := filepath.Join(storagePath, "foo")
				if subEntries, subErr := os.ReadDir(fooPath); subErr == nil {
					for _, subEntry := range subEntries {
						if subEntry.Name() == "bar" {
							t.Errorf("BUG CONFIRMED: Created nested 'foo/bar' directory structure")
						}
					}
				}
			}
		}
	}

	// Additional test cases for other invalid characters
	invalidNames := []string{
		"foo/bar/baz",           // Multiple separators
		"../backdoor",           // Parent directory traversal
		"./current",             // Current directory prefix
		"name\\with\\backslash", // Windows-style separator
	}

	for _, invalidName := range invalidNames {
		output, err := runDotsync("add", testFile, "--name", invalidName)
		if err == nil {
			t.Errorf("BUG: Command succeeded with invalid name: %q", invalidName)
			t.Logf("  Output: %s", output)
		}
	}
}
