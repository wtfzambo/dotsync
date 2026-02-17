package symlink

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestCreate tests symlink creation
func TestCreate(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "target.txt")
	linkFile := filepath.Join(tmpDir, "link.txt")

	// Create target file
	if err := os.WriteFile(targetFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	// Create symlink
	if err := Create(linkFile, targetFile); err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Verify symlink exists
	info, err := os.Lstat(linkFile)
	if err != nil {
		t.Fatalf("symlink doesn't exist: %v", err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("created file is not a symlink")
	}

	// Verify target
	target, err := os.Readlink(linkFile)
	if err != nil {
		t.Fatalf("failed to read symlink: %v", err)
	}

	if target != targetFile {
		t.Errorf("symlink target = %q, want %q", target, targetFile)
	}
}

// TestCreate_WithNestedDirectories tests symlink creation with parent dirs
func TestCreate_WithNestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "target.txt")
	linkFile := filepath.Join(tmpDir, "nested", "dir", "link.txt")

	// Create target file
	if err := os.WriteFile(targetFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	// Create symlink (should create parent directories)
	if err := Create(linkFile, targetFile); err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Verify symlink exists
	info, err := os.Lstat(linkFile)
	if err != nil {
		t.Fatalf("symlink doesn't exist: %v", err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("created file is not a symlink")
	}
}

// TestCreate_ParentDirectoryPermissions tests that parent directories are created with 755 permissions
func TestCreate_ParentDirectoryPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping permission test on Windows - permissions work differently")
	}

	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "target.txt")
	parentDir := filepath.Join(tmpDir, "nested", "deep")
	linkFile := filepath.Join(parentDir, "link.txt")

	// Create target file
	if err := os.WriteFile(targetFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	// Create symlink (should create parent directories)
	if err := Create(linkFile, targetFile); err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Check each parent directory has 0755 permissions
	expectedPerm := os.FileMode(0755)

	// Check "nested" directory
	nestedInfo, err := os.Stat(filepath.Join(tmpDir, "nested"))
	if err != nil {
		t.Fatalf("failed to stat nested directory: %v", err)
	}
	if nestedInfo.Mode().Perm() != expectedPerm {
		t.Errorf("nested directory permissions = %v (%04o), want %v (%04o)",
			nestedInfo.Mode().Perm(), nestedInfo.Mode().Perm(), expectedPerm, expectedPerm)
	}

	// Check "deep" directory
	deepInfo, err := os.Stat(parentDir)
	if err != nil {
		t.Fatalf("failed to stat deep directory: %v", err)
	}
	if deepInfo.Mode().Perm() != expectedPerm {
		t.Errorf("deep directory permissions = %v (%04o), want %v (%04o)",
			deepInfo.Mode().Perm(), deepInfo.Mode().Perm(), expectedPerm, expectedPerm)
	}
}

// TestRemove tests symlink removal
func TestRemove(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "target.txt")
	linkFile := filepath.Join(tmpDir, "link.txt")

	// Create target and symlink
	os.WriteFile(targetFile, []byte("content"), 0644)
	os.Symlink(targetFile, linkFile)

	// Remove symlink
	if err := Remove(linkFile); err != nil {
		t.Fatalf("Remove() failed: %v", err)
	}

	// Verify symlink is gone
	if _, err := os.Lstat(linkFile); !os.IsNotExist(err) {
		t.Error("symlink should be removed")
	}

	// Verify target still exists
	if _, err := os.Stat(targetFile); err != nil {
		t.Error("target file should not be removed")
	}
}

// TestRemove_NotSymlink tests removing non-symlink fails
func TestRemove_NotSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	regularFile := filepath.Join(tmpDir, "regular.txt")

	// Create a regular file
	if err := os.WriteFile(regularFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Try to remove as symlink
	err := Remove(regularFile)
	if err == nil {
		t.Fatal("Remove() should fail for regular file")
	}
}

// TestRemove_AlreadyGone tests removing non-existent symlink
func TestRemove_AlreadyGone(t *testing.T) {
	tmpDir := t.TempDir()
	linkFile := filepath.Join(tmpDir, "nonexistent.txt")

	// Should not error
	if err := Remove(linkFile); err != nil {
		t.Errorf("Remove() should not error for non-existent file: %v", err)
	}
}

// TestIsSymlink tests symlink detection
func TestIsSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "target.txt")
	linkFile := filepath.Join(tmpDir, "link.txt")
	regularFile := filepath.Join(tmpDir, "regular.txt")

	// Create files
	os.WriteFile(targetFile, []byte("content"), 0644)
	os.WriteFile(regularFile, []byte("content"), 0644)
	os.Symlink(targetFile, linkFile)

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"symlink", linkFile, true},
		{"regular file", regularFile, false},
		{"non-existent", filepath.Join(tmpDir, "nonexistent"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsSymlink(tt.path)
			if err != nil {
				t.Fatalf("IsSymlink() error: %v", err)
			}
			if got != tt.want {
				t.Errorf("IsSymlink() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestReadTarget tests reading symlink target
func TestReadTarget(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "target.txt")
	linkFile := filepath.Join(tmpDir, "link.txt")

	// Create target and symlink
	os.WriteFile(targetFile, []byte("content"), 0644)
	os.Symlink(targetFile, linkFile)

	// Read target
	target, err := ReadTarget(linkFile)
	if err != nil {
		t.Fatalf("ReadTarget() failed: %v", err)
	}

	if target != targetFile {
		t.Errorf("ReadTarget() = %q, want %q", target, targetFile)
	}
}

// TestCheck tests symlink status checking
func TestCheck(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "target.txt")
	linkFile := filepath.Join(tmpDir, "link.txt")
	incorrectLink := filepath.Join(tmpDir, "incorrect.txt")
	wrongTarget := filepath.Join(tmpDir, "wrong.txt")
	brokenLink := filepath.Join(tmpDir, "broken.txt")
	regularFile := filepath.Join(tmpDir, "regular.txt")

	// Create target file
	os.WriteFile(targetFile, []byte("content"), 0644)

	// Create wrong target file
	os.WriteFile(wrongTarget, []byte("wrong"), 0644)

	// Create correct symlink
	os.Symlink(targetFile, linkFile)

	// Create incorrect symlink (points to wrong target, but target exists)
	os.Symlink(wrongTarget, incorrectLink)

	// Create broken symlink (target doesn't exist)
	os.Symlink(filepath.Join(tmpDir, "nonexistent"), brokenLink)

	// Create regular file
	os.WriteFile(regularFile, []byte("content"), 0644)

	tests := []struct {
		name           string
		path           string
		expectedTarget string
		wantStatus     Status
	}{
		{"linked correctly", linkFile, targetFile, StatusLinked},
		{"broken symlink", brokenLink, filepath.Join(tmpDir, "nonexistent"), StatusBroken},
		{"incorrect target", incorrectLink, targetFile, StatusIncorrect},
		{"not linked", regularFile, targetFile, StatusNotLinked},
		{"doesn't exist", filepath.Join(tmpDir, "nonexistent"), targetFile, StatusNotExist},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, _, err := Check(tt.path, tt.expectedTarget)
			if err != nil {
				t.Fatalf("Check() error: %v", err)
			}
			if status != tt.wantStatus {
				t.Errorf("Check() status = %v (%s), want %v (%s)",
					status, status.String(), tt.wantStatus, tt.wantStatus.String())
			}
		})
	}
}

// TestStatus_String tests Status string representation
func TestStatus_String(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{StatusNotExist, "not exist"},
		{StatusLinked, "linked"},
		{StatusBroken, "broken"},
		{StatusIncorrect, "incorrect"},
		{StatusNotLinked, "not linked"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.status.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestMoveFile tests file moving
func TestMoveFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "src.txt")
	dstFile := filepath.Join(tmpDir, "subdir", "dst.txt")
	content := []byte("test content")

	// Create source file
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatalf("failed to create source: %v", err)
	}

	// Move file
	if err := MoveFile(srcFile, dstFile); err != nil {
		t.Fatalf("MoveFile() failed: %v", err)
	}

	// Verify source is gone
	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Error("source file should be removed")
	}

	// Verify destination exists with correct content
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read destination: %v", err)
	}

	if string(dstContent) != string(content) {
		t.Errorf("content = %q, want %q", dstContent, content)
	}
}

// TestMoveFile_PreservesPermissions tests file moving preserves permissions
func TestMoveFile_PreservesPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "src.txt")
	dstFile := filepath.Join(tmpDir, "dst.txt")

	// Create source file with specific permissions
	if err := os.WriteFile(srcFile, []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create source: %v", err)
	}

	srcInfo, _ := os.Stat(srcFile)

	// Move file
	if err := MoveFile(srcFile, dstFile); err != nil {
		t.Fatalf("MoveFile() failed: %v", err)
	}

	// Check permissions
	dstInfo, err := os.Stat(dstFile)
	if err != nil {
		t.Fatalf("failed to stat destination: %v", err)
	}

	if srcInfo.Mode() != dstInfo.Mode() {
		t.Errorf("permissions not preserved: src %v, dst %v", srcInfo.Mode(), dstInfo.Mode())
	}
}

// TestCopyFile tests file copying
func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "src.txt")
	dstFile := filepath.Join(tmpDir, "subdir", "dst.txt")
	content := []byte("test content")

	// Create source file
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatalf("failed to create source: %v", err)
	}

	// Copy file
	if err := CopyFile(srcFile, dstFile); err != nil {
		t.Fatalf("CopyFile() failed: %v", err)
	}

	// Verify source still exists
	if _, err := os.Stat(srcFile); err != nil {
		t.Error("source file should still exist")
	}

	// Verify destination exists with correct content
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read destination: %v", err)
	}

	if string(dstContent) != string(content) {
		t.Errorf("content = %q, want %q", dstContent, content)
	}
}

// TestCopyFile_PreservesPermissions tests file copying preserves permissions
func TestCopyFile_PreservesPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "src.txt")
	dstFile := filepath.Join(tmpDir, "dst.txt")

	// Create source file with specific permissions
	if err := os.WriteFile(srcFile, []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create source: %v", err)
	}

	srcInfo, _ := os.Stat(srcFile)

	// Copy file
	if err := CopyFile(srcFile, dstFile); err != nil {
		t.Fatalf("CopyFile() failed: %v", err)
	}

	// Check permissions
	dstInfo, err := os.Stat(dstFile)
	if err != nil {
		t.Fatalf("failed to stat destination: %v", err)
	}

	if srcInfo.Mode() != dstInfo.Mode() {
		t.Errorf("permissions not preserved: src %v, dst %v", srcInfo.Mode(), dstInfo.Mode())
	}
}

// TestCheck_AbsolutePaths tests Check with absolute paths
func TestCheck_AbsolutePaths(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "target.txt")
	linkFile := filepath.Join(tmpDir, "link.txt")

	// Create target file
	os.WriteFile(targetFile, []byte("content"), 0644)

	// Create symlink with absolute path
	os.Symlink(targetFile, linkFile)

	// Check should work with absolute path
	status, actualTarget, err := Check(linkFile, targetFile)
	if err != nil {
		t.Fatalf("Check() error: %v", err)
	}

	if status != StatusLinked {
		t.Errorf("status = %v, want %v", status, StatusLinked)
	}

	if actualTarget != targetFile {
		t.Errorf("actualTarget = %q, want %q", actualTarget, targetFile)
	}
}
