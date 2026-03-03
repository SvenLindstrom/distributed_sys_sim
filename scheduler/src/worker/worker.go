package worker

import (
	"schdeuler/src/job"
	"schdeuler/src/networkInterface"
)

type WorkerState int

const (
	READY WorkerState = iota
	IDLE
	BUSY
	OFFLINE
)

var stateName = map[WorkerState]string{
	READY:   "ready",
	IDLE:    "idle",
	BUSY:    "busy",
	OFFLINE: "offline",
}

type Worker struct {
	IP     string
	ID     string
	state  WorkerState
	job    string
	client networkinterface.RPCClient
}

func (w *Worker) JobFinished() {
	w.state = READY
	w.job = ""
}

func (w *Worker) AssignJob(job job.Job) bool {
	var ok bool
	w.client.Call("Worker.AssignJob", &job, &ok)
	if ok {
		w.job = job.ID
		w.state = BUSY
	}
	return ok
}
