package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dotsync",
	Short: "Sync config files across machines using cloud storage",
	Long: `dotsync syncs developer config files (dotfiles, AI tool configs, etc.)
across multiple machines by leveraging cloud storage providers like
Google Drive, Dropbox, and iCloud.

The tool manages symlinks between your config files and cloud storage,
letting the cloud provider handle the actual synchronization.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags can be added here
}
