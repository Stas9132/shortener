package logger_test

import (
	"context"
	"github.com/Stas9132/shortener/internal/logger"
)

func Example() {
	l, err := logger.NewLogrusLogger(context.Background())
	if err != nil {
		panic(err)
	}

	l.Trace("Trace message")
	l.Info("Info message")
	l.Warn("Warn message")
	l.Error("Error message")
}
