package generator

import (
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
}

func NewGenerator(duration int, interval Interval, timeout int) Generator {
	g := Generator{
		duration,
		interval,
		time.Duration(timeout),
		TaskTable{Tasks: make(map[string]*TaskRec)},
		make(chan bool),
		nil,
	}

	return g
}

func (g *Generator) TaskCompleted(taskId *string, ok *bool) error {
	// g.taskTable.TaskDone(*taskId)
	g.taskTable.RemoveTask(taskId)
	*ok = true
	slog.Info(
		"done",
		"type",
		"task",
		"taskID",
		*taskId,
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
	} else {
		println("new client")
		g.client = &client
	}

	println("task generation started")

	*ok = true
	return nil
}

func (g *Generator) checkReSub() {
	c := *g.client
	var ok bool

	for _, task := range g.taskTable.GetCopy() {
		if time.Now().Sub(task.Submit_time) > g.timout*time.Second {
			err := c.Call("Scheduler.CreateTask", task, &ok)
			if err != nil {
				continue
			}
			g.taskTable.ReSubmit(task)

			slog.Info(
				"re-submitted",
				"type",
				"task",
				"taskID",
				task.Task.ID,
			)
		}
	}
}

func (g *Generator) Run() {
	for {
		c := *g.client

		task, err := task.CreateTask(g.duration)
		if err != nil {
			continue
		}

		var ok bool

		err = c.Call("Scheduler.CreateTask", task, &ok)
		if err != nil {
			continue
		}

		g.taskTable.TaskSubmitted(*task)
		slog.Info(
			"submitted",
			"type",
			"task",
			"taskID",
			task.ID,
		)

		g.checkReSub()

		time.Sleep(g.interval.GetInterval())
	}
}
