package scheduler

import (
	"errors"
	"log/slog"
	"schdeuler/src/job"
	"schdeuler/src/misc"
	"schdeuler/src/worker"
)

type Scheduler struct {
	workers worker.WorkerManager
	jobs    chan *job.Job
}

func NewSchdular(workers worker.WorkerManager, jobQueueSize int) Scheduler {
	jobs := make(chan *job.Job, jobQueueSize)
	return Scheduler{workers: workers, jobs: jobs}
}

type WorkerRegistration struct {
	IP string
}

func (s *Scheduler) CreateJob(args *job.NewJob, reply *string) error {
	job, err := job.CreateJob(args.Duration)

	if err != nil {
		return err
	}

	s.jobs <- job

	*reply = job.ID
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

func (s *Scheduler) RegisterWorker(args *WorkerRegistration, reply *string) error {
	id, err := misc.GenID()
	if err != nil {
		return err
	}
	ok := s.workers.NewWorker(args.IP, id)
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
		args.IP,
		"success",
		reply,
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
