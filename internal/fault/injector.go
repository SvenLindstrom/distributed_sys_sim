package fault

import (
	"log/slog"
	"os"
	"time"
)

func InjectCrash(durationStr string) {
	duration, err := time.ParseDuration(durationStr)
	if err != nil || duration <= 0 {
		slog.Info(
			"Crash injection disabled because of invalid duration value",
			"duration", duration,
		)
		return
	}

	go crash(duration)
}

func crash(duration time.Duration) {
	slog.Info(
		"Fault injection scheduler for Scheduler process crash",
		"after", duration,
	)

	// timer starts
	time.Sleep(duration)

	slog.Info(
		"Scheduler crashing now",
	)

	// crash
	os.Exit(1)
}
