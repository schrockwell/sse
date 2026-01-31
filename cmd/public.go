package cmd

import (
	"fmt"

	"github.com/schrockwell/sse/internal/keyfile"
	"github.com/spf13/cobra"
)

var publicCmd = &cobra.Command{
	Use:   "public",
	Short: "Print the public key from master.key",
	RunE: func(cmd *cobra.Command, args []string) error {
		recipient, err := keyfile.LoadRecipient()
		if err != nil {
			return err
		}
		fmt.Println(recipient.String())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(publicCmd)
}
