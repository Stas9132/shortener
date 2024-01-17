package logger

import "github.com/sirupsen/logrus"

// Dummy - dummy logger
type Dummy struct {
}

// NewDummy() - Creates a dummy new logger.
func NewDummy() *Dummy {
	return &Dummy{}
}

// Trace - dummy method
func (d *Dummy) Trace(...interface{}) {}

// Tracef - dummy method
func (d *Dummy) Tracef(string, ...interface{}) {}

// Debug - dummy method
func (d *Dummy) Debug(...interface{}) {}

// Debugf - dummy method
func (d *Dummy) Debugf(string, ...interface{}) {}

// Info - dummy method
func (d *Dummy) Info(...interface{}) {}

// Infof - dummy method
func (d *Dummy) Infof(string, ...interface{}) {}

// Warn - dummy method
func (d *Dummy) Warn(...interface{}) {}

// Warnf - dummy method
func (d *Dummy) Warnf(string, ...interface{}) {}

// Error - dummy method
func (d *Dummy) Error(...interface{}) {}

// Errorf - dummy method
func (d *Dummy) Errorf(string, ...interface{}) {}

// WithField - dummy method
func (d *Dummy) WithField(key string, value interface{}) *logrus.Entry {
	return logrus.NewEntry(logger)
}

// WithFields - dummy method
func (d *Dummy) WithFields(fields logrus.Fields) *logrus.Entry { return logrus.NewEntry(logger) }
