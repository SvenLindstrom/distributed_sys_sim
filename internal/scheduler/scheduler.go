package scheduler

import (
	"dssim/internal/network"
	"dssim/internal/scheduler/manager"
	"dssim/internal/scheduler/state"
	"dssim/internal/scheduler/worker"
	"log"
	"os"
	"time"

	"github.com/hashicorp/raft"
)

func NewSchdular(workerQueueSize int, taskQueueSize int) manager.Manager {
	name := os.Getenv("HOSTNAME")
	workerManager := worker.NewWorkerManager(workerQueueSize)

	m := manager.NewManager(workerManager, state.NewSchedulerState(taskQueueSize))
	m.RpcInterface.NotifiyGenerator(name)
	go m.Run()

	return m
}

func leaderChange(leader <-chan bool, m manager.Manager) {
	name := os.Getenv("HOSTNAME")
	for {
		amLeader := <-leader

		if amLeader {
			m.RpcInterface.NotifiyGenerator(name)
			go m.Run()
		} else {
			m.RunLoop = false
		}
	}
}

func NewSchedulerRaft(workerQueueSize int, taskQueueSize int) manager.Manager {
	println("In Raft")
	workerManager := worker.NewWorkerManager(workerQueueSize)

	name := os.Getenv("HOSTNAME")
	conf := raft.DefaultConfig()
	conf.LocalID = raft.ServerID(name)
	println(name)

	addr := name + ":6000"

	fsm := state.NewSchedulerState(taskQueueSize)
	store := raft.NewInmemStore()
	snapStore := raft.NewInmemSnapshotStore()
	transport, err := raft.NewTCPTransport(
		addr,
		nil,
		3,
		10*time.Second,
		os.Stdout,
	)

	if err != nil {
		println(err.Error())
	}

	println("starting raft node")

	r, err := raft.NewRaft(conf, fsm, store, store, snapStore, transport)

	if err != nil {
		log.Fatal(err.Error())
	}

	leader := os.Getenv("LEADER")

	server := raft.Server{ID: raft.ServerID(name),
		Address: transport.LocalAddr(),
	}

	if leader == "true" {
		println("Is Leader")
		configuration := raft.Configuration{
			Servers: []raft.Server{
				server,
			},
		}
		println(transport.LocalAddr())
		println("--- bootstraping cluster ---")
		r.BootstrapCluster(configuration)
	} else {
		println("Is Follower")
		port := ":" + os.Getenv("SCHEDULER_PORT")
		for {
			c, err := network.RealRPCDialer().Dial("Scheduler" + port)
			if err != nil {
				continue
			}
			var res bool
			for {
				err := c.Call("Scheduler.RegisterScheduler", &server, &res)

				if err != nil {
					println("got error")

					println(err.Error())
					time.Sleep(1 * time.Second)
					continue
				}
				break
			}
			println("connected to leader")
			break
		}
	}

	println("returning manager")

	state := state.NewRaftState(r, fsm)
	m := manager.NewManager(workerManager, &state)

	go leaderChange(r.LeaderCh(), m)

	return m
}
