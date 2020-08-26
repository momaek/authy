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
