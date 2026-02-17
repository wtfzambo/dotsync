package pathutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestInferFromPath_XDGConfig tests inference for ~/.config/<name>/* pattern
func TestInferFromPath_XDGConfig(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		wantName string
		wantRoot string
		wantRel  string
	}{
		{
			name:     "opencode config file",
			path:     filepath.Join(home, ".config", "opencode", "config.json"),
			wantName: "opencode",
			wantRoot: filepath.Join("~", ".config", "opencode"),
			wantRel:  "config.json",
		},
		{
			name:     "nested config file",
			path:     filepath.Join(home, ".config", "cursor", "settings", "user.json"),
			wantName: "cursor",
			wantRoot: filepath.Join("~", ".config", "cursor"),
			wantRel:  filepath.Join("settings", "user.json"),
		},
		{
			name:     "vscode config",
			path:     filepath.Join(home, ".config", "Code", "User", "settings.json"),
			wantName: "Code",
			wantRoot: filepath.Join("~", ".config", "Code"),
			wantRel:  filepath.Join("User", "settings.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InferFromPath(tt.path)
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if result.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", result.Name, tt.wantName)
			}
			if result.Root != tt.wantRoot {
				t.Errorf("Root = %q, want %q", result.Root, tt.wantRoot)
			}
			if result.RelPath != tt.wantRel {
				t.Errorf("RelPath = %q, want %q", result.RelPath, tt.wantRel)
			}
		})
	}
}

// TestInferFromPath_HiddenDir tests inference for ~/.<name>/* pattern
func TestInferFromPath_HiddenDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		wantName string
		wantRoot string
		wantRel  string
	}{
		{
			name:     "ssh config",
			path:     filepath.Join(home, ".ssh", "config"),
			wantName: "ssh",
			wantRoot: filepath.Join("~", ".ssh"),
			wantRel:  "config",
		},
		{
			name:     "aws credentials",
			path:     filepath.Join(home, ".aws", "credentials"),
			wantName: "aws",
			wantRoot: filepath.Join("~", ".aws"),
			wantRel:  "credentials",
		},
		{
			name:     "nested hidden dir",
			path:     filepath.Join(home, ".vscode", "extensions", "settings.json"),
			wantName: "vscode",
			wantRoot: filepath.Join("~", ".vscode"),
			wantRel:  filepath.Join("extensions", "settings.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InferFromPath(tt.path)
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if result.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", result.Name, tt.wantName)
			}
			if result.Root != tt.wantRoot {
				t.Errorf("Root = %q, want %q", result.Root, tt.wantRoot)
			}
			if result.RelPath != tt.wantRel {
				t.Errorf("RelPath = %q, want %q", result.RelPath, tt.wantRel)
			}
		})
	}
}

// TestInferFromPath_Dotfile tests inference for ~/.<name> pattern (single dotfiles)
func TestInferFromPath_Dotfile(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		wantName string
		wantRoot string
		wantRel  string
	}{
		{
			name:     "zshrc",
			path:     filepath.Join(home, ".zshrc"),
			wantName: "zsh", // 'rc' suffix is trimmed
			wantRoot: "~",
			wantRel:  ".zshrc",
		},
		{
			name:     "bashrc",
			path:     filepath.Join(home, ".bashrc"),
			wantName: "bash",
			wantRoot: "~",
			wantRel:  ".bashrc",
		},
		{
			name:     "gitconfig",
			path:     filepath.Join(home, ".gitconfig"),
			wantName: "gitconfig",
			wantRoot: "~",
			wantRel:  ".gitconfig",
		},
		{
			name:     "vimrc",
			path:     filepath.Join(home, ".vimrc"),
			wantName: "vim",
			wantRoot: "~",
			wantRel:  ".vimrc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InferFromPath(tt.path)
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if result.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", result.Name, tt.wantName)
			}
			if result.Root != tt.wantRoot {
				t.Errorf("Root = %q, want %q", result.Root, tt.wantRoot)
			}
			if result.RelPath != tt.wantRel {
				t.Errorf("RelPath = %q, want %q", result.RelPath, tt.wantRel)
			}
		})
	}
}

// TestInferFromPath_macOSApplicationSupport tests macOS-specific pattern
func TestInferFromPath_macOSApplicationSupport(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping macOS-specific test")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		wantName string
		wantRoot string
		wantRel  string
	}{
		{
			name:     "cursor app support",
			path:     filepath.Join(home, "Library", "Application Support", "Cursor", "settings.json"),
			wantName: "Cursor",
			wantRoot: "~/Library/Application Support/Cursor",
			wantRel:  "settings.json",
		},
		{
			name:     "vscode app support",
			path:     filepath.Join(home, "Library", "Application Support", "Code", "User", "snippets.json"),
			wantName: "Code",
			wantRoot: "~/Library/Application Support/Code",
			wantRel:  filepath.Join("User", "snippets.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InferFromPath(tt.path)
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if result.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", result.Name, tt.wantName)
			}
			if result.Root != tt.wantRoot {
				t.Errorf("Root = %q, want %q", result.Root, tt.wantRoot)
			}
			if result.RelPath != tt.wantRel {
				t.Errorf("RelPath = %q, want %q", result.RelPath, tt.wantRel)
			}
		})
	}
}

// TestInferFromPath_EdgeCases tests edge cases
func TestInferFromPath_EdgeCases(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantNil bool
		descr   string
	}{
		{
			name:    "path outside home",
			path:    "/usr/local/bin/something",
			wantNil: true,
			descr:   "paths outside home should return nil",
		},
		{
			name:    "non-hidden file in home",
			path:    filepath.Join(home, "Documents", "file.txt"),
			wantNil: true,
			descr:   "non-hidden paths should return nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InferFromPath(tt.path)
			if tt.wantNil && result != nil {
				t.Errorf("expected nil result: %s, got %+v", tt.descr, result)
			}
			if !tt.wantNil && result == nil {
				t.Errorf("expected non-nil result: %s", tt.descr)
			}
		})
	}
}

// TestExpandHome tests home directory expansion
func TestExpandHome(t *testing.T) {
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
			name: "tilde only",
			path: "~",
			want: home,
		},
		{
			name: "tilde with path",
			path: "~/.config/opencode",
			want: filepath.Join(home, ".config", "opencode"),
		},
		{
			name: "no tilde",
			path: "/usr/local/bin",
			want: "/usr/local/bin",
		},
		{
			name: "relative path",
			path: "relative/path",
			want: "relative/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandHome(tt.path)
			if got != tt.want {
				t.Errorf("ExpandHome(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

// TestContractHome tests home directory contraction
func TestContractHome(t *testing.T) {
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
			name: "home directory",
			path: home,
			want: "~",
		},
		{
			name: "path under home",
			path: filepath.Join(home, ".config", "opencode"),
			want: filepath.Join("~", ".config", "opencode"),
		},
		{
			name: "path outside home",
			path: "/usr/local/bin",
			want: "/usr/local/bin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContractHome(tt.path)
			if got != tt.want {
				t.Errorf("ContractHome(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

// TestIsUnderHome tests home directory checking
func TestIsUnderHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "under home",
			path: filepath.Join(home, ".config", "opencode"),
			want: true,
		},
		{
			name: "home directory itself",
			path: home,
			want: true,
		},
		{
			name: "outside home",
			path: "/usr/local/bin",
			want: false,
		},
		{
			name: "relative to home but absolute path outside",
			path: "/home/otheruser/.config",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsUnderHome(tt.path)
			if got != tt.want {
				t.Errorf("IsUnderHome(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// TestAbsolutePath tests absolute path conversion
func TestAbsolutePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "tilde expansion",
			path:    "~/.config",
			want:    filepath.Join(home, ".config"),
			wantErr: false,
		},
		{
			name:    "already absolute",
			path:    filepath.Join(home, "test"),
			want:    filepath.Join(home, "test"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AbsolutePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("AbsolutePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("AbsolutePath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
