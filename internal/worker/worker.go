package worker

import (
	"dssim/internal/job"
	"dssim/internal/network"
	"dssim/internal/scheduler/scheduler"
	"log/slog"
	"net/http"
	"net/rpc"
	"time"
)

type WorkerState int

const (
	IDLE WorkerState = iota
	BUSY
)

type Worker struct {
	ID              string
	address         string
	SchedulerAddr   string
	state           WorkerState
	currentJob      string
	schedulerClient network.RPCClient
}

func NewWorker(address, SchedulerAddr string) *Worker {
	return &Worker{
		address:       address,
		SchedulerAddr: SchedulerAddr,
		state:         IDLE,
	}
}

func (w *Worker) Run() error {
	err := rpc.Register(w)
	if err != nil {
		return err
	}

	rpc.HandleHTTP()

	err = w.getClient()
	if err != nil {
		return err
	}

	go http.ListenAndServe(w.address, nil)

	// register Worker
	err = w.registerWorker()
	if err != nil {
		return err
	}

	return nil
}

func (w *Worker) getClient() error {
	dialer := network.RealRPCDialer{}
	client, err := dialer.Dial(w.SchedulerAddr)
	if err != nil {
		return err
	}
	// store client
	w.schedulerClient = client

	return nil
}

// requests from Scheduler

func (w *Worker) AssignJob(job *job.Job, reply *bool) error {
	w.currentJob = job.ID
	w.state = BUSY

	go w.executeJob(*job)

	*reply = true
	return nil
}

func (w *Worker) executeJob(job job.Job) {
	// simulate job
	time.Sleep(time.Duration(job.Duration))

	w.state = IDLE
	w.currentJob = ""

	// report to Scheduler
	err := w.completeJob(job.ID)
	if err != nil {
		slog.Error(
			"Job completion could not be reported",
			"worker",
			w.ID,
			"job",
			job.ID,
			"error",
			err,
		)
	}
}

// requests to Scheduler

func (w *Worker) registerWorker() error {
	// call Scheduler to register Worker
	var workerID string
	args := scheduler.WorkerRegistration{IP: w.address}
	err := w.schedulerClient.Call("Scheduler.RegisterWorker", &args, &workerID)
	if err != nil {
		return err
	}

	// save Worker ID
	w.ID = workerID

	// log
	slog.Info(
		"Worker registered",
		"ID",
		w.ID,
		"address",
		w.address,
	)

	return nil
}

func (w *Worker) completeJob(jobID string) error {
	// define JobResult
	args := job.JobResult{
		JobID:    jobID,
		WorkerID: w.ID,
		Status:   "completed",
	}

	// call Scheduler to report job completion
	var reply bool
	err := w.schedulerClient.Call("Scheduler.CompleteJob", &args, &reply)
	if err != nil {
		return err
	}

	// log
	slog.Info(
		"Job completed and reported",
		"worker",
		w.ID,
		"job",
		jobID,
	)

	return nil
}
