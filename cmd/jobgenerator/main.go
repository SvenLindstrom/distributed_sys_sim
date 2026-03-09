package main

import (
	"dssim/internal/generator"
	"dssim/internal/misc"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
)

func main() {

	interval, err := strconv.Atoi(os.Getenv("INTERVAL"))
	if err != nil {
		println("interval not set")
		interval = 1000
	}
	duration, err := strconv.Atoi(os.Getenv("DURATION"))
	if err != nil {
		println("duration not set")
		duration = 1000
	}

	timeout, err := strconv.Atoi(os.Getenv("TIMEOUT"))
	if err != nil {
		println("timeout not set")
		timeout = 5
	}

	port := ":" + os.Getenv("GENERATOR_PORT")

	f, err := misc.Loginit("generator")

	if err != nil {
		log.Fatal(err.Error())
	}
	defer f.Close()

	generator := generator.NewGnerator(interval, duration, timeout)

	rpc.Register(&generator)
	rpc.HandleHTTP()

	err = http.ListenAndServe(port, nil)

	if err != nil {
		log.Fatal(err.Error())
	}
}
