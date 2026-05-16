package misc

import (
	"log/slog"
	"os"
	"path/filepath"
)

func Loginit(component string) (*os.File, error) {
	path := os.Getenv("LOGPATH")

	fileName := component + ".log"
	filePath := filepath.Join(path, fileName)

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
