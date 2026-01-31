package cmd

import (
	"fmt"

	"github.com/schrockwell/sse/internal/keyfile"
	"github.com/spf13/cobra"
)

var privateCmd = &cobra.Command{
	Use:   "private",
	Short: "Print the private key from master.key",
	RunE: func(cmd *cobra.Command, args []string) error {
		identity, err := keyfile.LoadIdentity()
		if err != nil {
			return err
		}
		fmt.Println(identity.String())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(privateCmd)
}
