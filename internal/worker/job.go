package worker

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// Job represents a job created to run a Linux command
type Job struct {
	ID               string
	Pid              int
	Status           string
	Stdout           string
	Stderr           string
	RawCommand       string
	commandName      string
	commandArguments []string
}

// Start creates a process to run the command and save the Job to the given store.
func (job *Job) Start(store JobStore) error {
	job.parseCommand()
	cmd := exec.Command(job.commandName, job.commandArguments...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	job.ID = uuid.NewV4().String()

	err := cmd.Start()
	if err != nil {
		job.Status = "errored"
		store.AddJob(job)

		return errors.Wrap(err, "Unable to start job")
	}

	job.Pid = cmd.Process.Pid
	job.Status = "running"
	store.AddJob(job)

	// This goroutine will exit when the command completes or is stopped by calling Stop
	go job.Wait(cmd, &stdout, &stderr, store)

	return nil
}

// Wait waits for the command to finish and update the Job results in the store.
func (job *Job) Wait(cmd *exec.Cmd, stdout *bytes.Buffer, stderr *bytes.Buffer, store JobStore) {
	err := cmd.Wait()

	store.Lock()
	defer store.Unlock()

	var newStatus string
	if err != nil {
		// TODO: Add a logger instead of printing to stdout
		fmt.Println(err)
		newStatus = "errored"
	} else {
		newStatus = "completed"
	}

	if job.isRunning() {
		store.UpdateJobResults(job, newStatus, stdout.String(), stderr.String())
	}
}

// Stop attempts to stop a running command and update the Job results in the store.
func (job *Job) Stop(store JobStore) error {
	process, err := os.FindProcess(job.Pid)
	if err != nil {
		return errors.Wrap(err, "Error finding job's process")
	}

	store.Lock()
	defer store.Unlock()

	if !job.isRunning() {
		return nil
	}

	err = process.Kill()
	if err != nil {
		return errors.Wrap(err, "Error stopping job")
	}

	store.UpdateJobResults(job, "stopped", "", "")

	return nil
}

func (job *Job) parseCommand() {
	splitCommand := strings.Split(job.RawCommand, " ")
	job.commandName = splitCommand[0]
	job.commandArguments = splitCommand[1:]
}

func (job *Job) isRunning() bool {
	return job.Status == "running"
}
