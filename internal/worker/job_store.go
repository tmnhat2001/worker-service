package worker

import (
	"errors"
	"sync"
)

// JobStore defines an interface for saving, updating and finding a Job.
type JobStore interface {
	AddJob(job *Job)
	UpdateJobResults(job *Job, status string, stdout string, stderr string) error
	Lock()
	Unlock()
	FindJob(id string) (*Job, error)
}

// MemoryJobStore implements the JobStore interface and stores Jobs in memory
type MemoryJobStore struct {
	Jobs  map[string]*Job
	mutex sync.Mutex
}

// AddJob adds a Job to the memory store
func (store *MemoryJobStore) AddJob(job *Job) {
	store.Jobs[job.ID] = job
}

// UpdateJobResults updates the status and outputs of the given job in the store.
// This method returns an error if the Job cannot be found in the store.
func (store *MemoryJobStore) UpdateJobResults(job *Job, status string, stdout string, stderr string) error {
	storedJob, ok := store.Jobs[job.ID]
	if !ok {
		return errors.New("worker: Unable to find job in store")
	}

	storedJob.Status = status
	storedJob.Stdout = stdout
	storedJob.Stderr = stderr

	return nil
}

// Lock acquires a lock on the store
func (store *MemoryJobStore) Lock() {
	store.mutex.Lock()
}

// Unlock releases a lock on the store
func (store *MemoryJobStore) Unlock() {
	store.mutex.Unlock()
}

// FindJob returns a Job given its ID. This method returns an error if the Job cannot be found.
func (store *MemoryJobStore) FindJob(id string) (*Job, error) {
	job, ok := store.Jobs[id]
	if !ok {
		return nil, errors.New("worker: Unable to find job in store")
	}

	return job, nil
}
