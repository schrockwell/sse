package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version = "0.1.3"

var rootCmd = &cobra.Command{
	Use:     "sse",
	Version: Version,
	Short:   "Stupidly Simple Environments - encrypted environment management",
	Long: `Stupidly Simple Environments (sse) manages encrypted environment variables
for small projects using age encryption.

Files:
  master.key  - age keypair (add to .gitignore)
  env.toml    - environment file with encrypted values (safe to commit)

The env.toml file contains sections for each environment:
  [development]
  API_KEY = "ENC[...]"

  [production]
  API_KEY = "ENC[...]"

Keys are human-readable, only values are encrypted.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
