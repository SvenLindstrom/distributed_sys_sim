package worker

import (
	"dssim/internal/job"
	"testing"
)

func TestWorkerAssignJob(t *testing.T) {
	worker := Worker{IP: "ip Address", ID: "testID", job: "", state: IDLE, client: &MockRPCClient{}}
	job := &job.Job{ID: "jobID", Duration: 123}

	ok := worker.AssignJob(job)
	if !ok {
		t.Fatal("expected success")
	}
	if worker.job != "jobID" {
		t.Fatal("jobID not set")
	}
	if worker.state != BUSY {
		t.Fatal("worker state not busy")
	}
}

func TestWorkerJobFinished(t *testing.T) {
	worker := Worker{
		IP:     "ip Address",
		ID:     "testID",
		job:    "jobID",
		state:  BUSY,
		client: &MockRPCClient{},
	}
	worker.JobFinished()
	if worker.job != "" {
		t.Fatal("jobID not set")
	}
	if worker.state != READY {
		t.Fatal("worker state not busy")
	}
}
