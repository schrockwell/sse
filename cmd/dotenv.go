package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/schrockwell/sse/internal/keyfile"
	"github.com/schrockwell/sse/internal/secrets"
	"github.com/spf13/cobra"
)

var dotenvCmd = &cobra.Command{
	Use:   "dotenv [environment]",
	Short: "Print variables in dotenv format",
	Long: `Output decrypted variables in dotenv format (KEY="value").

Examples:
  sse dotenv                     # development (default)
  sse dotenv production          # production
  sse dotenv > .env              # write to .env file`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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

		keys := make([]string, 0, len(decrypted))
		for k := range decrypted {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			// Escape backslashes, double quotes, and newlines for dotenv format
			escaped := decrypted[key]
			escaped = strings.ReplaceAll(escaped, `\`, `\\`)
			escaped = strings.ReplaceAll(escaped, `"`, `\"`)
			escaped = strings.ReplaceAll(escaped, "\n", `\n`)
			fmt.Printf("%s=\"%s\"\n", key, escaped)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(dotenvCmd)
}
