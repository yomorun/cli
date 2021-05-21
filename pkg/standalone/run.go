package standalone

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	// "github.com/dapr/dapr/pkg/components"
	// modes "github.com/dapr/dapr/pkg/config/modes"
)

const sentryDefaultAddress = "localhost:50001"

type baseOptions struct {
	// Filename is the name of Serverless function file (default is app.go).
	Filename string
}

// RunOptions are the options for run command.
type RunOptions struct {
	baseOptions
	// Port is the port number of UDP host for Serverless function (default is 4242).
	Url       string
	Name      string
	Arguments []string
}

// RunConfig represents the application configuration parameters.
type RunConfig struct {
	AppID              string
	AppPort            int
	HTTPPort           int
	GRPCPort           int
	ConfigFile         string
	Protocol           string
	Arguments          []string
	EnableProfiling    bool
	ProfilePort        int
	LogLevel           string
	MaxConcurrency     int
	PlacementHost      string
	ComponentsPath     string
	AppSSL             bool
	MetricsPort        int
	MaxRequestBodySize int
}

// RunOutput represents the run output.
type RunOutput struct {
	RootCMD      *exec.Cmd
	DaprHTTPPort int
	DaprGRPCPort int
	AppID        string
	AppCMD       *exec.Cmd
}

// func getDaprCommand(appID string, daprHTTPPort int, daprGRPCPort int, appPort int, configFile, protocol string, enableProfiling bool, profilePort int, logLevel string, maxConcurrency int, placementHost string, componentsPath string, appSSL bool, metricsPort int, requestBodySize int) (*exec.Cmd, int, int, int, error) {
// 	if daprHTTPPort < 0 {
// 		port, err := freeport.GetFreePort()
// 		if err != nil {
// 			return nil, -1, -1, -1, err
// 		}

// 		daprHTTPPort = port
// 	}

// 	if daprGRPCPort < 0 {
// 		grpcPort, err := freeport.GetFreePort()
// 		if err != nil {
// 			return nil, -1, -1, -1, err
// 		}

// 		daprGRPCPort = grpcPort
// 	}

// 	if metricsPort < 0 {
// 		var err error
// 		metricsPort, err = freeport.GetFreePort()
// 		if err != nil {
// 			return nil, -1, -1, -1, err
// 		}
// 	}

// 	if maxConcurrency < 1 {
// 		maxConcurrency = -1
// 	}

// 	if requestBodySize < 0 {
// 		requestBodySize = -1
// 	}

// 	daprCMD := binaryFilePath(defaultDaprBinPath(), "daprd")

// 	args := []string{
// 		"--app-id", appID,
// 		"--dapr-http-port", strconv.Itoa(daprHTTPPort),
// 		"--dapr-grpc-port", strconv.Itoa(daprGRPCPort),
// 		"--log-level", logLevel,
// 		"--app-max-concurrency", strconv.Itoa(maxConcurrency),
// 		"--app-protocol", protocol,
// 		"--components-path", componentsPath,
// 		"--metrics-port", strconv.Itoa(metricsPort),
// 		"--dapr-http-max-request-size", strconv.Itoa(requestBodySize),
// 	}
// 	if appPort > -1 {
// 		args = append(args, "--app-port", strconv.Itoa(appPort))
// 	}
// 	args = append(args, "--placement-host-address")

// 	if runtime.GOOS == daprWindowsOS {
// 		args = append(args, fmt.Sprintf("%s:6050", placementHost))
// 	} else {
// 		args = append(args, fmt.Sprintf("%s:50005", placementHost))
// 	}

// 	if configFile != "" {
// 		args = append(args, "--config", configFile)
// 		sentryAddress := mtlsEndpoint(configFile)
// 		if sentryAddress != "" {
// 			// mTLS is enabled locally, set it up
// 			args = append(args, "--enable-mtls", "--sentry-address", sentryAddress)
// 		}
// 	}

// 	if enableProfiling {
// 		if profilePort == -1 {
// 			pp, err := freeport.GetFreePort()
// 			if err != nil {
// 				return nil, -1, -1, -1, err
// 			}
// 			profilePort = pp
// 		}

// 		args = append(
// 			args,
// 			"--enable-profiling", "true",
// 			"--profile-port", strconv.Itoa(profilePort))
// 	}

// 	if appSSL {
// 		args = append(args, "--app-ssl", "true")
// 	}

// 	cmd := exec.Command(daprCMD, args...)
// 	return cmd, daprHTTPPort, daprGRPCPort, metricsPort, nil
// }

// func mtlsEndpoint(configFile string) string {
// 	if configFile == "" {
// 		return ""
// 	}

// 	b, err := ioutil.ReadFile(configFile)
// 	if err != nil {
// 		return ""
// 	}

// 	var config mtlsConfig
// 	err = yaml.Unmarshal(b, &config)
// 	if err != nil {
// 		return ""
// 	}

// 	if config.Spec.MTLS.Enabled {
// 		return sentryDefaultAddress
// 	}
// 	return ""
// }

// func getAppCommand(httpPort, grpcPort, metricsPort int, command string, args []string) (*exec.Cmd, error) {
func getAppCommand(command string, args []string) (*exec.Cmd, error) {
	cmd := exec.Command(command, args...)
	cmd.Env = os.Environ()
	// cmd.Env = append(
	// 	cmd.Env,
	// 	fmt.Sprintf("DAPR_HTTP_PORT=%v", httpPort),
	// 	fmt.Sprintf("DAPR_GRPC_PORT=%v", grpcPort),
	// 	fmt.Sprintf("DAPR_METRICS_PORT=%v", metricsPort))

	return cmd, nil
}

// func Run(config *RunConfig) (*RunOutput, error) {
func Run(config *RunOptions) (*RunOutput, error) {
	appID := config.Name
	// if appID == "" {
	// 	appID = strings.ReplaceAll(sillyname.GenerateStupidName(), " ", "-")
	// }

	// _, err := os.Stat(config.ComponentsPath)
	// if err != nil {
	// 	return nil, err
	// }

	apps, err := List()
	if err != nil {
		return nil, err
	}

	for _, a := range apps {
		fmt.Printf("app: %+v\n", a)
		if appID == a.AppID {
			return nil, fmt.Errorf(`YoMo serverless application with name "%s" is already running`, appID)
		}
	}

	// componentsLoader := components.NewStandaloneComponents(modes.StandaloneConfig{ComponentsPath: config.ComponentsPath})
	// _, err = componentsLoader.LoadComponents()
	// if err != nil {
	// 	return nil, err
	// }

	// daprCMD, daprHTTPPort, daprGRPCPort, metricsPort, err := getDaprCommand(appID, config.HTTPPort, config.GRPCPort, config.AppPort, config.ConfigFile, config.Protocol, config.EnableProfiling, config.ProfilePort, config.LogLevel, config.MaxConcurrency, config.PlacementHost, config.ComponentsPath, config.AppSSL, config.MetricsPort, config.MaxRequestBodySize)
	// if err != nil {
	// 	return nil, err
	// }

	// for _, a := range dapr {
	// 	if daprHTTPPort == a.HTTPPort {
	// 		return nil, fmt.Errorf("there's already a Dapr instance running with http port %v", daprHTTPPort)
	// 	} else if daprGRPCPort == a.GRPCPort {
	// 		return nil, fmt.Errorf("there's already a Dapr instance running with gRPC port %v", daprGRPCPort)
	// 	}
	// }

	argCount := len(config.Arguments)
	runArgs := []string{}
	var appCMD *exec.Cmd

	if argCount > 0 {
		cmd := config.Arguments[0]
		if len(config.Arguments) > 1 {
			runArgs = config.Arguments[1:]
		}

		// appCMD, err = getAppCommand(daprHTTPPort, daprGRPCPort, metricsPort, cmd, runArgs)
		appCMD, err = getAppCommand(cmd, runArgs)
		if err != nil {
			return nil, err
		}
	}

	return &RunOutput{
		// DaprCMD:      daprCMD,
		AppCMD: appCMD,
		AppID:  appID,
		// DaprHTTPPort: daprHTTPPort,
		// DaprGRPCPort: daprGRPCPort,
	}, nil
}

func RunCmdAndWait(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}

	err = cmd.Start()
	if err != nil {
		return "", err
	}

	resp, err := ioutil.ReadAll(stdout)
	if err != nil {
		return "", err
	}
	errB, err := ioutil.ReadAll(stderr)
	if err != nil {
		return "", nil
	}

	err = cmd.Wait()
	if err != nil {
		// in case of error, capture the exact message
		if len(errB) > 0 {
			return "", errors.New(string(errB))
		}
		return "", err
	}

	return string(resp), nil
}
