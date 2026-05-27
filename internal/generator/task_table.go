package generator

import (
	"container/heap"
	"dssim/internal/task"
	"sync"
	"time"
)

type TaskRec struct {
	Task       task.Task
	reSubCount int
}

type TaskTable struct {
	Tasks   map[string]*TaskRec
	mu      sync.Mutex
	pq      *PriorityQueue
	timeout time.Duration
}

func NewTaskTable(timeout time.Duration) TaskTable {
	pq := make(PriorityQueue, 0)
	heap.Init(&pq)
	tasks := make(map[string]*TaskRec)
	return TaskTable{Tasks: tasks, pq: &pq, timeout: timeout}
}

func (jt *TaskTable) Exists(id string) bool {
	jt.mu.Lock()
	defer jt.mu.Unlock()
	return jt.Tasks[id] != nil
}

// func (jt *TaskTable) GetCopy() map[string]*task.Task {
// jt.mu.Lock()
// defer jt.mu.Unlock()
// taskCopy := make(map[string]*task.Task, len(jt.Tasks))
// for id, job := range jt.Tasks {
// 	taskCopy[id] = job
// }
// return taskCopy
// }

func (jt *TaskTable) RemoveTask(taskID *string) *TaskRec {
	jt.mu.Lock()
	defer jt.mu.Unlock()
	t := jt.Tasks[*taskID]
	delete(jt.Tasks, *taskID)
	return t
}

func (jt *TaskTable) ReSubmit(taskRec *task.Task) {
	// jt.mu.Lock()
	// defer jt.mu.Unlock()
	// taskRec.Submit_time = time.Now()
	// jt.Tasks[taskRec.Task.ID] = taskRec
}

func (jt *TaskTable) TaskSubmitted(task task.Task) {
	jt.mu.Lock()
	defer jt.mu.Unlock()
	taskRec := TaskRec{Task: task, reSubCount: 0}
	jt.Tasks[task.ID] = &taskRec
	i := Item{task: &taskRec, retryAt: time.Now().Add(jt.timeout * time.Second)}
	heap.Push(jt.pq, &i)
}
