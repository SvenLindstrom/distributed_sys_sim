package scheduler

import (
	"dssim/internal/scheduler/state"
	"dssim/internal/scheduler/worker"
	"dssim/internal/task"
	"log/slog"
	"time"
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

func (s *Scheduler) sendTask(t *task.Task, w *worker.Worker) {
	err := s.state.AssignedTask(t, w.ID)
	if err != nil {
		return
	}

	ok := w.AssignTask(t)
	slog.Info(
		"assigned",
		"type",
		"task",
		"taskID",
		t.ID,
		"workerID",
		w.ID,
		"success",
		ok,
	)
}

func (s *Scheduler) Run() {
	println("scheduler running")
	for s.RunLoop {
		t := s.state.NextTask()
		if t == nil {
			time.Sleep(1 * time.Millisecond)
			continue
		}

		w := s.workers.GetWorker()

		go s.sendTask(t, w)
	}
}
