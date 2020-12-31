package worker

import (
	"io"
	"log"
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

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "Unable to get stdout pipe of the job")
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return errors.Wrap(err, "Unable to get stderr pipe of the job")
	}

	job.ID = uuid.NewV4().String()

	err = cmd.Start()
	if err != nil {
		job.Status = Errored
		store.AddJob(job)

		return errors.Wrap(err, "Unable to start job")
	}

	job.Pid = cmd.Process.Pid
	job.Status = Running
	store.AddJob(job)

	// These go routines will exit when the command completes or
	// when an error occurs from reading the pipes
	go job.saveOutput(stdoutPipe, store, "stdout")
	go job.saveOutput(stderrPipe, store, "stderr")

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

	job.Status = Stopped
	job.ExitCode = -1
	store.UpdateJob(job)

	return nil
}

func (job *Job) wait(cmd *exec.Cmd, store JobStore) {
	err := cmd.Wait()

	if commandStoppedBySignal(cmd) {
		return
	}

	if err != nil {
		log.Println(err)
		job.Status = Errored
	} else {
		job.Status = Completed
	}

	job.ExitCode = cmd.ProcessState.ExitCode()
	store.UpdateJob(job)
}

func (job *Job) saveOutput(pipe io.ReadCloser, store JobStore, outputType string) {
	var builder strings.Builder
	buf := make([]byte, 1024, 1024)
	for {
		n, err := pipe.Read(buf)
		if n > 0 {
			builder.Write(buf[:n])
			if outputType == "stdout" {
				job.Stdout = builder.String()
			} else {
				job.Stderr = builder.String()
			}
			store.UpdateJob(job)
		}

		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}

			return
		}
	}
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
