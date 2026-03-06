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

	// get Worker's address
	address := os.Getenv("HOSTNAME") + ":9000"

	// create and run Worker
	w := worker.NewWorker(address, "scheduler:9000")
	w.Run()

}
