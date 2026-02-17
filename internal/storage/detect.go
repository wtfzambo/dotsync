package storage

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/wtfzambo/dotsync/internal/pathutil"
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
	expanded := pathutil.ExpandHome(pattern)

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

// ExpandHome is exported for use by other packages.
// Deprecated: Use pathutil.ExpandHome instead.
func ExpandHome(path string) string {
	return pathutil.ExpandHome(path)
}

// ContractHome replaces the home directory with ~ in a path.
// Deprecated: Use pathutil.ContractHome instead.
func ContractHome(path string) string {
	return pathutil.ContractHome(path)
}
