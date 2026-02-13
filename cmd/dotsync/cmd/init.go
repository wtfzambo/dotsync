package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wtfzambo/dotsync/internal/config"
	"github.com/wtfzambo/dotsync/internal/manifest"
	"github.com/wtfzambo/dotsync/internal/storage"
)

var initCmd = &cobra.Command{
	Use:   "init <provider>",
	Short: "Initialize dotsync with a cloud storage provider",
	Long: `Initialize dotsync by specifying which cloud storage provider to use.

Supported providers:
  gdrive   - Google Drive
  dropbox  - Dropbox
  icloud   - iCloud Drive

The command will attempt to auto-detect the storage location.
If not found, you'll be prompted to enter the path manually.

You can also specify an explicit path using the --path flag.`,
	Example: `  dotsync init gdrive
  dotsync init dropbox
  dotsync init --path ~/my-cloud-folder`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

var initPath string

func init() {
	initCmd.Flags().StringVarP(&initPath, "path", "p", "", "Explicit storage path (skips provider detection)")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Check if already initialized
	exists, err := config.Exists()
	if err != nil {
		return fmt.Errorf("checking config: %w", err)
	}
	if exists {
		// Prompt for reinitialize
		if !confirmReinit() {
			fmt.Println("Aborted.")
			return nil
		}
	}

	var storagePath string

	// If explicit path provided, use it
	if initPath != "" {
		storagePath = initPath
	} else {
		// Need a provider argument
		if len(args) == 0 {
			return fmt.Errorf("please specify a provider (gdrive, dropbox, icloud) or use --path")
		}

		provider := storage.ParseProvider(args[0])
		if provider == "" {
			return fmt.Errorf("unknown provider: %s. Supported: gdrive, dropbox, icloud", args[0])
		}

		// Try to detect the storage path
		storagePath = storage.DetectPath(provider)
		if storagePath == "" {
			// Prompt for manual entry
			var err error
			storagePath, err = promptForPath(provider)
			if err != nil {
				return err
			}
		} else {
			fmt.Printf("Found %s at: %s\n", provider.DisplayName(), storagePath)
		}
	}

	// Validate the path
	if err := storage.ValidatePath(storagePath); err != nil {
		return err
	}

	// Ensure dotsync directory exists
	_, err = storage.EnsureDotsyncDir(storagePath)
	if err != nil {
		return err
	}

	// Create manifest if it doesn't exist
	expandedPath := storage.ExpandPath(storagePath)
	if !manifest.Exists(expandedPath) {
		m := manifest.New()
		if err := m.Save(expandedPath); err != nil {
			return fmt.Errorf("creating manifest: %w", err)
		}
		fmt.Println("Created new manifest.")
	} else {
		fmt.Println("Using existing manifest.")
	}

	// Save config
	cfg := config.New(storagePath)
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("dotsync initialized! Storage: %s\n", storagePath)
	return nil
}

func confirmReinit() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("dotsync is already initialized. Reinitialize? [y/N] ")
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func promptForPath(provider storage.Provider) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s not found at known locations.\n", provider.DisplayName())
	fmt.Print("Enter path (or 'q' to quit): ")

	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading input: %w", err)
	}

	response = strings.TrimSpace(response)
	if response == "q" || response == "" {
		return "", fmt.Errorf("aborted")
	}

	return response, nil
}
