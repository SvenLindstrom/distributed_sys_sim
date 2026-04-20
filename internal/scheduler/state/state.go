package state

import (
	"dssim/internal/task"
	"errors"
)

type State interface {
	AddTask(task *task.Task) error
	NextTask() (*task.Task, bool)
	AssignedTask(task *task.Task, workerId string) error
	CompleteTask(taskId string, workerId string) error
	RegisterScheduler(id string, address string) error
	IsLeader() (bool, string)
}

type SchedulerState struct {
	pendding []*task.Task
	running  map[string]*task.Task
}

func NewSchedulerState(taskQueueSize int) *SchedulerState {
	var list []*task.Task
	taskMap := make(map[string]*task.Task)
	return &SchedulerState{list, taskMap}
}

func (ss *SchedulerState) AddTask(task *task.Task) error {
	ss.pendding = append(ss.pendding, task)
	return nil
}

func (ss *SchedulerState) NextTask() (*task.Task, bool) {
	if len(ss.pendding) == 0 {
		return &task.Task{}, false
	}
	return ss.pendding[0], true
}

func (ss *SchedulerState) AssignedTask(task *task.Task, workerId string) error {
	ss.pendding = ss.pendding[1:]
	ss.running[workerId] = task
	return nil
}

func (ss *SchedulerState) CompleteTask(taskId string, workerId string) error {
	if task := ss.running[workerId]; task != nil {
		if task.ID == taskId {
			delete(ss.running, workerId)
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
