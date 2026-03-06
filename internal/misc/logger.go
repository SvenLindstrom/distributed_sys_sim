package misc

import (
	"fmt"
	"log/slog"
	"os"
)

func Loginit(component string) (*os.File, error) {

	fileName := fmt.Sprintf("./logs/%s.log", component)

	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	base := slog.New(
		slog.NewJSONHandler(f, nil),
	)

	logger := base.With("component", component)

	slog.SetDefault(logger)

	return f, nil
}
