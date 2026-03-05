package main

import (
	"dssim/internal/misc"
	"dssim/internal/scheduler/scheduler"
	"dssim/internal/scheduler/worker"
	"log"
	"net/http"
	"net/rpc"
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

	workerManager := worker.NewWorkerManager(WorkerQueueSize)

	scheduler := scheduler.NewSchdular(workerManager, JobQueueSize)

	println("scheduler created")
	go scheduler.Run()

	println("scheduler started")
	rpc.Register(&scheduler)
	rpc.HandleHTTP()

	println("rcp server ready")
	http.ListenAndServe(":8080", nil)
}
