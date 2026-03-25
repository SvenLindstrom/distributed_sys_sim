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
}

func NewGnerator(duration int, interval int, timeout int) Generator {
	return Generator{
		duration,
		interval,
		time.Duration(timeout),
		TaskTable{make(map[string]*TaskRec)},
	}
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

func (g *Generator) ReadyForWork(address *string, ok *bool) error {
	port := os.Getenv("SCHEDULER_PORT")
	dialer := network.RealRPCDialer{}

	client, err := dialer.Dial("scheduler:" + port)

	if err != nil {
		log.Fatal(err.Error())
	} else {
		println("connected to scheduler")
	}
	go g.Run(client)
	println("task generation started")

	*ok = true
	return nil
}

func (g *Generator) Run(client network.RPCClient) {
	for {
		task, err := task.CreateTask(g.duration)
		if err != nil {
			println("")
		}

		var ok bool

		client.Call("Scheduler.CreateTask", task, &ok)

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
				client.Call("Scheduler.CreateTask", task, &ok)
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
