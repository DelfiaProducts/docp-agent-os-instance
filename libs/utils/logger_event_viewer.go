//go:build windows

package utils

import (
	"fmt"

	"golang.org/x/sys/windows/svc/eventlog"
)

// OryaLoggerEventViewer is struct for logger the orya for windows
type OryaLoggerEventViewer struct {
	serviceName string
}

// NewOryaLoggerEventViewer return instance of orya logger event viewer
func NewOryaLoggerEventViewer(serviceName string) *OryaLoggerEventViewer {
	return &OryaLoggerEventViewer{
		serviceName: serviceName,
	}
}

// RegisterEventViewer execute register event viewer
func (d *OryaLoggerEventViewer) RegisterEventViewer() error {
	err := eventlog.InstallAsEventCreate(d.serviceName, eventlog.Info|eventlog.Warning|eventlog.Error)
	if err != nil {
		return err
	}
	return nil
}

// Debug execute logging the debug
func (d *OryaLoggerEventViewer) Debug(msg string, args ...any) {
	elog, _ := eventlog.Open(d.serviceName)
	defer elog.Close()
	var formattedArgs string
	if len(args) > 0 {
		formattedArgs = fmt.Sprintf(msg, args...)
	} else {
		formattedArgs = msg
	}
	elog.Info(1, formattedArgs)
}

// Info execute logging the info
func (d *OryaLoggerEventViewer) Info(msg string, args ...any) {
	elog, _ := eventlog.Open(d.serviceName)
	defer elog.Close()
	var formattedArgs string
	if len(args) > 0 {
		formattedArgs = fmt.Sprintf(msg, args...)
	} else {
		formattedArgs = msg
	}
	elog.Info(1, formattedArgs)
}

// Warn execute logging the warning
func (d *OryaLoggerEventViewer) Warn(msg string, args ...any) {
	elog, _ := eventlog.Open(d.serviceName)
	defer elog.Close()
	var formattedArgs string
	if len(args) > 0 {
		formattedArgs = fmt.Sprintf(msg, args...)
	} else {
		formattedArgs = msg
	}
	elog.Warning(2, formattedArgs)
}

// Error execut logging the error
func (d *OryaLoggerEventViewer) Error(msg string, args ...any) {
	elog, _ := eventlog.Open(d.serviceName)
	defer elog.Close()
	var formattedArgs string
	if len(args) > 0 {
		formattedArgs = fmt.Sprintf(msg, args...)
	} else {
		formattedArgs = msg
	}
	elog.Error(3, formattedArgs)
}

// Close close viewer event
func (d *OryaLoggerEventViewer) Close() error {
	return nil
}
