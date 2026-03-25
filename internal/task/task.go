package task

import "dssim/internal/misc"

type Task struct {
	ID       string
	Duration int
}

type NewTask struct {
	Duration int
}

type TaskResult struct {
	TaskID   string
	WorkerID string
	Status   string
}

func CreateTask(duration int) (*Task, error) {
	id, err := misc.GenID()
	if err != nil {
		return &Task{}, err
	}
	task := &Task{id, duration}

	return task, nil
}
