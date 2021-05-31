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

	"github.com/spf13/cobra"

	"github.com/yomorun/cli/pkg/log"
	"github.com/yomorun/cli/serverless"
)

var (
	config string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run a YoMo Serverless instance",
	Long:  "Run a YoMo Serverless instance",
	Run: func(cmd *cobra.Command, args []string) {
		if config == "" {
			log.FailureStatusEvent(os.Stdout, "Please input the file name of workflow config")
			return
		}
		conf, err := serverless.ParseConfig(config)
		if err != nil {
			log.FailureStatusEvent(os.Stdout, err.Error())
			return
		}
		printZipperConf(conf)

		endpoint := fmt.Sprintf("%s:%d", conf.Host, conf.Port)

		log.InfoStatusEvent(os.Stdout, "Running YoMo Serverless...")
		err = serverless.Start(endpoint, serverless.NewQuicHandler(conf))
		if err != nil {
			log.FailureStatusEvent(os.Stdout, err.Error())
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVarP(&config, "config", "c", "workflow.yaml", "Workflow config file")
	serveCmd.MarkFlagRequired("config")
}

func printZipperConf(wfConf *serverless.WorkflowConfig) {
	log.InfoStatusEvent(os.Stdout, "Found %d flows in zipper config", len(wfConf.Flows))
	for i, flow := range wfConf.Flows {
		log.InfoStatusEvent(os.Stdout, "Flow %d: %s", i+1, flow.Name)
	}

	log.InfoStatusEvent(os.Stdout, "Found %d sinks in zipper config", len(wfConf.Sinks))
	for i, sink := range wfConf.Sinks {
		log.InfoStatusEvent(os.Stdout, "Sink %d: %s", i+1, sink.Name)
	}
}
