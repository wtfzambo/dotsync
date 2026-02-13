package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wtfzambo/dotsync/internal/config"
	"github.com/wtfzambo/dotsync/internal/manifest"
	"github.com/wtfzambo/dotsync/internal/pathutil"
	"github.com/wtfzambo/dotsync/internal/symlink"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tracked entries and their status",
	Long: `List all tracked entries from the manifest.

Shows entry names, file counts, and link status on this machine.
Use --details to see individual files within each entry.`,
	Example: `  dotsync list           # Show entries overview
  dotsync list --details # Show all files in each entry`,
	Args: cobra.NoArgs,
	RunE: runList,
}

var listDetails bool

func init() {
	listCmd.Flags().BoolVarP(&listDetails, "details", "d", false, "Show detailed file list for each entry")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
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
			fmt.Println("No entries tracked yet.")
			fmt.Println("Use 'dotsync add <path>' to start tracking files.")
			return nil
		}
		return fmt.Errorf("loading manifest: %w", err)
	}

	if len(m.Entries) == 0 {
		fmt.Println("No entries tracked yet.")
		fmt.Println("Use 'dotsync add <path>' to start tracking files.")
		return nil
	}

	// 3. Sort entry names for consistent output
	names := make([]string, 0, len(m.Entries))
	for name := range m.Entries {
		names = append(names, name)
	}
	sort.Strings(names)

	// 4. Display entries
	for _, name := range names {
		entry := m.Entries[name]
		displayEntry(name, entry, storagePath, listDetails)
	}

	return nil
}

// displayEntry prints information about a single entry.
func displayEntry(name string, entry manifest.Entry, storagePath string, showDetails bool) {
	entryRoot := pathutil.ExpandHome(entry.Root)

	// Count file statuses
	var linked, notLinked, broken, incorrect int
	fileStatuses := make([]struct {
		file   string
		status symlink.Status
	}, 0, len(entry.Files))

	for _, relPath := range entry.Files {
		originalPath := filepath.Join(entryRoot, relPath)
		cloudPath := filepath.Join(storagePath, "dotsync", name, relPath)

		status, _, _ := symlink.Check(originalPath, cloudPath)
		fileStatuses = append(fileStatuses, struct {
			file   string
			status symlink.Status
		}{relPath, status})

		switch status {
		case symlink.StatusLinked:
			linked++
		case symlink.StatusNotLinked:
			notLinked++
		case symlink.StatusBroken:
			broken++
		case symlink.StatusIncorrect:
			incorrect++
		case symlink.StatusNotExist:
			notLinked++
		}
	}

	// Print entry header
	totalFiles := len(entry.Files)
	statusSummary := formatStatusSummary(linked, notLinked, broken, incorrect, totalFiles)
	fmt.Printf("%s (%s)\n", name, entry.Root)
	fmt.Printf("  %d file(s) - %s\n", totalFiles, statusSummary)

	// Print file details if requested
	if showDetails {
		for _, fs := range fileStatuses {
			statusIcon := statusIcon(fs.status)
			fmt.Printf("    %s %s\n", statusIcon, fs.file)
		}
	}

	fmt.Println()
}

// formatStatusSummary creates a summary string of file statuses.
func formatStatusSummary(linked, notLinked, broken, incorrect, total int) string {
	if linked == total {
		return "all linked"
	}
	if notLinked == total {
		return "not linked"
	}

	parts := []string{}
	if linked > 0 {
		parts = append(parts, fmt.Sprintf("%d linked", linked))
	}
	if notLinked > 0 {
		parts = append(parts, fmt.Sprintf("%d not linked", notLinked))
	}
	if broken > 0 {
		parts = append(parts, fmt.Sprintf("%d broken", broken))
	}
	if incorrect > 0 {
		parts = append(parts, fmt.Sprintf("%d incorrect", incorrect))
	}

	return strings.Join(parts, ", ")
}

// statusIcon returns a visual indicator for a file's status.
func statusIcon(status symlink.Status) string {
	switch status {
	case symlink.StatusLinked:
		return "[ok]     "
	case symlink.StatusNotLinked:
		return "[not lnk]"
	case symlink.StatusNotExist:
		return "[missing]"
	case symlink.StatusBroken:
		return "[broken] "
	case symlink.StatusIncorrect:
		return "[wrong]  "
	default:
		return "[?]      "
	}
}
