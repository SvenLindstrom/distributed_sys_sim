package manager

import (
	"dssim/internal/misc"
	"dssim/internal/network"
	"dssim/internal/scheduler/state"
	"dssim/internal/scheduler/worker"
	"dssim/internal/task"
	"errors"
	"log/slog"
)

type Scheduler struct {
	workers worker.WorkerManager
	state   *state.SchedulerState
	client  network.RPCClient
}

func NewScheduler(worker worker.WorkerManager, state *state.SchedulerState) Scheduler {
	dialer := network.RealRPCDialer{}
	client, err := dialer.Dial("Generator:8081")

	if err != nil {
		println("failed to dial generator")
		println(err.Error())
	}

	return Scheduler{worker, state, client}
}

func (s *Scheduler) NotifiyGenerator() bool {
	var ok bool
	err := s.client.Call("Generator.ReadyForWork", "", &ok)

	if err != nil || !ok {
		println("failed to connect to generator")
	}
	return ok
}

func (s *Scheduler) CreateTask(task *task.Task, reply *bool) error {
	s.state.AddTask(task)
	*reply = true
	slog.Info(
		"create",
		"type",
		"task",
		"taskID",
		task.ID,
	)
	return nil
}

func (s *Scheduler) CompleteTask(args *task.TaskResult, reply *bool) error {

	s.workers.TaskCompleted(args.WorkerID)
	*reply = true
	s.state.CompleteTask(args.TaskID, args.WorkerID)

	var ok bool
	s.client.Call("Generator.TaskCompleted", &args.TaskID, &ok)

	slog.Info(
		"completed",
		"type",
		"task",
		"taskID",
		args.TaskID,
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
