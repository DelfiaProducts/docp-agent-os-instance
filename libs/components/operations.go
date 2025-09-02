package components

import (
	"errors"
	"runtime"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
)

// SystemOperation return os operations for manager
func SystemOperation(logger interfaces.ILogger) (interfaces.IOSOperation, error) {
	switch runtime.GOOS {
	case "linux":
		return NewLinuxOperations(logger), nil
	case "darwin":
		return NewMacosOperations(logger), nil
	case "windows":
		return NewWindowsOperations(logger), nil
	}
	return nil, errors.New("os operation not configured")
}

// DatadogOperation return datadog operation
func DatadogOperation(logger interfaces.ILogger) (interfaces.IDatadogOperation, error) {
	switch runtime.GOOS {
	case "linux":
		return NewDatadogLinuxOperation(logger), nil
	case "darwin":
		return nil, nil
	case "windows":
		return NewDatadogWindowsOperation(logger), nil
	}
	return nil, errors.New("datadog operation not configured")
}
