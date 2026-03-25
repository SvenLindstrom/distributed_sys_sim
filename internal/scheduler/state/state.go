package state

import (
	"dssim/internal/task"
)

type SchedulerState struct {
	pendding []*task.Task
	running  map[string]*task.Task
}

func NewSchedulerState(taskQueueSize int) *SchedulerState {
	var list []*task.Task
	taskMap := make(map[string]*task.Task)
	return &SchedulerState{list, taskMap}
}

func (ss *SchedulerState) AddTask(task *task.Task) {
	ss.pendding = append(ss.pendding, task)
}

func (ss *SchedulerState) NextTask() (*task.Task, bool) {
	if len(ss.pendding) == 0 {
		return &task.Task{}, false
	}
	return ss.pendding[0], true
}

func (ss *SchedulerState) AssignedTask(task *task.Task, workerId string) {
	ss.pendding = ss.pendding[1:]
	ss.running[workerId] = task
}

func (ss *SchedulerState) CompleteTask(taskId string, workerId string) {
	task := ss.running[workerId]
	if task.ID == taskId {
		delete(ss.running, workerId)
	}
}
