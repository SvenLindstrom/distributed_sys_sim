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

func (rs *RaftState) RegisterScheduler(id string, address string) error {
	println("adding new Follower")
	if rs.raft.State() != raft.Leader {
		addrs, _ := rs.raft.LeaderWithID()
		println("redirecting to Leader")
		println(addrs)
		// addrsStr := string(addrs)
		// return errors.New(string(addrs))
		return errors.New(" ")
		// return errors.New("TTTT")
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

	println("task recived")
	cmd := RaftCMD{Op: "ADD", Task: task}
	data, err := json.Marshal(cmd)

	if err != nil {
		panic(err)
	}
	var cmdd RaftCMD
	json.Unmarshal(data, &cmdd)

	rs.raft.Apply(data, 5*time.Second)
	return nil
}

func (rs *RaftState) NextTask() (*task.Task, bool) {
	if len(rs.data.pendding) == 0 {
		return &task.Task{}, false
	}
	return rs.data.pendding[0], true
}

func (rs *RaftState) AssignedTask(task *task.Task, workerId string) error {
	println("task given")
	cmd := RaftCMD{Op: "ASSIGN", Task: task, WorkerId: workerId}
	data, err := json.Marshal(cmd)

	if err != nil {
		panic(err)
	}

	rs.raft.Apply(data, 5*time.Second)
	return nil
}

func (rs *RaftState) CompleteTask(taskId string, workerId string) error {
	println("task done")
	cmd := RaftCMD{Op: "COMP", WorkerId: workerId, TaskId: taskId}
	data, err := json.Marshal(cmd)

	if err != nil {
		panic(err)
	}

	rs.raft.Apply(data, 5*time.Second)
	return nil
}

func (rs *SchedulerState) Apply(log *raft.Log) interface{} {
	var cmd RaftCMD
	json.Unmarshal(log.Data, &cmd)

	switch cmd.Op {
	case "ADD":
		rs.AddTask(cmd.Task)
	case "ASSIGN":
		rs.AssignedTask(cmd.Task, cmd.WorkerId)
	case "COMP":
		rs.CompleteTask(cmd.TaskId, cmd.WorkerId)
	}

	return nil
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
