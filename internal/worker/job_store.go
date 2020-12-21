package worker

// JobStore defines an interface for saving, updating and finding a Job.
type JobStore interface {
	AddJob(job *Job) error
	UpdateJobResults(job *Job, status string, stdout string, stderr string) error
	Lock()
	Unlock()
	FindJob(id string) (*Job, error)
}
