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
	"os"

	"github.com/spf13/cobra"
	"github.com/yomorun/cli/pkg/log"
	"github.com/yomorun/cli/serverless"
)

// devCmd represents the dev command
var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Dev a YoMo Serverless Function",
	Long:  "Dev a YoMo Serverless Function with mocking yomo-source data from YCloud.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			opts.Filename = args[0]
		}
		// Serverless
		log.InfoStatusEvent(os.Stdout, "YoMo serverless function file: %v", opts.Filename)
		// resolve serverless
		log.PendingStatusEvent(os.Stdout, "Create YoMo serverless instance...")

		// TODO: change the host and port to the remote dev-server's address.
		// Connect the serverless to YoMo dev-server, it will automatically emit the mock data.
		opts.Host = "localhost"
		opts.Port = 9000
		opts.Name = "YoMo Stream Function"

		s, err := serverless.Create(&opts)
		if err != nil {
			log.FailureStatusEvent(os.Stdout, err.Error())
			return
		}

		// build
		log.PendingStatusEvent(os.Stdout, "YoMo serverless function building...")
		if err := s.Build(true); err != nil {
			log.FailureStatusEvent(os.Stdout, err.Error())
			return
		}
		log.SuccessStatusEvent(os.Stdout, "Success! YoMo serverless function build.")
		// run
		log.InfoStatusEvent(os.Stdout, "YoMo serverless function is running...")
		if err := s.Run(); err != nil {
			log.FailureStatusEvent(os.Stdout, err.Error())
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(devCmd)

	devCmd.Flags().StringVarP(&opts.Filename, "file-name", "f", "app.go", "Serverless function file")
}
