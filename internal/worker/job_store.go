package worker

import (
	"errors"
	"sync"
)

// JobStore defines an interface for saving, updating and finding a Job.
type JobStore interface {
	AddJob(job *Job)
	UpdateJob(job *Job) error
	FindJob(id string) (*Job, error)
}

// MemoryJobStore implements the JobStore interface and stores Jobs in memory
type MemoryJobStore struct {
	Jobs  map[string]*Job
	mutex sync.RWMutex
}

// AddJob adds a Job to the memory store
func (store *MemoryJobStore) AddJob(job *Job) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

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

// FindJob returns a copy of the Job if it is found. Otherwise, returns an error.
func (store *MemoryJobStore) FindJob(id string) (*Job, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	job, ok := store.Jobs[id]
	if !ok {
		return nil, errors.New("worker: Unable to find job in store")
	}

	return job, nil
}
