package main

import (
	"dssim/internal/misc"
	"dssim/internal/worker"
	"log"
	"os"
)

func main() {

	// initialise Logger
	file, err := misc.Loginit("worker")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// get Worker and Scheduler addresses
	address := os.Getenv("HOSTNAME") + ":" + os.Getenv("WORKER_PORT")
	schedulerAddress := "scheduler" + ":" + os.Getenv("SCHEDULER_PORT")

	// create and run Worker
	w := worker.NewWorker(address, schedulerAddress)
	w.Run()

}
