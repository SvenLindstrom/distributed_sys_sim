package worker

import (
	"dssim/internal/network"
	"dssim/internal/task"
	"log/slog"
	"net/http"
	"net/rpc"
	"os"
	"strings"
	"time"
)

type WorkerState int

type RegistrationReply struct {
	IsLeader bool
	ID       string
	addr     string
}

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

	err = w.getClient(w.schedulers[0])
	if err != nil {
		return err
	}

	go http.ListenAndServe(w.address, nil)

	// register Worker and retry
	w.retryRegistration()

	// start hearbeat
	go w.startHeartbeat()

	select {}
}

func (w *Worker) getClient(address string) error {
	client, err := network.RealRPCDialer().Dial(address)

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

func (w *Worker) registerWorker() (*RegistrationReply, error) {
	// call Scheduler to register Worker
	var rr RegistrationReply
	err := w.schedulerClient.Call("Scheduler.RegisterWorker", &w.address, &rr)

	if err != nil && !rr.IsLeader {
		return nil, err
	} else if err != nil && rr.addr != "" {
		return &rr, err
	}

	// save Worker ID
	w.ID = rr.ID

	// log
	slog.Info(
		"Worker registered",
		"ID",
		w.ID,
		"address",
		w.address,
	)

	println("Worker registration successful.")

	return &rr, nil
}

func (w *Worker) retryRegistration() {
	for {
		_, err := w.registerWorker()
		if err == nil {
			break
		}

		slog.Info(
			"Worker registration failed, will retry",
			"error", err,
		)
		time.Sleep(1 * time.Second)
	}
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

			println("Leader heartbeat failed. Finding new leader...")
			w.findNewLeader()
		}
	}
}

func (w *Worker) findNewLeader() {
	followerAddr := w.schedulers[1] // either the new leader or actual follower

	err := w.getClient(followerAddr)
	if err == nil {
		rr, err := w.registerWorker()

		if err != nil && !rr.IsLeader {
			// parse new leader address
			newLeaderAddr := strings.Split(rr.addr, ":")[0] + ":" + os.Getenv("SCHEDULER_PORT")
			println("Follower contacted instead. Using provided Leader address.")

			// dial and re-register with new leader
			w.getClient(newLeaderAddr)
			w.retryRegistration()
		}

		println("New Leader found. Proceeding as usual.")
	}
}
