package worker

import (
	"dssim/internal/network"
	"dssim/internal/task"
	"errors"
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
	Addr     string
}

const (
	IDLE WorkerState = iota
	BUSY
)

const heartbeatInterval = 2 * time.Second

type Worker struct {
	ID              string
	address         string
	leaderAddr      string
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
	err = w.retryRegistration()
	if err != nil {
		return err
	}

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

	if err != nil {
		return nil, err
	}

	if !rr.IsLeader && rr.Addr == "" {
		// retry reg
		return nil, errors.New("Leader not yet elected")
	} else if !rr.IsLeader && rr.Addr != "" {
		// redirected
		return &rr, errors.New("Contacted follower; redirected")
	}

	// save Worker ID and Leader address
	w.ID = rr.ID
	newLeaderAddr := strings.Split(rr.Addr, ":")[0] + ":" + os.Getenv("SCHEDULER_PORT")
	w.leaderAddr = newLeaderAddr

	// log
	slog.Info(
		"Worker registered",
		"ID",
		w.ID,
		"address",
		w.address,
	)

	println("Worker registration successful")

	return &rr, nil
}

func (w *Worker) retryRegistration() error {
	for {
		rr, err := w.registerWorker()
		if err == nil {
			break
		}

		switch err.Error() {
		case "Leader not yet elected":
			time.Sleep(1 * time.Second)
			continue
		case "Contacted follower; redirected":
			println("Follower contacted instead. Using provided Leader address.")
			newLeaderAddr := strings.Split(rr.Addr, ":")[0] + ":" + os.Getenv("SCHEDULER_PORT")
			if strings.Compare(newLeaderAddr, w.leaderAddr) != 0 {
				w.getClient(newLeaderAddr)
			}
		default:
			return err
		}

		slog.Info(
			"Worker registration failed, will retry",
			"error", err,
		)
		time.Sleep(1 * time.Second)
	}

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

			println("Leader heartbeat failed. Finding new leader...")
			w.findNewLeader()
		}
	}
}

func (w *Worker) findNewLeader() error {
	followerAddr := w.schedulers[1] // either the new leader or actual follower

	err := w.getClient(followerAddr)
	if err != nil {
		return err
	}

	err = w.retryRegistration()
	if err == nil {
		println("New Leader found")
	}

	return err
}
