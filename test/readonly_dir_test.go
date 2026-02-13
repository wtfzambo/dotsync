package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// TestAddFromReadOnlyDirectory_FailsBeforeCopy tests bug #1:
// Adding a file from a read-only parent directory should fail BEFORE copying the file,
// not leave it in both locations.
//
// EXPECTED: Error before any copy happens
// ACTUAL (bug): File gets copied to storage but not deleted from original location
func TestAddFromReadOnlyDirectory_FailsBeforeCopy(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping read-only directory test on Windows")
	}

	// 1. Create temp storage dir
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

	// 2. Initialize dotsync config
	if output, err := runDotsync("init", "--path", tmpStorage); err != nil {
		t.Fatalf("init failed: %v\n%s", err, output)
	}

	// 3. Create a file in a directory, make dir read-only (chmod 555)
	readonlyDir := filepath.Join(tmpHome, "readonly")
	if err := os.MkdirAll(readonlyDir, 0755); err != nil {
		t.Fatalf("failed to create readonly dir: %v", err)
	}

	testFile := filepath.Join(readonlyDir, "config.txt")
	testContent := []byte("test content")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Make directory read-only (no write/execute permissions)
	if err := os.Chmod(readonlyDir, 0555); err != nil {
		t.Fatalf("failed to chmod directory: %v", err)
	}
	defer os.Chmod(readonlyDir, 0755) // Restore permissions for cleanup

	// 4. Try to add the file
	output, err := runDotsync("add", testFile, "--name", "testapp")

	// 5. EXPECTED (currently fails): Error before any copy happens
	// The command should fail because we can't delete the file from read-only directory
	if err == nil {
		t.Error("EXPECTED: add should fail for file in read-only directory")
	}

	// 6. ACTUAL (bug): File gets copied to storage but not deleted from original
	// The file should NOT exist in both places
	destPath := filepath.Join(tmpStorage, "dotsync", "testapp", "config.txt")

	originalExists := false
	if _, err := os.Stat(testFile); err == nil {
		originalExists = true
	}

	storageExists := false
	if _, err := os.Stat(destPath); err == nil {
		storageExists = true
	}

	t.Logf("Add output: %s", output)
	t.Logf("Original file exists: %v", originalExists)
	t.Logf("Storage file exists: %v", storageExists)

	// BUG: Currently both files exist because:
	// - File is copied to storage successfully
	// - Delete from original location fails (read-only directory)
	// - No rollback happens, so file remains in both locations
	if originalExists && storageExists {
		t.Errorf("BUG DETECTED: File exists in both locations after failed add")
		t.Errorf("  - Original: %s (exists: %v)", testFile, originalExists)
		t.Errorf("  - Storage:  %s (exists: %v)", destPath, storageExists)
		t.Errorf("EXPECTED: Either operation succeeds completely OR fails completely (no partial state)")
	}

	// The correct behavior would be:
	// - Check if we can delete the original file BEFORE copying
	// - If not, fail immediately without copying
	// - OR: Copy succeeds but delete fails, so rollback the copy
	if !originalExists && storageExists {
		t.Logf("CORRECT: File only exists in storage (operation succeeded)")
	} else if originalExists && !storageExists {
		t.Logf("CORRECT: File only exists in original location (operation failed cleanly)")
	}
}
