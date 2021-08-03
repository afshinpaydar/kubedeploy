/*
Copyright © 2021 Afshin Paydar <afshinpaydar@gmail.com>

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
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Differentiate between current deployment and intended",
	Long: `This command shows what is difference between current deployment manifest in k8s
And the intended manifests to successfully release with this tool`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logger("Please pass SERVICENAME as arguments", Fatal)
		}
		show(args[0])
	},
}

func init() {
	rootCmd.AddCommand(showCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// showCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// showCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func show(appName string) {
	sName, sAppLabel, sVerLabel := findService(appName)
	bDeployName, bDeployAppLabel, bDeployVerLabel := findDeployment(appName, "blue")
	gDeployName, gDeployAppLabel, gDeployVerLabel := findDeployment(appName, "green")

	dataService := [][]string{
		[]string{sName,
			sAppLabel, sVerLabel},
	}
	dataDeployment := [][]string{
		[]string{bDeployName,
			bDeployAppLabel, bDeployVerLabel},
		[]string{gDeployName,
			gDeployAppLabel, gDeployVerLabel},
	}
	deploymentTableCreate(dataDeployment)
	serviceTableCreate(dataService)
}
