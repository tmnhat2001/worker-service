package worker

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// The following constants are possible values for the Status of a Job
const (
	Completed = "completed"
	Errored   = "errored"
	Running   = "running"
	Stopped   = "stopped"
)

// Job represents a job created to run a Linux command
type Job struct {
	ID         string
	Pid        int
	Status     string
	Stdout     string
	Stderr     string
	RawCommand string
	ExitCode   int
}

// Start creates a process to run the command and save the Job to the given store.
func (job *Job) Start(store JobStore) error {
	commandName, commandArguments := parseCommand(job.RawCommand)
	cmd := exec.Command(commandName, commandArguments...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	job.ID = uuid.NewV4().String()

	err := cmd.Start()
	if err != nil {
		job.Status = Errored
		store.AddJob(job)

		return errors.Wrap(err, "Unable to start job")
	}

	job.Pid = cmd.Process.Pid
	job.Status = Running
	store.AddJob(job)

	// This goroutine will exit when the command completes or is stopped by calling Stop
	go job.wait(cmd, &stdout, &stderr, store)

	return nil
}

func (job *Job) wait(cmd *exec.Cmd, stdout *bytes.Buffer, stderr *bytes.Buffer, store JobStore) {
	err := cmd.Wait()

	var newStatus string
	if err != nil {
		// TODO: Add a logger instead of printing to stdout
		fmt.Println(err)
		newStatus = Errored
	} else {
		newStatus = Completed
	}

	if !commandStoppedBySignal(cmd) {
		job.Status = newStatus
	}
	job.Stdout = stdout.String()
	job.Stderr = stderr.String()
	job.ExitCode = cmd.ProcessState.ExitCode()
	store.UpdateJob(job)
}

// Stop attempts to stop a running command and update the Job results in the store.
func (job *Job) Stop(store JobStore) error {
	process, err := os.FindProcess(job.Pid)
	if err != nil {
		return errors.Wrap(err, "Error finding job's process")
	}

	if !job.isRunning() {
		return nil
	}

	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		return errors.Wrap(err, "Error stopping job")
	}

	store.UpdateJobStatus(job, Stopped)

	return nil
}

func (job *Job) isRunning() bool {
	return job.Status == Running
}

func parseCommand(rawCommand string) (string, []string) {
	splitCommand := strings.Split(rawCommand, " ")

	if len(splitCommand) < 2 {
		return rawCommand, []string{}
	}

	return splitCommand[0], splitCommand[1:]
}

func commandStoppedBySignal(cmd *exec.Cmd) bool {
	return cmd.ProcessState.ExitCode() == -1
}
