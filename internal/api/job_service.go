package api

import (
	"errors"

	"github.com/tmnhat2001/worker-service/internal/worker"
)

var errUnauthorizedUser = errors.New("The user is not authorized to access this job")

type jobService struct {
	jobStore worker.JobStore
}

func newJobService() *jobService {
	return &jobService{
		jobStore: &worker.MemoryJobStore{
			Jobs: make(map[string]worker.Job),
		},
	}
}

func (s jobService) startJob(config jobActionConfig) (worker.Job, error) {
	job := worker.Job{Command: config.command, User: config.user.Username}
	err := (&job).Start(s.jobStore)
	return job, err
}

func (s jobService) stopJob(config jobActionConfig) (worker.Job, error) {
	job, err := s.getJob(config)
	if err != nil {
		return job, err
	}

	err = job.Stop(s.jobStore)
	if err != nil {
		return job, err
	}

	updatedJob, err := s.jobStore.FindJob(config.jobID)
	if err != nil {
		return updatedJob, err
	}

	return updatedJob, nil
}

func (s jobService) getJob(config jobActionConfig) (worker.Job, error) {
	job, err := s.jobStore.FindJob(config.jobID)
	if err != nil {
		return job, err
	}

	if job.User != config.user.Username {
		return worker.Job{}, errUnauthorizedUser
	}

	return job, nil
}

type jobActionConfig struct {
	command string
	user    *User
	jobID   string
}
