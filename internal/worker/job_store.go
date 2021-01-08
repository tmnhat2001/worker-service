package worker

import (
	"errors"
	"sync"
)

// JobStore defines an interface for saving, updating and finding a Job.
type JobStore interface {
	AddJob(*Job)
	UpdateJob(string, map[string]string) error
	FindJob(string) (Job, error)
}

// MemoryJobStore implements the JobStore interface and stores Jobs in memory
type MemoryJobStore struct {
	Jobs  map[string]Job
	mutex sync.RWMutex
}

// AddJob adds a Job to the memory store
func (store *MemoryJobStore) AddJob(job *Job) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	jobCopy := Job{
		ID:       job.ID,
		Pid:      job.Pid,
		Status:   job.Status,
		Stdout:   job.Stdout,
		Stderr:   job.Stderr,
		Command:  job.Command,
		ExitCode: job.ExitCode,
	}
	store.Jobs[job.ID] = jobCopy
}

// UpdateJob updates the values of a Job in the store
func (store *MemoryJobStore) UpdateJob(jobID string, values map[string]string) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	job, ok := store.Jobs[jobID]
	if !ok {
		return errors.New("worker: Unable to find job in store")
	}

	newStatus, ok := values["Status"]
	if ok {
		job.Status = newStatus
	}

	newStdout, ok := values["Stdout"]
	if ok {
		job.Stdout = newStdout
	}

	newStderr, ok := values["Stderr"]
	if ok {
		job.Stderr = newStderr
	}

	newCommand, ok := values["Command"]
	if ok {
		job.Command = newCommand
	}

	newExitCode, ok := values["ExitCode"]
	if ok {
		job.ExitCode = newExitCode
	}

	store.Jobs[job.ID] = job

	return nil
}

// FindJob returns a copy of the Job if it is found. Otherwise, returns an error.
func (store *MemoryJobStore) FindJob(id string) (Job, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	job, ok := store.Jobs[id]
	if !ok {
		return Job{}, errors.New("worker: Unable to find job in store")
	}

	jobCopy := Job{
		ID:       job.ID,
		Pid:      job.Pid,
		Status:   job.Status,
		Stdout:   job.Stdout,
		Stderr:   job.Stderr,
		Command:  job.Command,
		ExitCode: job.ExitCode,
	}

	return jobCopy, nil
}
