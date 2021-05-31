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
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/yomorun/cli/pkg/log"
	"github.com/yomorun/cli/serverless"
	_ "github.com/yomorun/cli/serverless/golang"
)

var (
	url string
)

const (
	runtimeWaitTimeoutInSeconds = 60
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a YoMo Serverless Function",
	Long:  "Run a YoMo Serverless Function",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			opts.Filename = args[0]
		}
		// os signal
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		// Serverless
		log.InfoStatusEvent(os.Stdout, "YoMo serverless function file: %v", opts.Filename)
		// resolve serverless
		log.PendingStatusEvent(os.Stdout, "Create YoMo serverless instance...")
		if err := parseURL(url, &opts); err != nil {
			log.FailureStatusEvent(os.Stdout, err.Error())
			return
		}
		s, err := serverless.Create(&opts)
		if err != nil {
			log.FailureStatusEvent(os.Stdout, err.Error())
			return
		}
		log.InfoStatusEvent(os.Stdout,
			"Starting YoMo serverless instance with Name: %s. Host: %s. Port: %d.",
			opts.Name,
			opts.Host,
			opts.Port,
		)
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
		// Exit
		<-sigCh
		log.WarningStatusEvent(os.Stdout, "Terminated signal received: shutting down")
		log.InfoStatusEvent(os.Stdout, "Exited YoMo serverless instance.")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&opts.Filename, "file-name", "f", "app.go", "Serverless function file")
	// runCmd.Flags().StringVarP(&opts.Lang, "lang", "l", "go", "source language")
	runCmd.Flags().StringVarP(&url, "url", "u", "localhost:9000", "zipper server endpoint addr")
	runCmd.Flags().StringVarP(&opts.Name, "name", "n", "", "yomo serverless app name (required). It should match the specific service name in zipper config (workflow.yaml)")
	runCmd.MarkFlagRequired("name")

}
