package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// refreshCmd represents the refresh command
var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh token cache",
	Long: `When you add a new token.

You can use this cmd to refresh local token cache`,
	Run: func(cmd *cobra.Command, args []string) {
		devInfo, err := LoadExistingDeviceInfo()
		if err != nil {
			log.Fatal("Load device info failed", err)
		}

		getTokensFromAuthyServer(&devInfo)
	},
}

func init() {
	rootCmd.AddCommand(refreshCmd)
}
