package job

import "schdeuler/src/misc"

type Job struct {
	ID       string
	Duration int
}

type JobResult struct {
	JobID    string
	WorkerID string
	Status   string
}

func NewJob(duration int) (Job, error) {
	id, err := misc.GenID()
	if err != nil {
		return Job{}, err
	}
	job := Job{id, duration}

	return job, nil
}
