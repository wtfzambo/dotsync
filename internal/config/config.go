// Package config handles the local dotsync configuration.
// The config is machine-specific and stored in ~/.config/dotsync/config.json
package config

// Config represents the local dotsync configuration.
// This is NOT synced - it's machine-specific.
type Config struct {
	// StoragePath is the path to the cloud storage folder
	// e.g., "~/Library/CloudStorage/GoogleDrive-user@gmail.com/My Drive"
	StoragePath string `json:"storagePath"`
}

// New creates a new config with the given storage path.
func New(storagePath string) *Config {
	return &Config{
		StoragePath: storagePath,
	}
}
