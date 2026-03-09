package scheduler

import (
	"dssim/internal/job"
	"dssim/internal/misc"
	"dssim/internal/network"
	"dssim/internal/scheduler/worker"
	"errors"
	"log/slog"
)

type Scheduler struct {
	workers worker.WorkerManager
	jobs    chan *job.Job
	client  network.RPCClient
}

func NewSchdular(workers worker.WorkerManager, jobQueueSize int) Scheduler {
	dialer := network.RealRPCDialer{}

	client, _ := dialer.Dial("Generator:8081")

	jobs := make(chan *job.Job, jobQueueSize)
	return Scheduler{workers: workers, jobs: jobs, client: client}
}

func (s *Scheduler) CreateJob(job *job.Job, reply *bool) error {

	s.jobs <- job

	*reply = true
	slog.Info(
		"create",
		"type",
		"job",
		"jobID",
		job.ID,
	)

	return nil
}

func (s *Scheduler) notifyGenerator(jobId string) {
	var ok bool
	s.client.Call("Generator.JobCompleted", &jobId, &ok)
}

func (s *Scheduler) CompleteJob(args *job.JobResult, reply *bool) error {
	s.workers.JobCompleted(args.WorkerID)
	go s.notifyGenerator(args.JobID)
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

	var ok bool
	s.client.Call("Generator.ReadyForWork", "", &ok)

	for {
		job := <-s.jobs
		worker := s.workers.GetWorker()
		ok := worker.AssignJob(job)
		slog.Info("assigned", "type", "job", "jobID", job.ID, "workerID", worker.ID, "success", ok)
	}
}
