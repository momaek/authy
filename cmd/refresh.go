package cmd

import (
	"github.com/momaek/authy/service"
	"github.com/spf13/cobra"
)

// refreshCmd represents the refresh command
var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh token cache",
	Long: `When you add a new token.

You can use this cmd to refresh local token cache`,
	Run: func(cmd *cobra.Command, args []string) {
		device := service.NewDevice(service.NewDeviceConfig{})
		device.LoadTokenFromAuthyServer()
	},
}

func init() {
	rootCmd.AddCommand(refreshCmd)
}
