package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/schrockwell/sse/internal/keyfile"
	"github.com/schrockwell/sse/internal/secrets"
	"github.com/spf13/cobra"
)

var withEnv string

var withCmd = &cobra.Command{
	Use:   "with [environment] -- <command> [args...]",
	Short: "Run a command with decrypted environment",
	Long: `Decrypt secrets for the specified environment and run a command
with those environment variables. Decrypted values are never written to disk.

Examples:
  sse with -- env                     # use development (default)
  sse with -- npm start               # run npm with secrets
  sse with production -- ./deploy.sh  # use production secrets`,
	DisableFlagsInUseLine: true,
	Args:                  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		envName := secrets.DefaultEnvironment
		cmdArgs := args

		// Check if first arg looks like an environment name (not a command)
		// by seeing if it exists in env.toml
		if len(args) > 1 {
			f, err := secrets.Load(secrets.DefaultFile)
			if err != nil {
				return err
			}
			if _, exists := f.Environments[args[0]]; exists {
				envName = args[0]
				cmdArgs = args[1:]
			}
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

		// Build environment: current env + secrets
		environ := os.Environ()
		environ = append(environ, secrets.ToEnvList(decrypted)...)

		// Find the command
		binary, err := exec.LookPath(cmdArgs[0])
		if err != nil {
			return fmt.Errorf("command not found: %s", cmdArgs[0])
		}

		// Replace current process with the command
		return syscall.Exec(binary, cmdArgs, environ)
	},
}

func init() {
	rootCmd.AddCommand(withCmd)
}
