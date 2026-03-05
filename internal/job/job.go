package job

import "dssim/internal/misc"

type Job struct {
	ID       string
	Duration int
}

type NewJob struct {
	Duration int
}

type JobResult struct {
	JobID    string
	WorkerID string
	Status   string
}

func CreateJob(duration int) (*Job, error) {
	id, err := misc.GenID()
	if err != nil {
		return &Job{}, err
	}
	job := &Job{id, duration}

	return job, nil
}
