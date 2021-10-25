/*
Copyright © 2021 CELLA, Inc.

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

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/yomorun/cli/pkg/ga"
)

// demoListCmd represents the list sub command in demo.
var demoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the examples in YoMo.",
	Long:  "List the Streaming Serverless and Geo-Distributed examples provided by YoMo.",
	Run: func(cmd *cobra.Command, args []string) {
		selectDemo()
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		ga.Send(GetVersion(), "demo", "list")
	},
}

func init() {
	demoCmd.AddCommand(demoListCmd)
}

type promptContent struct {
	errorMsg string
	label    string
}

// selectDemo selects a demo and run it in local.
func selectDemo() {
	selectDemoContent := promptContent{
		"Please select a demo.",
		"Select a demo which you'd like to run",
	}
	index := promptGetDemoSelect(selectDemoContent)
	if index < 0 {
		fmt.Println("Invalid index", index)
		os.Exit(1)
	}

	switch index {
	case 0:
		opts.Filename = "https://play.yomo.run/static/demosfn/basic.go"
		runDemo()
	default:
		fmt.Println("TODO: add this demo.")
	}

}

func promptGetDemoSelect(pc promptContent) int {
	items := []string{"Streaming Serverless", "Customized yomo-source", "Geo-Distributed Streaming", "IoT Streaming"}
	index := -1
	var result string
	var err error

	for index < 0 {
		prompt := promptui.Select{
			Label: pc.label,
			Items: items,
		}

		index, result, err = prompt.Run()
	}

	if err != nil {
		fmt.Println("Prompt failed", err)
		os.Exit(1)
	}

	fmt.Println("Select the demo:", result)

	return index
}
