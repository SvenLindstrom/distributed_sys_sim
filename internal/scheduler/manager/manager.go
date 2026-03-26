package manager

import (
	"dssim/internal/scheduler/state"
	"dssim/internal/scheduler/worker"
	"log/slog"
)

type Manager struct {
	workers      worker.WorkerManager
	RpcInterface Scheduler
	state        state.State
	RunLoop      bool
}

func NewManager(
	workers worker.WorkerManager,
	state state.State,
) Manager {
	return Manager{
		workers:      workers,
		RpcInterface: NewScheduler(workers, state),
		state:        state,
		RunLoop:      true,
	}
}

func (s *Manager) Run() {

	println("scheduler running")
	for s.RunLoop {
		task, available := s.state.NextTask()
		if !available {
			continue
		}

		// println(task.ID)

		worker := s.workers.GetWorker()

		err := s.state.AssignedTask(task, worker.ID)

		if err != nil {
			println(err.Error())
		}

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
