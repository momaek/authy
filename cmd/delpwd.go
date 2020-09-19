package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

// delpwdCmd represents the delpwd command
var delpwdCmd = &cobra.Command{
	Use:   "delpwd",
	Short: "Delete saved backup password",
	Long: `Delete save backup password
Another way to reset backup password`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.AddCommand(delpwdCmd)
}

func deleteDevInfoPassword() {
	devInfo, err := LoadExistingDeviceInfo()
	if err != nil {
		log.Fatal("Load device info failed", err)
	}

	devInfo.MainPassword = ""
	SaveDeviceInfo(devInfo)
	log.Println("Backup password delete successfully!")
}
