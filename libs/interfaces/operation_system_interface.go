package interfaces

// IOperationSystem is interface for os
type IOperationSystem interface {
	Prepare() error
}

// IOSOperation is interface for os operations
type IOSOperation interface {
	Setup() error
	Execute() error
	Status(serviceName string) (string, error)
	AlreadyInstalled(serviceName string) (bool, error)
	RestartService(serviceName string) error
	StopService(serviceName string) error
	InstallAgent(version string) error
	InstallUpdater(version string) error
	UpdateAgent(version string) error
	UninstallAgent(version string) error
	AutoUninstall(version string) error
	DaemonReload() error
}
