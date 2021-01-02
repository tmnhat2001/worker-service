package wkct

import (
	"errors"
	"os"

	"github.com/tmnhat2001/worker-service/client/api"
	"gopkg.in/alecthomas/kingpin.v2"
)

// CLI represents a commandline client for the Worker API
type CLI struct {
	api *api.WorkerAPI
}

// NewCLI creates a new CLI. The configurations are gathered from environment variables
func NewCLI() (*CLI, error) {
	config, err := apiConfigFromEnvVars()
	if err != nil {
		return nil, err
	}

	apiClient, err := api.NewWorkerAPI(config)
	if err != nil {
		return nil, err
	}

	return &CLI{api: apiClient}, nil
}

func apiConfigFromEnvVars() (api.WorkerAPIConfig, error) {
	username, ok := os.LookupEnv("WORKER_USERNAME")
	if !ok {
		return api.WorkerAPIConfig{}, errors.New("Please make sure that the environment variable WORKER_USERNAME is set with the API username")
	}

	password, ok := os.LookupEnv("WORKER_PASSWORD")
	if !ok {
		return api.WorkerAPIConfig{}, errors.New("Please make sure that the environment variable WORKER_PASSWORD is set with the API password")
	}

	certFilePath, ok := os.LookupEnv("WORKER_CERT")
	if !ok {
		return api.WorkerAPIConfig{}, errors.New("Please make sure that the environment variable WORKER_CERT is set with the path to the server certificate")
	}

	return api.WorkerAPIConfig{
		Username:     username,
		Password:     password,
		CertFilePath: certFilePath,
	}, nil
}

// Run will parse the command arguments and call the appropriate handler for it
func (c *CLI) Run() {
	cli := kingpin.New("wkct", "A CLI tool to interact with the Worker API")

	cli.HelpFlag.Short('h')

	start := cli.Command("start", "Start a job to run the given Linux command")
	startCommandArg := start.Arg("command", "Linux command to be run").Required().String()

	stop := cli.Command("stop", "Stop a job")
	stopCommandArg := stop.Arg("job_id", "The job ID").Required().String()

	getJob := cli.Command("job", "Get the information about a job")
	getJobCommandArg := getJob.Arg("job_id", "The job ID").Required().String()

	commandHandler := &commandHandler{api: c.api}

	switch kingpin.MustParse(cli.Parse(os.Args[1:])) {
	case start.FullCommand():
		commandHandler.startJob(*startCommandArg)
	case stop.FullCommand():
		commandHandler.stopJob(*stopCommandArg)
	case getJob.FullCommand():
		commandHandler.getJob(*getJobCommandArg)
	}
}
