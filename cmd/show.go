package cmd

import (
	"fmt"
	"sort"

	"github.com/schrockwell/sse/internal/keyfile"
	"github.com/schrockwell/sse/internal/secrets"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Print decrypted env.toml",
	Long: `Print the entire env.toml with all values decrypted.

Examples:
  sse show`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		identity, err := keyfile.ReadIdentity(keyfile.DefaultKeyFile)
		if err != nil {
			return fmt.Errorf("failed to read key file: %w", err)
		}

		f, err := secrets.Load(secrets.DefaultFile)
		if err != nil {
			return err
		}

		// Sort environment names
		envNames := make([]string, 0, len(f.Environments))
		for name := range f.Environments {
			envNames = append(envNames, name)
		}
		sort.Strings(envNames)

		for i, envName := range envNames {
			if i > 0 {
				fmt.Println()
			}
			fmt.Printf("[%s]\n", envName)

			env := f.Environments[envName]
			decrypted, err := secrets.DecryptEnvironment(env, identity)
			if err != nil {
				return fmt.Errorf("failed to decrypt %s: %w", envName, err)
			}

			// Sort keys
			keys := make([]string, 0, len(decrypted))
			for k := range decrypted {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, key := range keys {
				fmt.Printf("%s = %q\n", key, decrypted[key])
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
