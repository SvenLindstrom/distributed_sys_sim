package generator

import (
	"dssim/internal/job"
	"time"
)

type JobRec struct {
	Job         job.Job
	Submit_time time.Time
}

type JobTable struct {
	Jobs map[string]JobRec
}

func (jt *JobTable) JobDone(id string) {
	delete(jt.Jobs, id)
}

func (jt *JobTable) JobSubmitted(job job.Job) {
	jobRec := JobRec{job, time.Now()}
	jt.Jobs[job.ID] = jobRec
}
