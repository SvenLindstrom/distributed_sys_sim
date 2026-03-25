package generator

import (
	"dssim/internal/task"
	"time"
)

type TaskRec struct {
	Task        task.Task
	Submit_time time.Time
	sub_count   int
	done        bool
}

type TaskTable struct {
	Tasks map[string]*TaskRec
}

func (jt *TaskTable) TaskDone(id string) {
	jt.Tasks[id].done = true
}

func (jt *TaskTable) RemoveTask(task *TaskRec) {
	delete(jt.Tasks, task.Task.ID)
}

func (jt *TaskTable) ReSubmit(taskRec *TaskRec) {
	taskRec.sub_count += 1
	taskRec.Submit_time = time.Now()
	jt.Tasks[taskRec.Task.ID] = taskRec
}

func (jt *TaskTable) TaskSubmitted(task task.Task) {
	taskRec := TaskRec{task, time.Now(), 1, false}
	jt.Tasks[task.ID] = &taskRec
}
