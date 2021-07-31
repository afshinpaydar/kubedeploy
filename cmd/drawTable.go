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
	"os"

	"github.com/olekukonko/tablewriter"
)

func serviceTableCreate(dataService [][]string) {
	serviceTable := tablewriter.NewWriter(os.Stdout)
	serviceTable.SetHeader([]string{"Service Name", "Current App Label", "Current Version Label"})
	serviceTable.SetHeaderColor(
		tablewriter.Colors{tablewriter.FgWhiteColor, tablewriter.Bold, tablewriter.BgBlackColor},
		tablewriter.Colors{tablewriter.FgWhiteColor, tablewriter.Bold, tablewriter.BgBlackColor},
		tablewriter.Colors{tablewriter.FgWhiteColor, tablewriter.BgRedColor, tablewriter.BgBlackColor})

	serviceTable.SetColumnColor(
		tablewriter.Colors{tablewriter.FgGreenColor},
		tablewriter.Colors{tablewriter.FgYellowColor},
		tablewriter.Colors{tablewriter.FgYellowColor})

	for _, v := range dataService {
		serviceTable.Append(v)
	}
	serviceTable.Render() // Send output
}

func deploymentTableCreate(dataDeployment [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Deployment Name", "Current App Label", "Current Version Label"})
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.FgWhiteColor, tablewriter.Bold, tablewriter.BgBlackColor},
		tablewriter.Colors{tablewriter.FgWhiteColor, tablewriter.Bold, tablewriter.BgBlackColor},
		tablewriter.Colors{tablewriter.FgWhiteColor, tablewriter.BgRedColor, tablewriter.BgBlackColor})

	table.SetColumnColor(
		tablewriter.Colors{tablewriter.FgGreenColor},
		tablewriter.Colors{tablewriter.FgYellowColor},
		tablewriter.Colors{tablewriter.FgYellowColor})

	for _, v := range dataDeployment {
		table.Append(v)
	}
	table.Render() // Send output
}
