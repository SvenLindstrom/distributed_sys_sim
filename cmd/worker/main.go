package main

import (
	"dssim/internal/misc"
	"dssim/internal/worker"
	"log"
	"os"
	"strings"
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
	port := os.Getenv("SCHEDULER_PORT")
	schedulerNames := strings.Split(os.Getenv("SCHEDULER_NAMES"), ",")
	schedulerAddresses := make([]string, len(schedulerNames))

	for i, name := range schedulerNames {
		schedulerAddresses[i] = name + ":" + port
	}

	// create and run Worker
	w := worker.NewWorker(address, schedulerAddresses)

	if err := w.Run(); err != nil {
		log.Fatal(err)
	}

}
