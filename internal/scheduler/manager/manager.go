package manager

import (
	"dssim/internal/scheduler/state"
	"dssim/internal/scheduler/worker"
	"log/slog"
)

type Manager struct {
	workers      worker.WorkerManager
	RpcInterface Scheduler
	state        *state.SchedulerState
}

func NewManager(
	workers worker.WorkerManager,
	state *state.SchedulerState,
) Manager {
	return Manager{
		workers:      workers,
		RpcInterface: NewScheduler(workers, state),
		state:        state,
	}
}

func (s *Manager) Run() {
	s.RpcInterface.NotifiyGenerator()

	for {
		task, available := s.state.NextTask()
		if !available {
			continue
		}

		worker := s.workers.GetWorker()
		s.state.AssignedTask(task, worker.ID)
		ok := worker.AssignTask(task)
		slog.Info(
			"assigned",
			"type",
			"task",
			"taskID",
			task.ID,
			"workerID",
			worker.ID,
			"success",
			ok,
		)
	}
}
