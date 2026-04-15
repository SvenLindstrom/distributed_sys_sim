package scheduler

import (
	"dssim/internal/network"
	"os"
	"time"

	"github.com/hashicorp/raft"
)

type Raft struct {
	raft   *raft.Raft
	server raft.Server
}

func (r *Raft) RegisterWithLeader() {
	println("Is Follower")
	port := ":" + os.Getenv("SCHEDULER_PORT")
	for {
		c, err := network.RealRPCDialer().Dial("Scheduler" + port)
		if err != nil {
			continue
		}
		var res bool
		for {
			err = c.Call("Scheduler.RegisterScheduler", &r.server, &res)

			if err != nil {
				println("got error")
				time.Sleep(1 * time.Second)
				continue
			}
			println("connected to leader")
			return
		}
	}
}

func (r *Raft) BootStrap() {
	println("--- bootstraping cluster ---")
	configuration := raft.Configuration{
		Servers: []raft.Server{
			r.server,
		},
	}
	r.raft.BootstrapCluster(configuration)
}

func NewRaft(fsm raft.FSM) (*Raft, error) {
	name := os.Getenv("HOSTNAME")

	conf := raft.DefaultConfig()
	conf.LocalID = raft.ServerID(name)
	println(name)

	addr := name + ":6000"

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

	server := raft.Server{ID: raft.ServerID(name),
		Address: transport.LocalAddr(),
	}

	r, err := raft.NewRaft(conf, fsm, store, store, snapStore, transport)

	if err != nil {
		return &Raft{}, nil
	}

	return &Raft{r, server}, err
}
