package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wtfzambo/dotsync/internal/backup"
	"github.com/wtfzambo/dotsync/internal/config"
	"github.com/wtfzambo/dotsync/internal/manifest"
	"github.com/wtfzambo/dotsync/internal/pathutil"
	"github.com/wtfzambo/dotsync/internal/symlink"
)

var linkCmd = &cobra.Command{
	Use:   "link [entry]",
	Short: "Create symlinks for tracked entries",
	Long: `Create symlinks at original locations pointing to cloud storage.

Use this command on a new machine to set up symlinks for entries
that were added on another machine.

If no entry name is provided, all entries will be linked.
If a file already exists at the target location, you'll be prompted
to backup, skip, or abort.`,
	Example: `  dotsync link           # Link all entries
  dotsync link opencode  # Link only the "opencode" entry
  dotsync link --backup  # Auto-backup existing files`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLink,
}

var linkBackup bool

func init() {
	linkCmd.Flags().BoolVarP(&linkBackup, "backup", "b", false, "Automatically backup existing files without prompting")
	rootCmd.AddCommand(linkCmd)
}

func runLink(cmd *cobra.Command, args []string) error {
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
			return fmt.Errorf("no manifest found. Nothing to link.\nUse 'dotsync add' to start tracking files")
		}
		return fmt.Errorf("loading manifest: %w", err)
	}

	if len(m.Entries) == 0 {
		fmt.Println("No entries to link")
		return nil
	}

	// 3. Determine which entries to link
	var entriesToLink map[string]manifest.Entry
	if len(args) > 0 {
		entryName := args[0]
		entry := m.GetEntry(entryName)
		if entry == nil {
			return fmt.Errorf("entry '%s' not found", entryName)
		}
		entriesToLink = map[string]manifest.Entry{entryName: *entry}
	} else {
		entriesToLink = m.Entries
	}

	// 4. Link each entry
	var linked, skipped, failed int

	for name, entry := range entriesToLink {
		fmt.Printf("\nLinking entry '%s':\n", name)

		entryRoot := pathutil.ExpandHome(entry.Root)
		for _, relPath := range entry.Files {
			originalPath := filepath.Join(entryRoot, relPath)
			cloudPath := filepath.Join(storagePath, "dotsync", name, relPath)

			result, err := linkFile(originalPath, cloudPath, linkBackup)
			switch result {
			case linkResultLinked:
				fmt.Printf("  [linked]  %s\n", relPath)
				linked++
			case linkResultSkipped:
				fmt.Printf("  [skipped] %s\n", relPath)
				skipped++
			case linkResultAlreadyLinked:
				fmt.Printf("  [ok]      %s (already linked)\n", relPath)
				// Don't count as linked or skipped
			case linkResultAborted:
				return fmt.Errorf("aborted")
			case linkResultFailed:
				fmt.Printf("  [failed]  %s: %v\n", relPath, err)
				failed++
			}
		}
	}

	// 5. Print summary
	fmt.Println()
	if linked > 0 || skipped > 0 || failed > 0 {
		fmt.Printf("Summary: %d linked, %d skipped, %d failed\n", linked, skipped, failed)
	}

	if failed > 0 {
		return fmt.Errorf("some files failed to link")
	}

	return nil
}

type linkResult int

const (
	linkResultLinked linkResult = iota
	linkResultSkipped
	linkResultAlreadyLinked
	linkResultAborted
	linkResultFailed
)

// linkFile creates a symlink at originalPath pointing to cloudPath.
// Handles existing files based on autoBackup flag or user prompt.
func linkFile(originalPath, cloudPath string, autoBackup bool) (linkResult, error) {
	// Check if cloud file exists
	if _, err := os.Stat(cloudPath); os.IsNotExist(err) {
		return linkResultFailed, fmt.Errorf("source file not found in cloud storage: %s", cloudPath)
	}

	// Check current state of original path
	status, actualTarget, err := symlink.Check(originalPath, cloudPath)
	if err != nil {
		return linkResultFailed, err
	}

	switch status {
	case symlink.StatusLinked:
		// Already correctly linked
		return linkResultAlreadyLinked, nil

	case symlink.StatusNotExist:
		// Path doesn't exist, safe to create symlink
		if err := symlink.Create(originalPath, cloudPath); err != nil {
			return linkResultFailed, err
		}
		return linkResultLinked, nil

	case symlink.StatusBroken:
		// Broken symlink - remove and recreate
		if err := symlink.Remove(originalPath); err != nil {
			return linkResultFailed, fmt.Errorf("removing broken symlink: %w", err)
		}
		if err := symlink.Create(originalPath, cloudPath); err != nil {
			return linkResultFailed, err
		}
		return linkResultLinked, nil

	case symlink.StatusIncorrect:
		// Symlink exists but points elsewhere
		fmt.Printf("  Symlink exists but points to: %s\n", actualTarget)
		fmt.Printf("  Expected: %s\n", cloudPath)
		action := promptConflictAction(originalPath, autoBackup)
		return handleConflict(originalPath, cloudPath, action)

	case symlink.StatusNotLinked:
		// Regular file exists - need to handle conflict
		action := promptConflictAction(originalPath, autoBackup)
		return handleConflict(originalPath, cloudPath, action)

	default:
		return linkResultFailed, fmt.Errorf("unexpected symlink status: %v", status)
	}
}

type conflictAction int

const (
	conflictBackup conflictAction = iota
	conflictSkip
	conflictAbort
)

// promptConflictAction prompts the user for how to handle an existing file.
func promptConflictAction(path string, autoBackup bool) conflictAction {
	if autoBackup {
		return conflictBackup
	}

	fmt.Printf("  File exists: %s\n", pathutil.ContractHome(path))
	fmt.Printf("  [b]ackup and link, [s]kip, [a]bort? ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	switch response {
	case "b", "backup":
		return conflictBackup
	case "s", "skip":
		return conflictSkip
	case "a", "abort":
		return conflictAbort
	default:
		// Default to skip for safety
		fmt.Println("  Invalid response, skipping")
		return conflictSkip
	}
}

// handleConflict handles a file conflict based on the chosen action.
func handleConflict(originalPath, cloudPath string, action conflictAction) (linkResult, error) {
	switch action {
	case conflictBackup:
		// Backup existing file
		bk, err := backup.Create(originalPath)
		if err != nil {
			return linkResultFailed, fmt.Errorf("creating backup: %w", err)
		}
		fmt.Printf("  Backed up to: %s\n", bk.BackupPath)

		// Remove existing file/symlink
		if err := os.Remove(originalPath); err != nil {
			bk.Restore()
			return linkResultFailed, fmt.Errorf("removing existing file: %w", err)
		}

		// Create symlink
		if err := symlink.Create(originalPath, cloudPath); err != nil {
			bk.Restore()
			return linkResultFailed, err
		}
		return linkResultLinked, nil

	case conflictSkip:
		return linkResultSkipped, nil

	case conflictAbort:
		return linkResultAborted, nil

	default:
		return linkResultSkipped, nil
	}
}
