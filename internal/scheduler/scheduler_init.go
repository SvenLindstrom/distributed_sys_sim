package scheduler

import (
	"dssim/internal/scheduler/state"
	"dssim/internal/scheduler/worker"
	"log"
	"os"
)

func InitScheduler(workerQueueSize int, taskQueueSize int) {
	permutation := os.Getenv("SETUP")
	if permutation == ""{
		permutation = "BASE"
	}

	switch permutation {
	case "BASE":
		println("In Base")
		leader := os.Getenv("LEADER")
		if leader == "" {
			os.Exit(0)
		}
		NewSchedular(workerQueueSize, taskQueueSize)
	case "ELECTION-REPLICATION":
		println("In Raft")
		NewSchedulerRaft(workerQueueSize, taskQueueSize)
	case "FAILOVER":
		println("In Failover")
		newSchedulerFailover(workerQueueSize, taskQueueSize)
	}
}

func NewSchedular(workerQueueSize int, taskQueueSize int) Scheduler {
	workerManager := worker.NewWorkerManager(workerQueueSize)
	state := state.NewSchedulerState(taskQueueSize)

	rpc := NewRPC(workerManager, state)
	s := NewScheduler(workerManager, state)

	go s.Run()
	rpc.Start()
	rpc.NotifiyGenerator()

	return s
}

func leaderChange(leader <-chan bool, s *Scheduler, rpc *RpcInterface) {
	for {
		amLeader := <-leader

		if amLeader {
			s.RunLoop = true
			go s.Run()
			rpc.NotifiyGenerator()
		} else {
			s.RunLoop = false
		}
	}
}

func newSchedulerFailover(workerQueueSize int, taskQueueSize int) Scheduler {
	workerManager := worker.NewWorkerManager(workerQueueSize)

	fsm := state.NewSchedulerState(taskQueueSize)

	r, err := NewRaft(fsm)

	if err != nil {
		log.Fatal(err.Error())
	}

	leader := os.Getenv("LEADER")

	if leader == "true" {
		r.BootStrap()
	} else {
		r.RegisterWithLeader()
	}

	state := state.NewFailOverState(r.raft, fsm)
	s := NewScheduler(workerManager, &state)

	RpcInterface := NewRPC(workerManager, &state)
	go leaderChange(r.raft.LeaderCh(), &s, &RpcInterface)

	println("rcp server ready")
	RpcInterface.Start()

	return s
}

func NewSchedulerRaft(workerQueueSize int, taskQueueSize int) Scheduler {
	workerManager := worker.NewWorkerManager(workerQueueSize)

	fsm := state.NewSchedulerState(taskQueueSize)

	r, err := NewRaft(fsm)

	if err != nil {
		log.Fatal(err.Error())
	}

	leader := os.Getenv("LEADER")

	if leader == "true" {
		r.BootStrap()
	} else {
		r.RegisterWithLeader()
	}

	println("returning manager")

	state := state.NewRaftState(r.raft, fsm)
	s := NewScheduler(workerManager, &state)

	RpcInterface := NewRPC(workerManager, &state)

	go leaderChange(r.raft.LeaderCh(), &s, &RpcInterface)

	println("rcp server ready")
	RpcInterface.Start()

	return s
}
