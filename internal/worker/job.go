package worker

import (
	"log"
	"os"
	"os/exec"
	"strconv"
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
	ID       string
	Pid      int
	Status   string
	Stdout   string
	Stderr   string
	Command  string
	ExitCode string
}

// Start creates a process to run the command and save the Job to the given store.
func (job *Job) Start(store JobStore) error {
	job.ID = uuid.NewV4().String()

	commandName, commandArguments := parseCommand(job.Command)
	cmd := exec.Command(commandName, commandArguments...)
	cmd.Stdout = &jobOutputWriter{outputType: "stdout", jobID: job.ID, store: store}
	cmd.Stderr = &jobOutputWriter{outputType: "stderr", jobID: job.ID, store: store}

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
	go job.wait(cmd, store)

	return nil
}

// Stop attempts to stop a running command
func (job *Job) Stop(store JobStore) error {
	process, err := os.FindProcess(job.Pid)
	if err != nil {
		return errors.Wrap(err, "Error finding job's process")
	}

	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		return errors.Wrap(err, "Error stopping job")
	}

	values := map[string]string{"Status": Stopped, "ExitCode": "-1"}
	store.UpdateJob(job.ID, values)

	return nil
}

func (job *Job) wait(cmd *exec.Cmd, store JobStore) {
	values := make(map[string]string)

	err := cmd.Wait()

	if commandStoppedBySignal(cmd) {
		return
	}

	if err != nil {
		log.Println(err)
		values["Status"] = Errored
	} else {
		values["Status"] = Completed
	}

	values["ExitCode"] = strconv.Itoa(cmd.ProcessState.ExitCode())
	store.UpdateJob(job.ID, values)
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
