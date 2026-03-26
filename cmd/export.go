package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/schrockwell/sse/internal/keyfile"
	"github.com/schrockwell/sse/internal/secrets"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export [environment]",
	Short: "Print variables in export format for shell eval",
	Long: `Output export statements for the specified environment.
Use with eval to load into your current shell.

Examples:
  eval "$(sse export)"             # load development (default)
  eval "$(sse export production)"  # load production`,
	Args: cobra.MaximumNArgs(1),
	RunE: runExport,
}

var loadCmd = &cobra.Command{
	Use:    "load [environment]",
	Short:  "Print variables in export format for shell eval",
	Hidden: true,
	Args:   cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(os.Stderr, "Warning: \"sse load\" is deprecated, use \"sse export\" instead")
		return runExport(cmd, args)
	},
}

func runExport(cmd *cobra.Command, args []string) error {
	envName := secrets.DefaultEnvironment
	if len(args) > 0 {
		envName = args[0]
	}

	identity, err := keyfile.LoadIdentity()
	if err != nil {
		return err
	}

	f, err := secrets.Load(secrets.DefaultFile)
	if err != nil {
		return err
	}

	env, err := f.GetEnvironment(envName)
	if err != nil {
		return err
	}

	decrypted, err := secrets.DecryptEnvironment(env, identity)
	if err != nil {
		return fmt.Errorf("failed to decrypt: %w", err)
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(decrypted))
	for k := range decrypted {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		// Use single quotes to prevent shell expansion, escape embedded single quotes
		escaped := strings.ReplaceAll(decrypted[key], "'", "'\"'\"'")
		fmt.Printf("export %s='%s'\n", key, escaped)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(loadCmd)
}
