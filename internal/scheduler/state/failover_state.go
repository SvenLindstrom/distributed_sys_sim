package state

import (
	"dssim/internal/task"
	"errors"
	"time"

	"github.com/hashicorp/raft"
)

type FailOverState struct {
	raft *raft.Raft
	data *SchedulerState
}

func NewFailOverState(raft *raft.Raft, ss *SchedulerState) FailOverState {
	return FailOverState{raft, ss}
}

func (rs *FailOverState) IsLeader() error {
	if rs.raft.State() != raft.Leader {
		addrs, _ := rs.raft.LeaderWithID()
		println("redirecting to Leader")
		addrsStr := string(addrs)
		if addrsStr == "" {
			return errors.New(" ")
		}
		return errors.New(addrsStr)
	}
	return nil
}

func (rs *FailOverState) RegisterScheduler(id string, address string) error {
	err := rs.IsLeader()
	if err != nil {
		return err
	}

	f := rs.raft.AddVoter(raft.ServerID(id), raft.ServerAddress(address), 0, 5*time.Second)

	if f.Error() != nil {
		println(f.Error().Error())
	}

	return nil
}

func (ss *FailOverState) AddTask(task *task.Task) error {
	return ss.data.AddTask(task)
}

func (ss *FailOverState) NextTask() (*task.Task, bool) {
	return ss.data.NextTask()
}

func (ss *FailOverState) AssignedTask(task *task.Task, workerId string) error {
	return ss.data.AssignedTask(task, workerId)
}

func (ss *FailOverState) CompleteTask(taskId string, workerId string) error {
	return ss.data.CompleteTask(taskId, workerId)
}
