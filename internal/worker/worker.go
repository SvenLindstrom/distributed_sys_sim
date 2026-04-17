package worker

import (
	"dssim/internal/network"
	"dssim/internal/task"
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

const heartbeatInterval = 2 * time.Second

type Worker struct {
	ID              string
	address         string
	schedulers      []string
	state           WorkerState
	currentTask     string
	schedulerClient network.RPCClient
}

func NewWorker(address string, schedulers []string) *Worker {
	return &Worker{
		address:    address,
		schedulers: schedulers,
		state:      IDLE,
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

	// start hearbeat
	go w.startHeartbeat()

	select {}
}

func (w *Worker) getClient() error {
	client, err := network.RealRPCDialer().Dial(w.schedulers[0])

	if err != nil {
		return err
	}
	// store client
	w.schedulerClient = client

	return nil
}

// requests from Scheduler

func (w *Worker) AssignTask(task *task.Task, reply *bool) error {
	w.currentTask = task.ID
	w.state = BUSY

	go w.executeTask(*task)

	*reply = true
	return nil
}

func (w *Worker) executeTask(task task.Task) {
	// simulate task
	time.Sleep(time.Duration(task.Duration) * time.Millisecond)

	w.state = IDLE
	w.currentTask = ""

	// report to Scheduler
	err := w.completeTask(task.ID)
	if err != nil {
		slog.Error(
			"Task completion could not be reported",
			"worker",
			w.ID,
			"task",
			task.ID,
			"error",
			err,
		)
	}
}

// requests to Scheduler

func (w *Worker) registerWorker() error {
	// call Scheduler to register Worker
	var workerID string
	err := w.schedulerClient.Call("Scheduler.RegisterWorker", &w.address, &workerID)
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

func (w *Worker) completeTask(taskID string) error {
	// define TaskResult
	args := task.TaskResult{
		TaskID:   taskID,
		WorkerID: w.ID,
		Status:   "completed",
	}

	// call Scheduler to report task completion
	var reply bool
	err := w.schedulerClient.Call("Scheduler.CompleteTask", &args, &reply)
	if err != nil {
		return err
	}

	// log
	slog.Info(
		"Task completed and reported",
		"worker",
		w.ID,
		"task",
		taskID,
	)

	return nil
}

func (w *Worker) startHeartbeat() {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for range ticker.C {
		var reply bool
		err := w.schedulerClient.Call("Scheduler.Ping", w.ID, &reply)

		if err != nil {
			// log detected leader crash
			slog.Warn(
				"Heartbeat failed",
				"worker", w.ID,
				"error", err,
			)
			// some actions to handle leader crash
		}
	}
}
