package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wtfzambo/dotsync/internal/config"
	"github.com/wtfzambo/dotsync/internal/manifest"
	"github.com/wtfzambo/dotsync/internal/pathutil"
	"github.com/wtfzambo/dotsync/internal/symlink"
)

var unlinkCmd = &cobra.Command{
	Use:   "unlink [entry]",
	Short: "Remove symlinks and restore files locally",
	Long: `Remove symlinks and copy files from cloud storage back to original locations.

This restores files to be regular files (not symlinks) while keeping
the cloud copy intact. You can re-link later with "dotsync link".

If no entry name is provided, all entries will be unlinked.`,
	Example: `  dotsync unlink           # Unlink all entries
  dotsync unlink opencode  # Unlink only the "opencode" entry`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUnlink,
}

func init() {
	rootCmd.AddCommand(unlinkCmd)
}

func runUnlink(cmd *cobra.Command, args []string) error {
	// 1. Load config (must be initialized)
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	if cfg == nil {
		return fmt.Errorf("dotsync not initialized. Run 'dotsync init <provider>' first")
	}

	storagePath := pathutil.ExpandHome(cfg.StoragePath)

	// Verify storage is available
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		return fmt.Errorf("storage unavailable: %s\nMake sure your cloud storage is mounted/syncing", storagePath)
	}

	// 2. Load manifest
	m, err := manifest.Load(storagePath)
	if err != nil {
		if strings.Contains(err.Error(), "manifest not found") {
			return fmt.Errorf("no manifest found. Nothing to unlink")
		}
		return fmt.Errorf("loading manifest: %w", err)
	}

	if len(m.Entries) == 0 {
		fmt.Println("No entries to unlink")
		return nil
	}

	// 3. Determine which entries to unlink
	var entriesToUnlink map[string]manifest.Entry
	if len(args) > 0 {
		entryName := args[0]
		entry := m.GetEntry(entryName)
		if entry == nil {
			return fmt.Errorf("entry '%s' not found", entryName)
		}
		entriesToUnlink = map[string]manifest.Entry{entryName: *entry}
	} else {
		entriesToUnlink = m.Entries
	}

	// 4. Unlink each entry
	var unlinked, skipped, failed int

	for name, entry := range entriesToUnlink {
		fmt.Printf("\nUnlinking entry '%s':\n", name)

		entryRoot := pathutil.ExpandHome(entry.Root)
		for _, relPath := range entry.Files {
			originalPath := filepath.Join(entryRoot, relPath)
			cloudPath := filepath.Join(storagePath, "dotsync", name, relPath)

			result, err := unlinkFile(originalPath, cloudPath)
			switch result {
			case unlinkResultUnlinked:
				fmt.Printf("  [unlinked] %s\n", relPath)
				unlinked++
			case unlinkResultSkipped:
				fmt.Printf("  [skipped]  %s (not a symlink)\n", relPath)
				skipped++
			case unlinkResultNotExist:
				fmt.Printf("  [skipped]  %s (doesn't exist)\n", relPath)
				skipped++
			case unlinkResultFailed:
				fmt.Printf("  [failed]   %s: %v\n", relPath, err)
				failed++
			}
		}
	}

	// 5. Print summary
	fmt.Println()
	if unlinked > 0 || skipped > 0 || failed > 0 {
		fmt.Printf("Summary: %d unlinked, %d skipped, %d failed\n", unlinked, skipped, failed)
	}

	// Note: We don't modify the manifest - entries stay tracked so they can be re-linked
	fmt.Println("\nFiles are now regular files. Use 'dotsync link' to restore symlinks.")

	if failed > 0 {
		return fmt.Errorf("some files failed to unlink")
	}

	return nil
}

type unlinkResult int

const (
	unlinkResultUnlinked unlinkResult = iota
	unlinkResultSkipped
	unlinkResultNotExist
	unlinkResultFailed
)

// unlinkFile removes a symlink and copies the file from cloud storage.
func unlinkFile(originalPath, cloudPath string) (unlinkResult, error) {
	// Check current state
	status, _, err := symlink.Check(originalPath, cloudPath)
	if err != nil {
		return unlinkResultFailed, err
	}

	switch status {
	case symlink.StatusNotExist:
		// Nothing to unlink
		return unlinkResultNotExist, nil

	case symlink.StatusNotLinked:
		// Already a regular file
		return unlinkResultSkipped, nil

	case symlink.StatusLinked, symlink.StatusIncorrect:
		// It's a symlink - remove it and copy file
		return doUnlink(originalPath, cloudPath)

	case symlink.StatusBroken:
		// Broken symlink - just remove it, warn about missing source
		if err := symlink.Remove(originalPath); err != nil {
			return unlinkResultFailed, fmt.Errorf("removing broken symlink: %w", err)
		}
		fmt.Printf("  Warning: source file missing in cloud storage: %s\n", pathutil.ContractHome(cloudPath))
		return unlinkResultUnlinked, nil

	default:
		return unlinkResultFailed, fmt.Errorf("unexpected status: %v", status)
	}
}

// doUnlink performs the actual unlink operation.
func doUnlink(originalPath, cloudPath string) (unlinkResult, error) {
	// Verify cloud file exists
	if _, err := os.Stat(cloudPath); os.IsNotExist(err) {
		// Cloud file missing - just remove symlink and warn
		if err := symlink.Remove(originalPath); err != nil {
			return unlinkResultFailed, fmt.Errorf("removing symlink: %w", err)
		}
		fmt.Printf("  Warning: source file missing in cloud storage, symlink removed\n")
		return unlinkResultUnlinked, nil
	}

	// Remove symlink first
	if err := symlink.Remove(originalPath); err != nil {
		return unlinkResultFailed, fmt.Errorf("removing symlink: %w", err)
	}

	// Copy file from cloud to original location
	if err := symlink.CopyFile(cloudPath, originalPath); err != nil {
		// Try to restore symlink on failure
		symlink.Create(originalPath, cloudPath)
		return unlinkResultFailed, fmt.Errorf("copying file: %w", err)
	}

	return unlinkResultUnlinked, nil
}
