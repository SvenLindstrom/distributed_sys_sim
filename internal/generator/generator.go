package generator

import (
	"dssim/internal/job"
	"dssim/internal/network"
	"log"
	"log/slog"
	"os"
	"time"
)

type Generator struct {
	duration int
	interval int
	timout   time.Duration
	jobTable JobTable
}

func NewGnerator(duration int, interval int, timeout int) Generator {
	return Generator{duration, interval, time.Duration(timeout), JobTable{make(map[string]JobRec)}}
}

func (g *Generator) JobCompleted(jobId *string, ok *bool) error {
	g.jobTable.JobDone(*jobId)
	*ok = true
	slog.Info(
		"job done",
		"type",
		"job",
		"jobID",
		*jobId,
	)
	return nil
}

func (g *Generator) ReadyForWork(address *string, ok *bool) error {
	port := os.Getenv("SCHEDULER_PORT")
	dialer := network.RealRPCDialer{}

	client, err := dialer.Dial("scheduler:" + port)

	if err != nil {
		log.Fatal(err.Error())
	}
	go g.Run(client)

	*ok = true
	return nil
}

func (g *Generator) Run(client network.RPCClient) {
	for {
		job, err := job.CreateJob(g.duration)
		if err != nil {
			println("")
		}

		var ok bool

		client.Call("Scheduler.CreateJob", job, &ok)

		g.jobTable.JobSubmitted(*job)
		slog.Info(
			"job submitted",
			"type",
			"job",
			"jobID",
			job.ID,
		)

		for _, job := range g.jobTable.Jobs {
			if time.Now().Sub(job.Submit_time) > g.timout*time.Second {
				client.Call("Scheduler.CreateJob", job, &ok)
				slog.Info(
					"job re-submitted",
					"type",
					"job",
					"jobID",
					job.Job.ID,
				)
			}
		}
		time.Sleep(time.Duration(g.interval) * time.Millisecond)
	}
}
