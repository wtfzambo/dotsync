// Package pathutil provides utilities for path inference and manipulation.
package pathutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// InferResult contains the result of path inference.
type InferResult struct {
	// Name is the inferred entry name (e.g., "opencode", "zsh")
	Name string
	// Root is the root directory for the entry (e.g., "~/.config/opencode")
	Root string
	// RelPath is the relative path from Root to the file
	RelPath string
}

// InferFromPath attempts to infer entry name and root from a file path.
// Returns nil if the path doesn't match any known pattern.
func InferFromPath(absPath string) *InferResult {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	// Normalize the path
	absPath = filepath.Clean(absPath)

	// Check if path is under home directory
	if !strings.HasPrefix(absPath, home) {
		return nil
	}

	// Get path relative to home
	relToHome, err := filepath.Rel(home, absPath)
	if err != nil {
		return nil
	}

	// Split into parts
	parts := strings.Split(relToHome, string(filepath.Separator))
	if len(parts) == 0 {
		return nil
	}

	// Pattern 1: ~/.config/<name>/*
	if parts[0] == ".config" && len(parts) >= 3 {
		name := parts[1]
		root := filepath.Join(home, ".config", name)
		relPath := filepath.Join(parts[2:]...)
		return &InferResult{
			Name:    name,
			Root:    contractHome(root, home),
			RelPath: relPath,
		}
	}

	// Pattern 2: ~/Library/Application Support/<name>/* (macOS)
	if runtime.GOOS == "darwin" && parts[0] == "Library" && len(parts) >= 4 && parts[1] == "Application Support" {
		name := parts[2]
		root := filepath.Join(home, "Library", "Application Support", name)
		relPath := filepath.Join(parts[3:]...)
		return &InferResult{
			Name:    name,
			Root:    contractHome(root, home),
			RelPath: relPath,
		}
	}

	// Pattern 3: ~/.<name>/* (hidden directory like ~/.aws/, ~/.vscode/)
	if strings.HasPrefix(parts[0], ".") && len(parts) >= 2 {
		name := strings.TrimPrefix(parts[0], ".")
		root := filepath.Join(home, parts[0])
		relPath := filepath.Join(parts[1:]...)
		return &InferResult{
			Name:    name,
			Root:    contractHome(root, home),
			RelPath: relPath,
		}
	}

	// Pattern 4: ~/.<name> (dotfile like ~/.zshrc, ~/.gitconfig)
	if strings.HasPrefix(parts[0], ".") && len(parts) == 1 {
		// Extract name: .zshrc -> zshrc, .gitconfig -> gitconfig
		name := strings.TrimPrefix(parts[0], ".")
		// Remove common extensions
		name = strings.TrimSuffix(name, "rc")
		if name == "" {
			name = strings.TrimPrefix(parts[0], ".")
		}
		return &InferResult{
			Name:    name,
			Root:    "~",
			RelPath: parts[0],
		}
	}

	return nil
}

// contractHome replaces the home directory with ~ in a path.
func contractHome(path, home string) string {
	if path == home {
		return "~"
	}
	if strings.HasPrefix(path, home+string(filepath.Separator)) {
		return "~" + path[len(home):]
	}
	return path
}

// ExpandHome expands ~ to the user's home directory.
func ExpandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if path == "~" {
		return home
	}
	// Handle both ~/ (Unix) and ~\ (Windows)
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
		return filepath.Join(home, path[2:])
	}

	return path
}

// ContractHome replaces the home directory with ~ in a path.
func ContractHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return contractHome(path, home)
}

// IsUnderHome checks if a path is under the user's home directory.
func IsUnderHome(absPath string) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	return strings.HasPrefix(absPath, home)
}

// AbsolutePath converts a path to an absolute path, expanding ~ if present.
func AbsolutePath(path string) (string, error) {
	expanded := ExpandHome(path)
	return filepath.Abs(expanded)
}
