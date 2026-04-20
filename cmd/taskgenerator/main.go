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

	intervalGen := generator.NewInterval()

	generator := generator.NewGenerator(duration, intervalGen, timeout)

	rpc.Register(&generator)
	rpc.HandleHTTP()
	println("generator ready")
	err = http.ListenAndServe(port, nil)

	if err != nil {
		log.Fatal(err.Error())
	}
}
