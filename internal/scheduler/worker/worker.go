package worker

import (
	"dssim/internal/network"
	"dssim/internal/task"
	"log"
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
	ID     string
	state  WorkerState
	task   string
	client network.RPCClient
}

func (w *Worker) TaskFinished() {
	w.state = READY
	w.task = ""
}

func (w *Worker) AssignTask(task *task.Task) bool {
	var ok bool
	err := w.client.Call("Worker.AssignTask", &task, &ok)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	if ok {
		w.task = task.ID
		w.state = BUSY
	}
	return ok
}
