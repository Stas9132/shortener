package logger

import "github.com/sirupsen/logrus"

type Dummy struct {
}

func NewDummy() *Dummy {
	return &Dummy{}
}

func (d *Dummy) Trace(...interface{})          {}
func (d *Dummy) Tracef(string, ...interface{}) {}
func (d *Dummy) Debug(...interface{})          {}
func (d *Dummy) Debugf(string, ...interface{}) {}
func (d *Dummy) Info(...interface{})           {}
func (d *Dummy) Infof(string, ...interface{})  {}
func (d *Dummy) Warn(...interface{})           {}
func (d *Dummy) Warnf(string, ...interface{})  {}
func (d *Dummy) Error(...interface{})          {}
func (d *Dummy) Errorf(string, ...interface{}) {}
func (d *Dummy) WithField(key string, value interface{}) *logrus.Entry {
	return logrus.NewEntry(logger)
}
func (d *Dummy) WithFields(fields logrus.Fields) *logrus.Entry { return logrus.NewEntry(logger) }
