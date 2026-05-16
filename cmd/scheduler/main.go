package main

import (
	"dssim/internal/fault"
	"dssim/internal/misc"
	"dssim/internal/scheduler"
	"log"
	"os"
	"strconv"
)

func main() {

	f, err := misc.Loginit("scheduler")

	if err != nil {
		log.Fatal(err.Error())
	}
	defer f.Close()

	workerQueueSize, err := strconv.Atoi(os.Getenv("WORKER_Q_SIZE"))
	if err != nil {
		log.Println("worker q size not set")
		workerQueueSize = 1
	}

	println("logger ready")

	// scheduler.NewSchedular(workerQueueSize, jobQueueSize)
	// scheduler.NewSchedulerRaft(workerQueueSize, jobQueueSize)

	scheduler.InitScheduler(workerQueueSize)

	leader := os.Getenv("LEADER")
	if leader == "true" {
		// program fault injection
		duration := os.Getenv("SCHEDULER_CRASH")
		fault.InjectCrash(duration)
		println("fault injector started")

	}

	select {}
}
