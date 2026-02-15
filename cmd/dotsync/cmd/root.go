package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
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

// SetVersion sets the version info at build time
func SetVersion(v, c, d, b string) {
	version, commit, date, builtBy = v, c, d, b
	rootCmd.Version = version
	rootCmd.SetVersionTemplate(fmt.Sprintf(`dotsync version {{.Version}}
  commit: %s
  built:  %s
  by:     %s
`, commit, date, builtBy))
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
