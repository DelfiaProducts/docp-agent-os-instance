//go:build windows

package utils

import (
	"fmt"

	"golang.org/x/sys/windows/svc/eventlog"
)

// DocpLoggerEventViewer is struct for logger the docp for windows
type DocpLoggerEventViewer struct {
	serviceName string
}

// NewDocpLoggerEventViewer return instance of docp logger event viewer
func NewDocpLoggerEventViewer(serviceName string) *DocpLoggerEventViewer {
	return &DocpLoggerEventViewer{
		serviceName: serviceName,
	}
}

// RegisterEventViewer execute register event viewer
func (d *DocpLoggerEventViewer) RegisterEventViewer() error {
	err := eventlog.InstallAsEventCreate(d.serviceName, eventlog.Info|eventlog.Warning|eventlog.Error)
	if err != nil {
		return err
	}
	return nil
}

// Debug execute logging the debug
func (d *DocpLoggerEventViewer) Debug(msg string, args ...any) {
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
func (d *DocpLoggerEventViewer) Info(msg string, args ...any) {
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
func (d *DocpLoggerEventViewer) Warn(msg string, args ...any) {
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
func (d *DocpLoggerEventViewer) Error(msg string, args ...any) {
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
func (d *DocpLoggerEventViewer) Close() error {
	return nil
}
