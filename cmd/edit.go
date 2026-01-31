package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"sort"

	"github.com/BurntSushi/toml"
	"github.com/schrockwell/sse/internal/keyfile"
	"github.com/schrockwell/sse/internal/secrets"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit env.toml",
	Long: `Decrypt all values in env.toml, open in your editor,
then re-encrypt when the editor closes.

Uses $EDITOR, $VISUAL, VS Code, or vim (in that order).

Examples:
  sse edit`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		identity, err := keyfile.LoadIdentity()
		if err != nil {
			return err
		}
		recipient := identity.Recipient()

		f, err := secrets.Load(secrets.DefaultFile)
		if err != nil {
			return err
		}

		// Decrypt all environments
		decryptedEnvs := make(map[string]map[string]string)
		for envName, env := range f.Environments {
			decrypted, err := secrets.DecryptEnvironment(env, identity)
			if err != nil {
				return fmt.Errorf("failed to decrypt %s: %w", envName, err)
			}
			decryptedEnvs[envName] = decrypted
		}

		// Create temp file with decrypted TOML
		tmpFile, err := os.CreateTemp("", "sss-edit-*.toml")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		// Write decrypted TOML
		envNames := make([]string, 0, len(decryptedEnvs))
		for name := range decryptedEnvs {
			envNames = append(envNames, name)
		}
		sort.Strings(envNames)

		for i, envName := range envNames {
			if i > 0 {
				fmt.Fprintln(tmpFile)
			}
			fmt.Fprintf(tmpFile, "[%s]\n", envName)

			env := decryptedEnvs[envName]
			keys := make([]string, 0, len(env))
			for k := range env {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, key := range keys {
				fmt.Fprintf(tmpFile, "%s = %q\n", key, env[key])
			}
		}
		tmpFile.Close()

		// Run editor: $EDITOR, $VISUAL, VS Code, or vim
		var editorCmd *exec.Cmd
		if editor := os.Getenv("EDITOR"); editor != "" {
			editorCmd = exec.Command(editor, tmpPath)
		} else if editor := os.Getenv("VISUAL"); editor != "" {
			editorCmd = exec.Command(editor, tmpPath)
		} else if _, err := exec.LookPath("code"); err == nil {
			editorCmd = exec.Command("code", "--wait", tmpPath)
		} else {
			editorCmd = exec.Command("vim", tmpPath)
		}
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr

		if err := editorCmd.Run(); err != nil {
			return fmt.Errorf("editor exited with error: %w", err)
		}

		// Read and parse edited TOML
		editedData, err := os.ReadFile(tmpPath)
		if err != nil {
			return fmt.Errorf("failed to read edited file: %w", err)
		}

		var editedEnvs map[string]map[string]string
		if err := toml.Unmarshal(editedData, &editedEnvs); err != nil {
			return fmt.Errorf("failed to parse edited TOML: %w", err)
		}

		// Encrypt all values
		encryptedEnvs := make(map[string]map[string]string)
		for envName, env := range editedEnvs {
			encrypted, err := secrets.EncryptEnvironment(env, recipient)
			if err != nil {
				return fmt.Errorf("failed to encrypt %s: %w", envName, err)
			}
			encryptedEnvs[envName] = encrypted
		}

		// Save
		f.Environments = encryptedEnvs
		if err := f.Save(secrets.DefaultFile); err != nil {
			return err
		}

		fmt.Println("Saved env.toml")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
