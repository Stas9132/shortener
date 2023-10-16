package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"shortener/config"
)

var logger = logrus.New()

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
}
type Log func(l ...interface{})
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

func Init() {
	lvl, err := logrus.ParseLevel(*config.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	logger.SetLevel(lvl)
}
