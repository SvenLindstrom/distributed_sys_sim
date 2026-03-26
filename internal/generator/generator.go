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
	interval  int
	timout    time.Duration
	taskTable TaskTable
	start     chan bool
	client    *network.RPCClient
}

func NewGnerator(duration int, interval int, timeout int) Generator {
	g := Generator{
		duration,
		interval,
		time.Duration(timeout),
		TaskTable{make(map[string]*TaskRec)},
		make(chan bool),
		nil,
	}

	return g
}

func (g *Generator) TaskCompleted(taskId *string, ok *bool) error {
	g.taskTable.TaskDone(*taskId)
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
		go g.Run(g.client)
	} else {
		g.client = &client
	}

	println("task generation started")

	*ok = true
	return nil
}

func (g *Generator) Run(client *network.RPCClient) {
	for {
		c := *client

		task, err := task.CreateTask(g.duration)
		if err != nil {

			println(err.Error())
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

		for _, task := range g.taskTable.Tasks {

			if task.done {
				g.taskTable.RemoveTask(task)
			} else if time.Now().Sub(task.Submit_time) > g.timout*time.Second {
				c.Call("Scheduler.CreateTask", task, &ok)
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
		time.Sleep(time.Duration(g.interval) * time.Millisecond)
	}
}
