package worker

import (
	"dssim/internal/network"
)

type WorkerManager interface {
	NewWorker(address string, id string) bool
	JobCompleted(workerID string)
	GetWorker() *Worker
}

type WorkerManagerImple struct {
	workers   map[string]*Worker
	ready     chan *Worker
	rpcDialer network.RPCDialer
}

func NewWorkerManager(queueSize int) WorkerManager {
	workers := make(chan *Worker, queueSize)
	wmap := make(map[string]*Worker)
	dialer := network.RealRPCDialer{}
	return &WorkerManagerImple{wmap, workers, &dialer}
}

func (w *WorkerManagerImple) NewWorker(address string, id string) bool {
	client, err := w.rpcDialer.Dial(address)
	if err != nil {
		println(err.Error())
		return false
	}
	worker := &Worker{address, id, IDLE, "", client}
	w.workers[worker.ID] = worker
	w.ready <- worker
	return true
}

func (w *WorkerManagerImple) JobCompleted(workerID string) {
	worker := w.workers[workerID]
	worker.JobFinished()
	w.ready <- worker
}

func (w *WorkerManagerImple) GetWorker() *Worker {
	status := OFFLINE
	var worker *Worker
	for status == OFFLINE {
		worker = <-w.ready
		status = worker.state
	}
	return worker
}
