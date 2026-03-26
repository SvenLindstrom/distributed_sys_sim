package manager

import (
	"dssim/internal/misc"
	"dssim/internal/network"
	"dssim/internal/scheduler/state"
	"dssim/internal/scheduler/worker"
	"dssim/internal/task"
	"errors"
	"log/slog"

	"github.com/hashicorp/raft"
)

type Scheduler struct {
	workers worker.WorkerManager
	state   state.State
	client  network.RPCClient
}

func NewScheduler(worker worker.WorkerManager, state state.State) Scheduler {
	client, err := network.RealRPCDialer().Dial("Generator:8081")

	if err != nil {
		println("failed to dial generator")
		println(err.Error())
	}

	return Scheduler{worker, state, client}
}

func (s *Scheduler) NotifiyGenerator(addr string) bool {
	var ok bool
	err := s.client.Call("Generator.ReadyForWork", &addr, &ok)

	if err != nil || !ok {
		println("failed to connect to generator")
	}
	return ok
}

func (s *Scheduler) CreateTask(task *task.Task, reply *bool) error {
	err := s.state.AddTask(task)
	if err != nil {
		*reply = false
		return err
	}
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

func (s *Scheduler) RegisterScheduler(reg *raft.Server, res *bool) error {
	err := s.state.RegisterScheduler(string(reg.ID), string(reg.Address))

	// res.Err = err
	// if err != nil {
	// 	*res = network.RPCRes{Val: "test", Err: err.Error()}
	// } else {
	// 	*res = network.RPCRes{Val: "no Teest", Err: ""}
	// }
	// res.Val = "test"
	// println(err.Error())
	return err
}
