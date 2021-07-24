package cmd

import (
	"github.com/momaek/authy/service"
	"github.com/spf13/cobra"
)

// delpwdCmd represents the delpwd command
var delpwdCmd = &cobra.Command{
	Use:   "delpwd",
	Short: "Delete saved backup password",
	Long: `Delete save backup password
Another way to reset backup password`,
	Run: func(cmd *cobra.Command, args []string) {
		d := service.NewDevice(service.NewDeviceConfig{})
		d.DeleteMainPassword()
	},
}

func init() {
	rootCmd.AddCommand(delpwdCmd)
}
