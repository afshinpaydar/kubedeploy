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
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// bluegreenCmd represents the bluegreen command
var bluegreenCmd = &cobra.Command{
	Use:   "bluegreen SERVICENAME NEWVERSION",
	Short: "blue/green deployment",
	Long: `**********************************************************************************************
| "bluegreen" helps you to implement blue/green deployment in your k8s cluster               |
| "bluegreen" expect two Deployments and one Service, that points to one of those            |
| in the active k8s cluster.                                                                 |
| the name of Deployments must ends with '-blue' and '-green' but Service name               |
| could be anything, and also how the Service exposed to outside of Kubernetes cluster.      |
**********************************************************************************************`,

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			logger("Please pass SERVICENAME and NEWVERSION as arguments", Fatal)
		}
		blueGreenDeploy(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(bluegreenCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// bluegreenCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bluegreenCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
