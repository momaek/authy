/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// fuzzCmd represents the fuzz command
var fuzzCmd = &cobra.Command{
	Use:   "fuzz",
	Short: "fuzzy search your otp tokens(case-insensitive)",
	Long: `Fuzzy search your otp tokens(case-insensitive)

First time(or after clean cache) , need your authy main password`,
	Run: func(cmd *cobra.Command, args []string) {
		fuzzySearch(args)
	},
}

func init() {
	rootCmd.AddCommand(fuzzCmd)
}

func fuzzySearch(args []string) {
	var keyword string

	oc := bufio.NewScanner(os.Stdin)
	if len(args) == 0 {
		fmt.Print("Please input search keyword: ")
		if !oc.Scan() {
			log.Println("Please provide a keyword, e.g. google")
			return
		}

		keyword = strings.TrimSpace(oc.Text())
	}

	devInfo, err := LoadExistingDeviceInfo()
	if err != nil {
		if os.IsNotExist(err) {
			devInfo, err = newRegistrationDevice()
			if err != nil {
				return
			}
		} else {
			log.Println("load device info failed", err)
			return
		}
	}

	if len(devInfo.MainPassword) == 0 {
		fmt.Print("\nPlease input Authy main password: ")
		if !oc.Scan() {
			log.Println("Please provide a phone country code, e.g. 86")
			return
		}

		devInfo.MainPassword = strings.TrimSpace(oc.Text())
	}
}

func loadCachedTokens() {}
