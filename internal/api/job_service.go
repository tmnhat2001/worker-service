package api

import (
	"errors"

	"github.com/tmnhat2001/worker-service/internal/worker"
)

var errUnauthorizedUser = errors.New("The user is not authorized to access this job")

type jobService struct {
	jobStore worker.JobStore
	user     *User
}

func (s jobService) startJob(job *worker.Job) (worker.Job, error) {
	job.User = s.user.Username
	err := job.Start(s.jobStore)
	return *job, err
}

func (s jobService) stopJob(jobID string) (worker.Job, error) {
	job, err := s.getJob(jobID)
	if err != nil {
		return job, err
	}

	err = job.Stop(s.jobStore)
	if err != nil {
		return job, err
	}

	updatedJob, err := s.jobStore.FindJob(jobID)
	if err != nil {
		return updatedJob, err
	}

	return updatedJob, nil
}

func (s jobService) getJob(jobID string) (worker.Job, error) {
	job, err := s.jobStore.FindJob(jobID)
	if err != nil {
		return job, err
	}

	if job.User != s.user.Username {
		return worker.Job{}, errUnauthorizedUser
	}

	return job, nil
}
