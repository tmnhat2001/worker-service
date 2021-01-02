package wkct

import (
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/tmnhat2001/worker-service/client/api"
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

	var job job
	json.Unmarshal(response, &job)

	displayJob(job)
}

func displayJob(job job) {
	tmpl := template.Must(template.New("job").Parse(jobTemplate))

	err := tmpl.Execute(os.Stdout, job)
	if err != nil {
		fmt.Println(err)
	}
}
