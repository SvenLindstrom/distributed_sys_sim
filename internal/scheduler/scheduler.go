package scheduler

import (
	"dssim/internal/scheduler/state"
	"dssim/internal/scheduler/worker"
	"log/slog"
)

type Scheduler struct {
	workers worker.WorkerManager
	state   state.State
	RunLoop bool
}

func NewScheduler(
	workers worker.WorkerManager,
	state state.State,
) Scheduler {
	return Scheduler{
		workers: workers,
		state:   state,
		RunLoop: true,
	}
}

func (s *Scheduler) Run() {
	println("scheduler running")
	for s.RunLoop {
		task, available := s.state.NextTask()
		if !available {
			continue
		}

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
