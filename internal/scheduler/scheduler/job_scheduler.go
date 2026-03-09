package scheduler

import (
	"dssim/internal/job"
	"dssim/internal/misc"
	"dssim/internal/scheduler/worker"
	"errors"
	"log/slog"
)

type Scheduler struct {
	workers worker.WorkerManager
	jobs    chan *job.Job
}

func NewSchdular(workers worker.WorkerManager, jobQueueSize int) Scheduler {
	jobs := make(chan *job.Job, jobQueueSize)
	return Scheduler{workers: workers, jobs: jobs}
}

func (s *Scheduler) CreateJob(args *job.NewJob, reply *string) error {
	job, err := job.CreateJob(args.Duration)

	if err != nil {
		return err
	}

	s.jobs <- job

	*reply = job.ID
	slog.Info(
		"create",
		"type",
		"job",
		"jobID",
		job.ID,
	)

	return nil
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

func (s *Scheduler) RegisterWorker(args *string, reply *string) error {
	id, err := misc.GenID()
	if err != nil {
		return err
	}
	ok := s.workers.NewWorker(*args, id)
	if !ok {
		return errors.New("failed to creat worker")
	}
	slog.Info(
		"registered",
		"type",
		"worker",
		"workerID",
		id,
		"workerIP",
		*args,
	)
	*reply = id
	return nil
}

func (s *Scheduler) Run() {
	for {
		job := <-s.jobs
		worker := s.workers.GetWorker()
		ok := worker.AssignJob(job)
		slog.Info("assigned", "type", "job", "jobID", job.ID, "workerID", worker.ID, "success", ok)
	}
}
