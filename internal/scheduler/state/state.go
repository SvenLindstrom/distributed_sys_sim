package state

import (
	"dssim/internal/task"
	"errors"
	"sync"
)

type State interface {
	AddTask(task *task.Task) error
	NextTask() *task.Task
	AssignedTask(task *task.Task, workerId string) error
	CompleteTask(taskId string, workerId string) error
	RegisterScheduler(id string, address string) error
	IsLeader() (bool, string)
	ReAddTask(task *task.Task) error
}

type SchedulerState struct {
	pendding []*task.Task
	running  map[string]string
	reserved map[string]bool
	mu       sync.Mutex
}

func NewSchedulerState() *SchedulerState {
	var list []*task.Task
	taskMap := make(map[string]string)
	reservedMap := make(map[string]bool)
	return &SchedulerState{pendding: list, running: taskMap, reserved: reservedMap}
}

func (ss *SchedulerState) AddTask(task *task.Task) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if ss.running[task.ID] != "" {
		return nil
	}
	for _, t := range ss.pendding {
		if t.ID == task.ID {
			return nil
		}
	}
	ss.pendding = append(ss.pendding, task)
	return nil
}

func (ss *SchedulerState) NextTask() *task.Task {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	for _, t := range ss.pendding {
		if !ss.reserved[t.ID] {
			ss.reserved[t.ID] = true
			return t
		}
	}

	return nil
}

func (ss *SchedulerState) AssignedTask(task *task.Task, workerId string) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	for i, t := range ss.pendding {
		if t.ID == task.ID {
			ss.pendding = append(ss.pendding[:i], ss.pendding[i+1:]...)
			break
		}
	}
	// ss.pendding = ss.pendding[1:]
	delete(ss.reserved, task.ID)
	ss.running[task.ID] = workerId
	return nil
}

func (ss *SchedulerState) CompleteTask(taskId string, workerId string) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if worker := ss.running[taskId]; worker != "" {
		if worker == workerId {
			delete(ss.running, taskId)
		}
		return nil
	}
	return errors.New("unknown worker")
}

func (ss *SchedulerState) RegisterScheduler(id string, address string) error {
	return nil
}

func (ss *SchedulerState) IsLeader() (bool, string) {
	return true, ""
}
func (ss *SchedulerState) ReAddTask(task *task.Task) error {
	return ss.AddTask(task)
}
