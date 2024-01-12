package logger

import (
	"context"
	"fmt"
	"github.com/Stas9132/shortener/config"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

// LogrusLogger - Mediator
type LogrusLogger struct {
	*logrus.Logger
}

// Logger - interface to logger package
type Logger interface {
	Trace(...interface{})
	Tracef(string, ...interface{})
	Debug(...interface{})
	Debugf(string, ...interface{})
	Info(...interface{})
	Infof(string, ...interface{})
	Warn(...interface{})
	Warnf(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
	WithField(key string, value interface{}) *logrus.Entry
	WithFields(fields logrus.Fields) *logrus.Entry
}

// Log will log a message at the level given as parameter.
type Log func(l ...interface{})

// Logf will log a format message at the level given as parameter.
type Logf func(s string, l ...interface{})

var WithField,
	WithFields = logger.WithField,
	logger.WithFields

// Trace logs at the Trace level
var Trace,
	// Debug logs at the Debug level
	Debug,
	// Info logs at the Info level
	Info,
	// Warn logs at the Warn level
	Warn,
	// Error logs at the Error level
	Error,
	Print Log = logger.Trace,
	logger.Debug,
	logger.Info,
	logger.Warn,
	logger.Error,
	func(l ...interface{}) {
		fmt.Println(l...)
	}

// Tracef logs at the trace level with formatting
var Tracef,
	// Debugf logs at the debug level with formatting
	Debugf,
	// Infof logs at the info level with formatting
	Infof,
	// Warnf logs at the warn level with formatting
	Warnf,
	// Errorf logs at the error level with formatting
	Errorf,
	Printf Logf = logger.Tracef,
	logger.Debugf,
	logger.Infof,
	logger.Warnf,
	logger.Errorf,
	func(s string, l ...interface{}) {
		fmt.Printf(s, l...)
		fmt.Printf("\n")
	}

// NewLogrusLogger - Creates a new logger.
func NewLogrusLogger(ctx context.Context) (*LogrusLogger, error) {
	lvl, err := logrus.ParseLevel(*config.LogLevel)
	if err != nil {
		return nil, err
	}
	logger.SetLevel(lvl)
	return &LogrusLogger{logger}, nil
}

// Trace will log a message at the trace level.
func (l *LogrusLogger) Trace(args ...interface{}) {
	l.Logger.Trace(args...)
}

// Tracef will log a format message at the trace level.
func (l *LogrusLogger) Tracef(fmt string, args ...interface{}) {
	l.Logger.Tracef(fmt, args...)
}

// Debug will log a message at the debug level.
func (l *LogrusLogger) Debug(args ...interface{}) {
	l.Logger.Debug(args...)
}

// Debugf will log a format message at the debug level.
func (l *LogrusLogger) Debugf(fmt string, args ...interface{}) {
	l.Logger.Debugf(fmt, args...)
}

// Info will log a message at the info level.
func (l *LogrusLogger) Info(args ...interface{}) {
	l.Logger.Info(args...)
}

// Infof will log a format message at the info level.
func (l *LogrusLogger) Infof(fmt string, args ...interface{}) {
	l.Logger.Infof(fmt, args...)
}

// Warn will log a message at the warn level.
func (l *LogrusLogger) Warn(args ...interface{}) {
	l.Logger.Warn(args...)
}

// Warnf will log a format message at the warn level.
func (l *LogrusLogger) Warnf(fmt string, args ...interface{}) {
	l.Logger.Warnf(fmt, args...)
}

// Error will log a message at the error level.
func (l *LogrusLogger) Error(args ...interface{}) {
	l.Logger.Error(args...)
}

// Errorf will log a format message at the error level.
func (l *LogrusLogger) Errorf(fmt string, args ...interface{}) {
	l.Logger.Errorf(fmt, args...)
}

// WithField will add field to the log message
func (l *LogrusLogger) WithField(key string, value interface{}) *logrus.Entry {
	return l.Logger.WithField(key, value)
}

// WithFields will add fields to the log message
func (l *LogrusLogger) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.Logger.WithFields(fields)
}
