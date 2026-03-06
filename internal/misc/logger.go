package misc

import (
	"log/slog"
	"os"
)

func Loginit() (*os.File, error) {

	f, err := os.OpenFile("./logs/job.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	base := slog.New(
		slog.NewJSONHandler(f, nil),
	)

	logger := base.With("component", "scheduler")

	slog.SetDefault(logger)

	return f, nil
}
