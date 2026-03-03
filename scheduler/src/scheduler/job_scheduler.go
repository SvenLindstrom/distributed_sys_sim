package scheduler

import (
	"log/slog"
	"schdeuler/src/job"
	"schdeuler/src/worker"
)

type Scheduler struct {
	workers worker.WorkerManager
}

func NewSchdular(workers worker.WorkerManager) Scheduler {
	return Scheduler{workers: workers}

}

type WorkerRegistration struct {
	IP string
	ID string
}

func (s *Scheduler) CompleteJob(args *job.JobResult, reply *bool) error {
	s.workers.JobCompleted(args.WorkerID)
	*reply = true
	slog.Info(
		"completed",
		"type",
		"job",
		"jobID",
		args.JobID,
		"workerID",
		args.WorkerID,
		"success",
		args.Status,
	)

	return nil
}

func (s *Scheduler) RegisterWorker(args *WorkerRegistration, reply *bool) error {
	*reply = s.workers.NewWorker(args.IP, args.ID)
	slog.Info(
		"registered",
		"type",
		"worker",
		"workerID",
		args.ID,
		"workerIP",
		args.IP,
		"success",
		reply,
	)
	return nil
}

func (s *Scheduler) Run(jobs <-chan job.Job) {
	for {
		job := <-jobs
		worker := s.workers.GetWorker()
		ok := worker.AssignJob(job)
		slog.Info("assigned", "type", "job", "jobID", job.ID, "workerID", worker.ID, "success", ok)

	}
}
