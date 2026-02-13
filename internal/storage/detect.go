package storage

import (
	"os"
	"path/filepath"
	"strings"
)

// DetectPath attempts to find the cloud storage path for a provider.
// Returns the detected path or empty string if not found.
func DetectPath(provider Provider) string {
	paths := KnownPaths()
	if paths == nil {
		return ""
	}

	pp, ok := paths[provider]
	if !ok || pp.Primary == "" {
		return ""
	}

	// Try primary path (may contain glob)
	if found := findPath(pp.Primary); found != "" {
		return found
	}

	// Try fallback path
	if pp.Fallback != "" {
		if found := findPath(pp.Fallback); found != "" {
			return found
		}
	}

	return ""
}

// findPath expands and checks if a path exists.
// Supports glob patterns and ~ expansion.
func findPath(pattern string) string {
	// Expand ~ to home directory
	expanded := expandHome(pattern)

	// Handle Windows %USERPROFILE%
	expanded = os.ExpandEnv(expanded)

	// Check if path contains glob characters
	if strings.ContainsAny(expanded, "*?[]") {
		matches, err := filepath.Glob(expanded)
		if err != nil || len(matches) == 0 {
			return ""
		}
		// Return first match that exists and is a directory
		for _, m := range matches {
			if info, err := os.Stat(m); err == nil && info.IsDir() {
				return m
			}
		}
		return ""
	}

	// Direct path check
	if info, err := os.Stat(expanded); err == nil && info.IsDir() {
		return expanded
	}

	return ""
}

// expandHome expands ~ to the user's home directory.
func expandHome(path string) string {
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
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}

	return path
}

// ExpandHome is exported for use by other packages.
func ExpandHome(path string) string {
	return expandHome(path)
}

// ContractHome replaces the home directory with ~ in a path.
func ContractHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if path == home {
		return "~"
	}
	if strings.HasPrefix(path, home+string(filepath.Separator)) {
		return "~" + path[len(home):]
	}

	return path
}
