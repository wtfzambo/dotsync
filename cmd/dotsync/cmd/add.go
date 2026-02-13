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

var addCmd = &cobra.Command{
	Use:   "add <path>",
	Short: "Add a file to be synced",
	Long: `Add a file to be tracked and synced via cloud storage.

The file will be moved to cloud storage and a symlink will be created
at the original location. The entry name is inferred from the path
(e.g., ~/.config/opencode/config.json becomes entry "opencode").

Use --name to specify a custom entry name.`,
	Example: `  dotsync add ~/.config/opencode/config.json
  dotsync add ~/.zshrc --name shell
  dotsync add ~/.aws/credentials`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

var addName string

func init() {
	addCmd.Flags().StringVarP(&addName, "name", "n", "", "Custom entry name (inferred from path if not specified)")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	inputPath := args[0]

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

	// 2. Convert to absolute path
	absPath, err := pathutil.AbsolutePath(inputPath)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	// 3. Validate the file
	if err := pathutil.ValidateForAdd(absPath); err != nil {
		if valErr, ok := err.(pathutil.ValidationError); ok {
			if valErr.IsWarn {
				// Warning - ask for confirmation
				fmt.Printf("Warning: %s\n", valErr.Message)
				if !confirmPrompt("Continue anyway?") {
					return fmt.Errorf("aborted")
				}
			} else {
				// Fatal error
				return fmt.Errorf("%s", valErr.Message)
			}
		} else {
			return err
		}
	}

	// 4. Load or create manifest
	m, err := manifest.Load(storagePath)
	if err != nil {
		if strings.Contains(err.Error(), "manifest not found") {
			m = manifest.New()
		} else {
			return fmt.Errorf("loading manifest: %w", err)
		}
	}

	// 5. Check if already tracked
	if entryName := pathutil.IsAlreadyTracked(absPath, m); entryName != "" {
		fmt.Printf("Already tracked in entry '%s'\n", entryName)
		return nil
	}

	// 6. Infer entry name and root
	var entryName, root, relPath string

	if addName != "" {
		// User specified name
		entryName = addName

		// Check for conflict with existing entry
		conflict, err := pathutil.CheckEntryConflict(absPath, addName, m)
		if err != nil {
			return fmt.Errorf("checking conflicts: %w", err)
		}
		if conflict != "" && conflict != addName {
			return fmt.Errorf("file is under entry '%s', cannot add to '%s'", conflict, addName)
		}

		// If entry exists, use its root
		if existing := m.GetEntry(addName); existing != nil {
			root = existing.Root
			expandedRoot := pathutil.ExpandHome(root)
			if !strings.HasPrefix(absPath, expandedRoot) {
				return fmt.Errorf("file is not under existing entry root: %s", root)
			}
			relPath, _ = filepath.Rel(expandedRoot, absPath)
		} else {
			// Try to infer root from path, or use parent directory
			inferred := pathutil.InferFromPath(absPath)
			if inferred != nil {
				root = inferred.Root
				relPath = inferred.RelPath
			} else {
				// Default: use parent directory as root
				root = pathutil.ContractHome(filepath.Dir(absPath))
				relPath = filepath.Base(absPath)
			}
		}
	} else {
		// Infer from path
		inferred := pathutil.InferFromPath(absPath)
		if inferred != nil {
			entryName = inferred.Name
			root = inferred.Root
			relPath = inferred.RelPath

			// Check if this entry already exists with a different root
			if existing := m.GetEntry(entryName); existing != nil {
				if existing.Root != root {
					return fmt.Errorf("entry '%s' exists with different root: %s (expected %s). Use --name to specify a different entry", entryName, existing.Root, root)
				}
				root = existing.Root
			}
		} else {
			// Cannot infer - prompt for name
			fmt.Printf("Cannot infer entry name from path: %s\n", absPath)
			entryName = promptForName()
			if entryName == "" {
				return fmt.Errorf("entry name is required")
			}

			// Use parent directory as root
			root = pathutil.ContractHome(filepath.Dir(absPath))
			relPath = filepath.Base(absPath)
		}

		// Check for conflict
		conflict, err := pathutil.CheckEntryConflict(absPath, "", m)
		if err != nil {
			return fmt.Errorf("checking conflicts: %w", err)
		}
		if conflict != "" && conflict != entryName {
			return fmt.Errorf("file is under entry '%s'. Use 'dotsync add %s --name %s' to add to that entry", conflict, inputPath, conflict)
		}
	}

	// 7. Calculate destination path in cloud storage
	// Structure: <storage>/dotsync/<name>/<relPath>
	destPath := filepath.Join(storagePath, "dotsync", entryName, relPath)

	// Check if destination already exists
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("file already exists in cloud storage: %s\nIf syncing from another machine, use 'dotsync link' instead", destPath)
	}

	// 8. Create backup
	bk, err := backup.Create(absPath)
	if err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}

	// 9. Move file to cloud storage
	fmt.Printf("Moving to cloud storage: %s -> %s\n", pathutil.ContractHome(absPath), pathutil.ContractHome(destPath))
	if err := symlink.MoveFile(absPath, destPath); err != nil {
		bk.Restore()
		return fmt.Errorf("moving file: %w", err)
	}

	// 10. Create symlink at original location
	fmt.Printf("Creating symlink: %s -> %s\n", pathutil.ContractHome(absPath), pathutil.ContractHome(destPath))
	if err := symlink.Create(absPath, destPath); err != nil {
		// Rollback: move file back
		symlink.MoveFile(destPath, absPath)
		bk.Restore()
		return fmt.Errorf("creating symlink: %w", err)
	}

	// 11. Update manifest
	m.AddFile(entryName, root, relPath)
	if err := m.Save(storagePath); err != nil {
		// Rollback: remove symlink, move file back
		symlink.Remove(absPath)
		symlink.MoveFile(destPath, absPath)
		bk.Restore()
		return fmt.Errorf("saving manifest: %w", err)
	}

	// 12. Cleanup backup
	bk.Cleanup()

	fmt.Printf("Added '%s' to entry '%s'\n", relPath, entryName)
	return nil
}

// confirmPrompt asks the user for yes/no confirmation.
func confirmPrompt(question string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", question)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// promptForName asks the user for an entry name.
func promptForName() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Entry name: ")
	name, _ := reader.ReadString('\n')
	return strings.TrimSpace(name)
}
