package state

import (
	"dssim/internal/task"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/hashicorp/raft"
)

type RaftState struct {
	raft *raft.Raft
	data *SchedulerState
}

type RaftCMD struct {
	Op       string
	Task     *task.Task
	WorkerId string
	TaskId   string
}

func NewRaftState(raft *raft.Raft, ss *SchedulerState) RaftState {
	return RaftState{raft, ss}
}

func (rs *RaftState) IsLeader() (bool, string) {
	isLeader := rs.raft.State() == raft.Leader
	addrs, _ := rs.raft.LeaderWithID()
	addrsStr := string(addrs)

	return isLeader, addrsStr
}

func (rs *RaftState) RegisterScheduler(id string, address string) error {
	isLeader, _ := rs.IsLeader()
	if !isLeader {
		return errors.New("not leader")
	}

	f := rs.raft.AddVoter(raft.ServerID(id), raft.ServerAddress(address), 0, 5*time.Second)

	if f.Error() != nil {
		println(f.Error().Error())
	}

	return nil
}

func (rs *RaftState) AddTask(task *task.Task) error {

	if rs.raft.State() != raft.Leader {
		addrs, _ := rs.raft.LeaderWithID()
		return errors.New(string(addrs))
	}

	cmd := RaftCMD{Op: "ADD", Task: task}
	data, err := json.Marshal(cmd)

	if err != nil {
		panic(err)
	}
	var cmdd RaftCMD
	json.Unmarshal(data, &cmdd)

	rs.raft.Apply(data, 5*time.Second).Error()
	return nil
}

func (rs *RaftState) NextTask() *task.Task {
	return rs.data.NextTask()
}

func (rs *RaftState) AssignedTask(task *task.Task, workerId string) error {
	cmd := RaftCMD{Op: "ASSIGN", Task: task, WorkerId: workerId}
	data, err := json.Marshal(cmd)

	if err != nil {
		panic(err)
	}
	
	rs.raft.Apply(data, 5*time.Second).Error()
	return nil
}

func (rs *RaftState) CompleteTask(taskId string, workerId string) error {
	cmd := RaftCMD{Op: "COMP", WorkerId: workerId, TaskId: taskId}
	data, err := json.Marshal(cmd)

	if err != nil {
		panic(err)
	}

	rs.raft.Apply(data, 5*time.Second).Error()
	return nil
}

func (rs *SchedulerState) Apply(log *raft.Log) interface{} {
	var cmd RaftCMD
	json.Unmarshal(log.Data, &cmd)

	var err error

	switch cmd.Op {
	case "ADD":
		err = rs.AddTask(cmd.Task)
	case "ASSIGN":
		err = rs.AssignedTask(cmd.Task, cmd.WorkerId)
	case "COMP":
		err = rs.CompleteTask(cmd.TaskId, cmd.WorkerId)
	}

	return err
}

type StateSnapshot struct {
	state SchedulerState
}

func (s *StateSnapshot) Persist(sink raft.SnapshotSink) error {
	data, err := json.Marshal(s.state)
	if err != nil {
		sink.Cancel()
		return err
	}

	_, err = sink.Write(data)

	if err != nil {
		sink.Cancel()
		return err
	}

	return sink.Close()
}

func (s *StateSnapshot) Release() {
}

func (rs *SchedulerState) Snapshot() (raft.FSMSnapshot, error) {

	pendding := make([]*task.Task, len(rs.pendding))
	copy(pendding, rs.pendding)

	running := make(map[string]*task.Task, len(rs.running))
	for k, v := range rs.running {
		running[k] = v
	}

	return &StateSnapshot{state: SchedulerState{pendding: pendding, running: running}}, nil
}

func (rs *SchedulerState) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	var state SchedulerState
	err := json.NewDecoder(rc).Decode(&state)

	if err != nil {
		return err
	}

	rs.pendding = state.pendding
	rs.running = state.running

	return nil
}
