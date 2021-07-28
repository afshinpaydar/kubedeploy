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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// bluegreenCmd represents the bluegreen command
var bluegreenCmd = &cobra.Command{
	Use:   "bluegreen",
	Short: "Simple blue/green deployment plugin for KUBECTL",
	Long: `"kube-deploy bluegreen" helps you to implement blue/green deployment in your k8s cluster
"kubectl-deploy bluegreen" expect two Deployments and one Service, that points to one of those in the active k8s cluster
the name of Deployments and Service doesn’t matter and could be anything,
and also how the Service exposed to outside of Kubernetes cluster.`,

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			fmt.Println("FATAL: Please pass APP_NAME and VERSION as arguments")
			os.Exit(1)
		} else {
			blueGreenDeploy(args[0], args[1])
		}
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
