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
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yomorun/cli/pkg/ga"
	"github.com/yomorun/cli/pkg/log"
	"github.com/yomorun/cli/serverless"
)

// demoCmd represents the demo command
var demoCmd = &cobra.Command{
	Use:                "demo",
	Short:              "Run YoMo demos",
	Long:               "Run the demos of Streaming Serverless functions and low-latency Geo-distributed applications in YoMo.",
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			opts.Filename = args[0]
		}

		runDemo()
	},
}

func init() {
	rootCmd.AddCommand(demoCmd)

	demoCmd.Flags().StringVarP(&opts.Filename, "file", "f", "app.go", "The URL or local path of Stream function file")
}

// runDemo runs a YoMo demo.
func runDemo() {
	// ga
	ga.Send(GetVersion(), "demo", opts.Filename)

	// Serverless
	log.InfoStatusEvent(os.Stdout, "YoMo Stream Function file: %v", opts.Filename)
	if isUrl(opts.Filename) {
		// download file
		log.PendingStatusEvent(os.Stdout, "Downloading file...")
		tempFile, err := download(opts.Filename)
		if err != nil {
			log.FailureStatusEvent(os.Stdout, err.Error())
			return
		}
		opts.Filename = tempFile
	}

	// resolve serverless
	log.PendingStatusEvent(os.Stdout, "Create YoMo Stream Function instance...")

	// Connect the serverless to YoMo dev-server, it will automatically emit the mock data.
	opts.Host = "dev.yomo.run"
	opts.Port = 9201
	opts.Name = "basic"

	s, err := serverless.Create(&opts)
	if err != nil {
		log.FailureStatusEvent(os.Stdout, err.Error())
		return
	}

	// build
	log.PendingStatusEvent(os.Stdout, "YoMo Stream Function building...")
	if err := s.Build(true); err != nil {
		log.FailureStatusEvent(os.Stdout, err.Error())
		return
	}
	log.SuccessStatusEvent(os.Stdout, "Success! YoMo Stream Function build.")
	// run
	log.InfoStatusEvent(os.Stdout, "YoMo Stream Function is running...")
	if err := s.Run(verbose); err != nil {
		log.FailureStatusEvent(os.Stdout, err.Error())
		return
	}
}

// isUrl validates if the filename is a URL.
func isUrl(filename string) bool {
	u, err := url.Parse(filename)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// download the sfn file and save it in a local temp folder.
func download(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	// Create a temp folder.
	tempDir, err := ioutil.TempDir("", "yomo_")
	if err != nil {
		return "", err
	}

	// Create the file
	tempFile := filepath.Join(tempDir, "app.go")
	out, err := os.Create(tempFile)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return tempFile, err
}
