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
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/yomorun/cli/pkg/log"
	"github.com/yomorun/cli/pkg/standalone"
)

var (
	opts standalone.RunOptions
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
		if len(args) == 0 {
			log.FailureStatusEvent(os.Stdout, "no application command found")
			return
		}

		opts.Arguments = args
		output, err := standalone.Run(&opts)
		if err != nil {
			log.FailureStatusEvent(os.Stdout, err.Error())
			return
		}

		// os signal
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		// yomo
		// daprRunning := make(chan bool, 1)
		appRunning := make(chan bool, 1)

		go func() {
			if output.AppCMD == nil {
				appRunning <- true
				return
			}

			stdErrPipe, pipeErr := output.AppCMD.StderrPipe()
			if pipeErr != nil {
				log.FailureStatusEvent(os.Stdout, fmt.Sprintf("Error creating stderr for App: %s", err.Error()))
				appRunning <- false
				return
			}

			stdOutPipe, pipeErr := output.AppCMD.StdoutPipe()
			if pipeErr != nil {
				log.FailureStatusEvent(os.Stdout, fmt.Sprintf("Error creating stdout for App: %s", err.Error()))
				appRunning <- false
				return
			}

			errScanner := bufio.NewScanner(stdErrPipe)
			outScanner := bufio.NewScanner(stdOutPipe)
			go func() {
				for errScanner.Scan() {
					fmt.Println(log.Blue(fmt.Sprintf("== YoMo APP == %s", errScanner.Text())))
				}
			}()

			go func() {
				for outScanner.Scan() {
					fmt.Println(log.Blue(fmt.Sprintf("== YoMo APP == %s", outScanner.Text())))
				}
			}()

			err = output.AppCMD.Start()
			if err != nil {
				log.FailureStatusEvent(os.Stdout, err.Error())
				appRunning <- false
				return
			}

			go func() {
				appErr := output.AppCMD.Wait()

				if appErr != nil {
					log.FailureStatusEvent(os.Stdout, "The App process exited with error code: %s", appErr.Error())
				} else {
					log.SuccessStatusEvent(os.Stdout, "Exited App successfully")
				}
				sigCh <- os.Interrupt
			}()

			appRunning <- true
		}()

		appRunStatus := <-appRunning
		if !appRunStatus {
			// Start App failed, try to stop Dapr and exit.
			// err = output.DaprCMD.Process.Kill()
			// if err != nil {
			// 	log.FailureStatusEvent(os.Stdout, fmt.Sprintf("Start App failed, try to stop Dapr Error: %s", err))
			// } else {
			// 	log.SuccessStatusEvent(os.Stdout, "Start App failed, try to stop Dapr successfully")
			// }
			os.Exit(1)
		}

		// Metadata API is only available if app has started listening to port, so wait for app to start before calling metadata API.
		// err = metadata.Put(output.DaprHTTPPort, "cliPID", strconv.Itoa(os.Getpid()))
		// if err != nil {
		// 	log.WarningStatusEvent(os.Stdout, "Could not update sidecar metadata for cliPID: %s", err.Error())
		// }

		// if output.AppCMD != nil {
		// 	appCommand := strings.Join(args, " ")
		// 	log.InfoStatusEvent(os.Stdout, fmt.Sprintf("Updating metadata for app command: %s", appCommand))
		// 	err = metadata.Put(output.DaprHTTPPort, "appCommand", appCommand)
		// 	if err != nil {
		// 		log.WarningStatusEvent(os.Stdout, "Could not update sidecar metadata for appCommand: %s", err.Error())
		// 	} else {
		// 		log.SuccessStatusEvent(os.Stdout, "You're up and running! Both Dapr and your app logs will appear here.\n")
		// 	}
		// } else {
		// 	log.SuccessStatusEvent(os.Stdout, "You're up and running! Dapr logs will appear here.\n")
		// }

		<-sigCh
		log.InfoStatusEvent(os.Stdout, "\nterminated signal received: shutting down")

		// if output.DaprCMD.ProcessState == nil || !output.DaprCMD.ProcessState.Exited() {
		// 	err = output.DaprCMD.Process.Kill()
		// 	if err != nil {
		// 		log.FailureStatusEvent(os.Stdout, fmt.Sprintf("Error exiting Dapr: %s", err))
		// 	} else {
		// 		log.SuccessStatusEvent(os.Stdout, "Exited Dapr successfully")
		// 	}
		// }

		if output.AppCMD != nil && (output.AppCMD.ProcessState == nil || !output.AppCMD.ProcessState.Exited()) {
			err = output.AppCMD.Process.Kill()
			if err != nil {
				log.FailureStatusEvent(os.Stdout, fmt.Sprintf("Error exiting App: %s", err))
			} else {
				log.SuccessStatusEvent(os.Stdout, "Exited App successfully")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&opts.Filename, "file-name", "f", "app.go", "Serverless function file")
	runCmd.Flags().StringVarP(&opts.Url, "url", "u", "localhost:9000", "zipper server endpoint addr")
	runCmd.Flags().StringVarP(&opts.Name, "name", "n", "", "yomo serverless app name (required). It should match the specific service name in zipper config (workflow.yaml)")
	runCmd.MarkFlagRequired("name")

}
