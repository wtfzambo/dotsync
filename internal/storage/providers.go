// Package storage handles cloud storage detection and validation.
package storage

import "runtime"

// Provider represents a cloud storage provider.
type Provider string

const (
	ProviderGoogleDrive Provider = "gdrive"
	ProviderDropbox     Provider = "dropbox"
	ProviderICloud      Provider = "icloud"
)

// ParseProvider converts a string to a Provider.
// Returns empty string if not recognized.
func ParseProvider(s string) Provider {
	switch s {
	case "gdrive", "googledrive", "google-drive":
		return ProviderGoogleDrive
	case "dropbox":
		return ProviderDropbox
	case "icloud":
		return ProviderICloud
	default:
		return ""
	}
}

// String returns the string representation of the provider.
func (p Provider) String() string {
	return string(p)
}

// DisplayName returns a human-readable name for the provider.
func (p Provider) DisplayName() string {
	switch p {
	case ProviderGoogleDrive:
		return "Google Drive"
	case ProviderDropbox:
		return "Dropbox"
	case ProviderICloud:
		return "iCloud Drive"
	default:
		return string(p)
	}
}

// ProviderPaths contains known paths for a cloud storage provider.
type ProviderPaths struct {
	// Primary is the main path to check (may contain glob patterns)
	Primary string
	// Fallback is an alternative path to check
	Fallback string
}

// KnownPaths returns the known paths for each provider on the current platform.
func KnownPaths() map[Provider]ProviderPaths {
	switch runtime.GOOS {
	case "darwin":
		return macOSPaths
	case "linux":
		return linuxPaths
	case "windows":
		return windowsPaths
	default:
		return nil
	}
}

// macOS cloud storage paths
var macOSPaths = map[Provider]ProviderPaths{
	ProviderGoogleDrive: {
		Primary:  "~/Library/CloudStorage/GoogleDrive-*/My Drive",
		Fallback: "~/Google Drive",
	},
	ProviderDropbox: {
		Primary:  "~/Library/CloudStorage/Dropbox",
		Fallback: "~/Dropbox",
	},
	ProviderICloud: {
		Primary:  "~/Library/Mobile Documents/com~apple~CloudDocs",
		Fallback: "",
	},
}

// Linux cloud storage paths
var linuxPaths = map[Provider]ProviderPaths{
	ProviderGoogleDrive: {
		Primary:  "~/Google Drive",
		Fallback: "~/google-drive",
	},
	ProviderDropbox: {
		Primary:  "~/Dropbox",
		Fallback: "",
	},
}

// Windows cloud storage paths (future - paths TBD)
var windowsPaths = map[Provider]ProviderPaths{
	ProviderGoogleDrive: {
		Primary:  "%USERPROFILE%\\Google Drive",
		Fallback: "",
	},
	ProviderDropbox: {
		Primary:  "%USERPROFILE%\\Dropbox",
		Fallback: "",
	},
	ProviderICloud: {
		Primary:  "%USERPROFILE%\\iCloudDrive",
		Fallback: "",
	},
}

// SupportedProviders returns the list of supported providers for the current platform.
func SupportedProviders() []Provider {
	paths := KnownPaths()
	if paths == nil {
		return nil
	}

	var providers []Provider
	for p, pp := range paths {
		if pp.Primary != "" {
			providers = append(providers, p)
		}
	}
	return providers
}
