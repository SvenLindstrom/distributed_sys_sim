package scheduler

import (
	"dssim/internal/misc"
	"dssim/internal/network"
	"dssim/internal/scheduler/state"
	"dssim/internal/scheduler/worker"
	"dssim/internal/task"
	"errors"
	"log/slog"
	"net/http"
	"net/rpc"
	"os"

	"github.com/hashicorp/raft"
)

type RegistrationReply struct {
	IsLeader bool
	ID       string
	Addr     string
}

type RpcInterface struct {
	workers worker.WorkerManager
	state   state.State
	client  network.RPCClient
}

func NewRPC(worker worker.WorkerManager, state state.State) RpcInterface {

	client, err := network.RealRPCDialer().Dial("Generator:8081")
	if err != nil {
		println("failed to dial generator")
		println(err.Error())
	}
	s := RpcInterface{worker, state, client}

	return s
}

func (s *RpcInterface) Start() {
	port := ":" + os.Getenv("SCHEDULER_PORT")
	rpc.RegisterName("Scheduler", s)
	rpc.HandleHTTP()
	go http.ListenAndServe(port, nil)
}

func (s *RpcInterface) NotifiyGenerator() bool {
	addr := os.Getenv("HOSTNAME")
	var ok bool
	err := s.client.Call("Generator.ReadyForWork", &addr, &ok)

	if err != nil || !ok {
		println("failed to connect to generator")
	}
	return ok
}

func (s *RpcInterface) CreateTask(task *task.Task, reply *bool) error {
	// println("Task recived")
	err := s.state.AddTask(task)
	if err != nil {
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

func (s *RpcInterface) notifyTaskCompleted(id string) {
	var ok bool
	s.client.Call("Generator.TaskCompleted", &id, &ok)
}

func (s *RpcInterface) CompleteTask(args *task.TaskResult, reply *bool) error {
	if isLeader, _ := s.state.IsLeader(); !isLeader {
		return errors.New("Not Leader")
	}

	err := s.state.CompleteTask(args.TaskID, args.WorkerID)
	if err != nil {
		println(err.Error())
		*reply = false
		return err
	}

	s.workers.TaskCompleted(args.WorkerID)
	*reply = true

	go s.notifyTaskCompleted(args.TaskID)
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

func (s *RpcInterface) RegisterWorker(args *string, reply *RegistrationReply) error {
	isLeader, addr := s.state.IsLeader()
	if !isLeader {
		*reply = RegistrationReply{IsLeader: false, Addr: addr, ID: ""}
		// fmt.Printf("%+v\n", reply)
		return nil
	}

	println("new worker")
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
	*reply = RegistrationReply{IsLeader: true, Addr: addr, ID: id}
	return nil
}

func (s *RpcInterface) RegisterScheduler(reg *raft.Server, res *bool) error {
	err := s.state.RegisterScheduler(string(reg.ID), string(reg.Address))
	return err
}

func (s *RpcInterface) Ping(workerID *string, reply *bool) error {
	*reply = true
	return nil
}
