package main

import (
	"log"
	"net/http"
	"net/rpc"
	"schdeuler/src/job"
	"schdeuler/src/misc"
	"schdeuler/src/scheduler"
	"schdeuler/src/worker"
)

const JobQueueSize = 3
const WorkerQueueSize = 3

func main() {

	f, err := misc.Loginit()

	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	println("logger ready")

	jobs := make(chan job.Job, JobQueueSize)

	workerManager := worker.NewWorkerManager(WorkerQueueSize)

	scheduler := scheduler.NewSchdular(workerManager)

	println("scheduler created")
	go scheduler.Run(jobs)

	println("scheduler started")
	rpc.Register(&scheduler)
	rpc.HandleHTTP()

	println("rcp server ready")
	http.ListenAndServe(":8080", nil)
}
