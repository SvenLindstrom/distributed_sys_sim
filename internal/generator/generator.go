package generator

import (
	"container/heap"
	"dssim/internal/network"
	"dssim/internal/task"
	"log"
	"log/slog"
	"os"
	"time"
)

type Generator struct {
	duration  int
	interval  Interval
	timout    time.Duration
	taskTable TaskTable
	start     chan bool
	client    *network.RPCClient
	sem       chan struct{}
	sem2      chan struct{}
}

func NewGenerator(duration int, interval Interval, timeout int) Generator {
	g := Generator{
		duration,
		interval,
		time.Duration(timeout),
		NewTaskTable(time.Duration(timeout)),
		make(chan bool),
		nil,
		make(chan struct{}, 100),
		make(chan struct{}, 100),
	}

	return g
}

func (g *Generator) TaskCompleted(taskId *string, ok *bool) error {
	// g.taskTable.TaskDone(*taskId)
	t := g.taskTable.RemoveTask(taskId)
	*ok = true
	var count int
	if t != nil {
		count = t.reSubCount
	} else {
		count = -1
	}
	slog.Info(
		"done",
		"type",
		"task",
		"taskID",
		*taskId,
		"retryCount",
		count,
	)
	return nil
}

func connectToScheduler(address string) (network.RPCClient, error) {
	port := os.Getenv("SCHEDULER_PORT")
	client, err := network.RealRPCDialer().Dial(address + ":" + port)

	return client, err
}

func (g *Generator) ReadyForWork(address *string, ok *bool) error {

	client, err := connectToScheduler(*address)

	if err != nil {
		println(err.Error())
		log.Fatal(err.Error())
		return err
	}
	println("connected to scheduler")

	if g.client == nil {
		g.client = &client
		go g.Run()
		go g.checkReSub()
	} else {
		println("new client")
		g.client = &client
	}

	println("task generation started")

	*ok = true
	return nil
}

type Item struct {
	task    *TaskRec
	retryAt time.Time
	index   int
}

type PriorityQueue []*Item

func (pq PriorityQueue) Len() int {
	return len(pq)
}

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].retryAt.Before(pq[j].retryAt)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq PriorityQueue) Peek() *Item {
	if pq.Len() == 0 {
		return nil
	}

	i := pq[0]
	return i
}

func (pq *PriorityQueue) update(item *Item, value *TaskRec, priority time.Time) {
	item.task = value
	item.retryAt = priority
	heap.Fix(pq, item.index)
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

func (g *Generator) reSend(t *Item) {

	var ok bool
	err := (*g.client).Call("Scheduler.CreateTask", t.task.Task, &ok)
	if err != nil {
		g.taskTable.mu.Lock()
		defer g.taskTable.mu.Unlock()
		heap.Push(g.taskTable.pq, t)
		return
	}

	g.taskTable.mu.Lock()
	defer g.taskTable.mu.Unlock()
	t.task.reSubCount += 1
	println("-------------------------")
	slog.Info(
		"re-submitted",
		"type",
		"task",
		"taskID",
		t.task.Task.ID,
		"retryCount",
		t.task.reSubCount,
	)

	newTime := time.Now().Add(g.timout * time.Second)
	t.retryAt = newTime

	heap.Push(g.taskTable.pq, t)

}

func (g *Generator) checkReSub() {
	for {
		pq := g.taskTable.pq
		t := pq.Peek()

		if t == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		now := time.Now()
		if now.Before(t.retryAt) {
			time.Sleep(t.retryAt.Sub(now))
		}

		if !g.taskTable.Exists(t.task.Task.ID) {
			heap.Pop(pq)
			continue
		}

		var ok bool
		err := (*g.client).Call("Scheduler.CreateTask", t.task.Task, &ok)
		if err != nil {
			continue
		}

		g.taskTable.mu.Lock()
		t.task.reSubCount += 1
		slog.Info(
			"re-submitted",
			"type",
			"task",
			"taskID",
			t.task.Task.ID,
			"retryCount",
			t.task.reSubCount,
		)

		newTime := time.Now().Add(g.timout * time.Second)
		pq.update(t, t.task, newTime)

		g.taskTable.mu.Unlock()
	}
}

func (g *Generator) send(t *task.Task) error {
	c := *g.client
	var ok bool
	err := c.Call("Scheduler.CreateTask", t, &ok)
	if err != nil {
		return err
	}
	g.taskTable.TaskSubmitted(*t)
	slog.Info(
		"submitted",
		"type",
		"task",
		"taskID",
		t.ID,
		"retryCount",
		0,
	)
	return nil
}

func (g *Generator) Run() {
	for {
		t, err := task.CreateTask(g.duration)
		if err != nil {
			continue
		}

		g.sem <- struct{}{}

		go func(t *task.Task) {
			defer func() { <-g.sem }()
			g.send(t)
		}(t)

		time.Sleep(g.interval.GetInterval())
	}
}
