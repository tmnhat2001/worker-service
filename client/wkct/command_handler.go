package wkct

import (
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/tmnhat2001/worker-service/client/api"
	"github.com/tmnhat2001/worker-service/internal/worker"
)

const jobTemplate = `Job ID: {{.ID}}
Command: {{.Command}}
Status: {{.Status}}
ExitCode: {{.ExitCode}}
Stdout: {{.Stdout}}
Stderr: {{.Stderr}}
User: {{.User}}
`

type commandHandler struct {
	api *api.WorkerAPI
}

func (c *commandHandler) startJob(command string) {
	response, err := c.api.StartJob(command)
	handleResponse(response, err)
}

func (c *commandHandler) stopJob(jobID string) {
	response, err := c.api.StopJob(jobID)
	handleResponse(response, err)
}

func (c *commandHandler) getJob(jobID string) {
	response, err := c.api.GetJob(jobID)
	handleResponse(response, err)
}

func handleResponse(response []byte, err error) {
	if err != nil {
		fmt.Println(err)
		return
	}

	var job worker.Job
	err = json.Unmarshal(response, &job)
	if err != nil {
		fmt.Println(err)
		return
	}

	displayJob(job)
}

func displayJob(job worker.Job) {
	tmpl, err := template.New("job").Parse(jobTemplate)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = tmpl.Execute(os.Stdout, job)
	if err != nil {
		fmt.Println(err)
	}
}
