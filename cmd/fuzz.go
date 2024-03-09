package cmd

import (

	//"time"

	"github.com/momaek/authy/service"
	"github.com/spf13/cobra"
)

// 'search' represents the fuzzy search
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Fuzzy search your otp tokens(case-insensitive)",
	Long: `Fuzzy search your otp tokens(case-insensitive)

First time(or after clean cache) , need your authy main password`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			isAlfred = false
			keyword  = ""
		)
		if alfredCount != nil && *alfredCount > 0 {
			isAlfred = true
		}

		if len(args) > 0 {
			keyword = args[0]
		}

		service.NewSearcher(keyword, isAlfred).Search()

	},
}

var alfredCount *int

func init() {
	rootCmd.AddCommand(searchCmd)
	alfredCount = searchCmd.Flags().CountP("alfred", "a", "Specify Output Mode AlfredWorkflow")
}
