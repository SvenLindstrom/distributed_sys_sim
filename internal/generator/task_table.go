package generator

import (
	"dssim/internal/task"
	"sync"
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
	mu    sync.Mutex
}

func (jt *TaskTable) Missing(id string) bool {
	jt.mu.Lock()
	missing := jt.Tasks[id] != nil
	jt.mu.Unlock()
	return missing
}

func (jt *TaskTable) GetCopy() map[string]*TaskRec {
	jt.mu.Lock()
	taskCopy := make(map[string]*TaskRec, len(jt.Tasks))
	for id, job := range jt.Tasks {
		taskCopy[id] = job
	}
	jt.mu.Unlock()
	return taskCopy
}

func (jt *TaskTable) TaskDone(id string) {
	jt.mu.Lock()
	jt.Tasks[id].done = true
	jt.mu.Unlock()
}

func (jt *TaskTable) RemoveTask(taskID *string) {
	jt.mu.Lock()
	delete(jt.Tasks, *taskID)
	jt.mu.Unlock()
}

func (jt *TaskTable) ReSubmit(taskRec *TaskRec) {
	taskRec.sub_count += 1
	taskRec.Submit_time = time.Now()
	jt.mu.Lock()
	jt.Tasks[taskRec.Task.ID] = taskRec
	jt.mu.Unlock()
}

func (jt *TaskTable) TaskSubmitted(task task.Task) {
	taskRec := TaskRec{task, time.Now(), 1, false}
	jt.mu.Lock()
	jt.Tasks[task.ID] = &taskRec
	jt.mu.Unlock()
}
