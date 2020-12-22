package worker

import (
	"errors"
	"sync"
)

// JobStore defines an interface for saving, updating and finding a Job.
type JobStore interface {
	AddJob(job *Job)
	UpdateJob(job *Job) error
	UpdateJobStatus(job *Job, status string) error
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

// UpdateJob updates the values of the given Job in the store
func (store *MemoryJobStore) UpdateJob(job *Job) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	_, ok := store.Jobs[job.ID]
	if !ok {
		return errors.New("worker: Unable to find job in store")
	}

	store.Jobs[job.ID] = job

	return nil
}

// UpdateJobStatus updates the Status of the given Job in the store
func (store *MemoryJobStore) UpdateJobStatus(job *Job, status string) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	_, ok := store.Jobs[job.ID]
	if !ok {
		return errors.New("worker: Unable to find job in store")
	}

	store.Jobs[job.ID].Status = status

	return nil
}

// FindJob returns a copy of the Job if it is found. Otherwise, returns an error.
func (store *MemoryJobStore) FindJob(id string) (*Job, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	job, ok := store.Jobs[id]
	if !ok {
		return nil, errors.New("worker: Unable to find job in store")
	}

	return job, nil
}
