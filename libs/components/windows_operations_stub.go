//go:build !windows

package components

import "github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"

// WindowsOperations is instance of windows operations
type WindowsOperations struct{}

func NewWindowsOperations(logger interfaces.ILogger) *WindowsOperations {
	return nil
}

func (l *WindowsOperations) Setup() error {
	return nil
}

// isAdmin verify if admin
func (l *WindowsOperations) isAdmin() bool {
	return false
}

// runAsAdmin reexecute with admin
func (l *WindowsOperations) runAsAdmin(args []string) {
}

// getVersionAgent return version the agent
func (l *WindowsOperations) getVersionAgent() (string, error) {
	return "", nil
}

// downloadFile get binary file from bucket
func (l *WindowsOperations) downloadFile(url, dest string) error {
	return nil
}

// getEnvsForAgent get environments for running agent
func (l *WindowsOperations) getEnvsForAgent() ([]string, error) {
	return []string{}, nil
}

// scheduleServiceStopAndRemoval agenda a remoção do serviço no Task Scheduler
func (l *WindowsOperations) scheduleServiceStopAndRemoval(serviceName string, delaySeconds int) error {
	return nil
}

func (l *WindowsOperations) Status(serviceName string) (string, error) {
	return "", nil
}

// AlreadyInstalled verify if service installed
func (l *WindowsOperations) AlreadyInstalled(serviceName string) (bool, error) {
	return false, nil
}

// DaemonReload execute daemon reload the service in systemd
func (l *WindowsOperations) DaemonReload() error {
	return nil
}

// RestartService execute restart the service in systemd
func (l *WindowsOperations) RestartService(serviceName string) error {
	return nil
}

// StopService execute stop the service in systemd
func (l *WindowsOperations) StopService(serviceName string) error {
	return nil
}

// InstallAgent execute install the agent docp
func (l *WindowsOperations) InstallAgent(version string) error {
	return nil
}

// InstallUpdater execute install the updater docp
func (l *WindowsOperations) InstallUpdater(version string) error {
	return nil
}

// UpdateAgent execute update the agent docp
func (l *WindowsOperations) UpdateAgent(version string) error {
	return nil
}

// UninstallAgent execute uninstall the agent docp
func (l *WindowsOperations) UninstallAgent() error {
	return nil
}

// AutoUninstall execute auto uninstall the manager
func (l *WindowsOperations) AutoUninstall() error {
	return nil
}

// Execute run handler functions the operation
func (l *WindowsOperations) Execute() error {
	return nil
}
