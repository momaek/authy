package cmd

import (
	"github.com/momaek/authy/service"
	"github.com/spf13/cobra"
)

// accountCmd represents the account command
var (
	countrycode, mobile, password string

	accountCmd = &cobra.Command{
		Use:   "account",
		Short: "Authy account info or register device",
		Long: `Register device or show registered account info. 

Can specify country code, mobile number and authy main password.
If not provided, will get from command line stdin`,
		Run: func(cmd *cobra.Command, args []string) {
			service.RegisterOrGetDeviceInfo()
		},
	}
)

func init() {
	rootCmd.AddCommand(accountCmd)

	accountCmd.Flags().StringVarP(&countrycode, "countrycode", "c", "", "phone number country code (e.g. 1 for United States), digitals only")
	accountCmd.Flags().StringVarP(&mobile, "mobilenumber", "m", "", "phone number, digitals only")
	accountCmd.Flags().StringVarP(&password, "password", "p", "", "authy main password")
}
