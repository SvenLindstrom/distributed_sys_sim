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
}

type SchedulerState struct {
	pendding []*task.Task
	running  map[string]*task.Task
	reserved map[string]bool
	mu       sync.Mutex
}

func NewSchedulerState(taskQueueSize int) *SchedulerState {
	var list []*task.Task
	taskMap := make(map[string]*task.Task)
	reservedMap := make(map[string]bool)
	return &SchedulerState{pendding: list, running: taskMap, reserved: reservedMap}
}

func (ss *SchedulerState) AddTask(task *task.Task) error {
	ss.mu.Lock()
	ss.pendding = append(ss.pendding, task)
	ss.mu.Unlock()
	return nil
}

func (ss *SchedulerState) NextTask() *task.Task {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	// if len(ss.pendding) != 0 {
	// 	return nil
	// }

	for _, t := range ss.pendding {
		if !ss.reserved[t.ID] {
			println(len(ss.pendding))
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
	ss.running[workerId] = task
	return nil
}

func (ss *SchedulerState) CompleteTask(taskId string, workerId string) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if task := ss.running[workerId]; task != nil {
		if task.ID == taskId {
			delete(ss.running, workerId)
		} else {
			println("AAAAAAAAAAAAAAAAAAAAA")
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
