package generator

import (
	"dssim/internal/network"
	"dssim/internal/task"
	"fmt"
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
		go g.checkReSub()
	} else {
		println("new client")
		g.client = &client
	}

	println("task generation started")

	*ok = true
	return nil
}

func (g *Generator) checkReSub() {
	for {
		c := *g.client
		var ok bool

		list := g.taskTable.GetCopy()
		println(len(list))
		for _, t := range list {
			if time.Now().Sub(t.Submit_time) > g.timout*time.Second {
				if !g.taskTable.Missing(t.Task.ID) {
					continue
				}
				fmt.Printf("%+v, time now %+v \n", t, time.Now())
				err := c.Call("Scheduler.CreateTask", t.Task, &ok)
				if err != nil {
					break
				}
				println("re doooooone")
				g.taskTable.ReSubmit(t)

				slog.Info(
					"re-submitted",
					"type",
					"task",
					"taskID",
					t.Task.ID,
				)
			}
		}
		time.Sleep(1 * time.Second)
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
	)
	return nil
}

func (g *Generator) Run() {
	sem := make(chan struct{}, 100)
	for {
		// c := *g.client

		t, err := task.CreateTask(g.duration)
		if err != nil {
			continue
		}

		// var ok bool
		// err = c.Call("Scheduler.CreateTask", task, &ok)
		// if err != nil {
		// 	continue
		// }
		//
		// g.taskTable.TaskSubmitted(*task)

		sem <- struct{}{}

		go func(t *task.Task) {
			defer func() { <-sem }()
			g.send(t)
		}(t)

		// go g.send(task)

		time.Sleep(g.interval.GetInterval())
		// time.Sleep(10 * time.Second)
	}
}
