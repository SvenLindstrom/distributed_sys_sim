package generator

import (
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type Interval interface {
	GetInterval() time.Duration
}

func NewInterval() Interval {
	rate := os.Getenv("TASK_RATE")
	interval := os.Getenv("INTERVAL")
	if val, err := strconv.ParseFloat(rate, 64); err == nil {
		println("Task With Poisson Rate ", val)
		return &PoissonInterval{EventsPerSec: val}
	} else if val, err := strconv.Atoi(interval); err == nil {
		println("Task With Static Interval: ", val)
		return &StaticInterval{Interval: val}
	}
	println("No delay specifications, using fall back value of static 1sec")
	return &StaticInterval{Interval: 1000}
}

type StaticInterval struct {
	Interval int
}

func (si *StaticInterval) GetInterval() time.Duration {
	return (time.Duration(si.Interval) * time.Millisecond)
}

type PoissonInterval struct {
	EventsPerSec float64
}

func (pi *PoissonInterval) GetInterval() time.Duration {
	u := rand.Float64()
	delay := -math.Log(1-u) / pi.EventsPerSec
	return time.Duration(delay * float64(time.Second))
}
