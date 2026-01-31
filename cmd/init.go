package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/schrockwell/sse/internal/keyfile"
	"github.com/schrockwell/sse/internal/secrets"
	"github.com/spf13/cobra"
)

var initForce bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new project",
	Long: `Initialize a new project by creating:
- master.key: age keypair for encryption/decryption
- env.toml: secrets file with development and production sections

Automatically adds master.key to .gitignore if it exists.
The env.toml file is safe to commit (values are encrypted).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Generate master.key
		if err := keyfile.Generate(keyfile.DefaultKeyFile, initForce); err != nil {
			return err
		}
		fmt.Printf("Created %s\n", keyfile.DefaultKeyFile)

		// Create env.toml if it doesn't exist
		if _, err := os.Stat(secrets.DefaultFile); os.IsNotExist(err) || initForce {
			if err := secrets.CreateDefault(secrets.DefaultFile); err != nil {
				return fmt.Errorf("failed to create env.toml: %w", err)
			}
			fmt.Printf("Created %s\n", secrets.DefaultFile)
		} else {
			fmt.Printf("Skipped %s (already exists)\n", secrets.DefaultFile)
		}

		// Add to .gitignore if it exists
		if err := addToGitignore("/master.key"); err != nil {
			fmt.Printf("Warning: %v\n", err)
		}

		return nil
	},
}

func addToGitignore(entry string) error {
	const gitignore = ".gitignore"

	// Check if .gitignore exists
	if _, err := os.Stat(gitignore); os.IsNotExist(err) {
		return nil // No .gitignore, nothing to do
	}

	// Read and check if entry already exists
	file, err := os.Open(gitignore)
	if err != nil {
		return fmt.Errorf("failed to read .gitignore: %w", err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == entry {
			file.Close()
			return nil // Already in .gitignore
		}
	}
	file.Close()

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read .gitignore: %w", err)
	}

	// Append entry
	f, err := os.OpenFile(gitignore, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open .gitignore: %w", err)
	}
	defer f.Close()

	// Add newline if file doesn't end with one
	info, _ := f.Stat()
	if info.Size() > 0 {
		// Read last byte to check for newline
		buf := make([]byte, 1)
		f.Seek(-1, 2)
		f.Read(buf)
		f.Seek(0, 2) // Back to end
		if buf[0] != '\n' {
			f.WriteString("\n")
		}
	}

	if _, err := fmt.Fprintln(f, entry); err != nil {
		return fmt.Errorf("failed to write to .gitignore: %w", err)
	}

	fmt.Printf("Added %s to .gitignore\n", entry)
	return nil
}

func init() {
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Overwrite existing files")
	rootCmd.AddCommand(initCmd)
}
