package scheduler

import (
	"dssim/internal/job"
	"dssim/internal/scheduler/worker"
	"testing"
)

type MockWorkerManager struct {
}

func (mm *MockWorkerManager) NewWorker(address string, id string) bool {
	return true
}

func (mm *MockWorkerManager) JobCompleted(workerID string) {
}

func (mm *MockWorkerManager) GetWorker() *worker.Worker {
	return &worker.Worker{}
}

func TestCreatJob(t *testing.T) {
	scheduler := NewSchdular(&MockWorkerManager{}, 3)
	newJob := job.NewJob{Duration: 10}
	var reply string
	scheduler.CreateJob(&newJob, &reply)

	if reply == "" {
		t.Fatal("idea not replyed")
	}
	job := <-scheduler.jobs

	if job.Duration != 10 {
		t.Fatal("job not created properly")
	}
}

func TestCompleteJob(t *testing.T) {
	scheduler := NewSchdular(&MockWorkerManager{}, 3)
	jobRes := job.JobResult{"11", "11", "ok"}
	var ok bool
	scheduler.CompleteJob(&jobRes, &ok)

	if !ok {
		t.Fatal("expected ok")
	}
}

func TestRegisterWorker(t *testing.T) {
	scheduler := NewSchdular(&MockWorkerManager{}, 3)
	reg := WorkerRegistration{"testIP"}

	var id string

	scheduler.RegisterWorker(&reg, &id)

	if id == "" {
		t.Fatal("id not set")
	}

}
