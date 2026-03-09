package main

import (
	"dssim/internal/fault"
	"dssim/internal/misc"
	"dssim/internal/scheduler/scheduler"
	"dssim/internal/scheduler/worker"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
)

func main() {

	f, err := misc.Loginit("scheduler")

	if err != nil {
		log.Fatal(err.Error())
	}
	defer f.Close()

	port := ":" + os.Getenv("SCHEDULER_PORT")
	jobQueueSize, err := strconv.Atoi(os.Getenv("JOB_Q_SIZE"))
	if err != nil {
		log.Println("job q size not set")
		jobQueueSize = 3
	}
	workerQueueSize, err := strconv.Atoi(os.Getenv("WORKER_Q_SIZE"))
	if err != nil {
		log.Println("worker q size not set")
		workerQueueSize = 1
	}

	println("logger ready")

	workerManager := worker.NewWorkerManager(workerQueueSize)

	scheduler := scheduler.NewSchdular(workerManager, jobQueueSize)

	println("scheduler created")
	go scheduler.Run()
	println("scheduler started")

	// program fault injection
	duration := os.Getenv("SCHEDULER_CRASH")
	fault.InjectCrash(duration)
	println("fault injector started")

	rpc.Register(&scheduler)
	rpc.HandleHTTP()

	println("rcp server ready")
	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
}
