package scheduler

import (
	"dssim/internal/scheduler/manager"
	"dssim/internal/scheduler/state"
	"dssim/internal/scheduler/worker"
)

func NewSchdular(workerQueueSize int, taskQueueSize int) manager.Manager {
	workerManager := worker.NewWorkerManager(workerQueueSize)
	return manager.NewManager(workerManager, state.NewSchedulerState(taskQueueSize))
}
