package worker

import (
	"net/rpc"
	"sync"
)

type WorkerState int

const (
	IDLE WorkerState = iota
	BUSY
)

type Worker struct {
	ID              string
	SchedulerAddr   string
	state           WorkerState
	jobID           string
	schedulerClient *rpc.Client
	mu              sync.Mutex
}
