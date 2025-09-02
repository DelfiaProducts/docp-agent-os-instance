package components

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/pkg"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

const (
	CURL_AGENT_LINUX_INSTALL_SH     = "curl -L https://test-docp-agent-data.s3.amazonaws.com/installer/install_agent_linux.sh | bash"
	CURL_AGENT_LINUX_UNINSTALL_SH   = "curl -L https://test-docp-agent-data.s3.amazonaws.com/installer/uninstall_agent_linux.sh | bash"
	CURL_UPDATER_LINUX_INSTALL_SH   = "curl -L https://test-docp-agent-data.s3.amazonaws.com/installer/install_updater_linux.sh | bash"
	CURL_UPDATER_LINUX_UNINSTALL_SH = "curl -L https://test-docp-agent-data.s3.amazonaws.com/installer/uninstall_updater_linux.sh | bash"
	CURL_LINUX_AUTO_UNINSTALL_SH    = "curl -L https://test-docp-agent-data.s3.amazonaws.com/installer/uninstall_manager_linux.sh | bash"
)

// LinuxOperations is instance of linux operations
type LinuxOperations struct {
	logger     interfaces.ILogger
	systemd    *pkg.SystemdClient
	fileSystem *pkg.FileSystem
	program    *pkg.ExecProgram
}

// NewLinuxOperations return instance of linux operations
func NewLinuxOperations(looger interfaces.ILogger) *LinuxOperations {
	return &LinuxOperations{
		logger:  looger,
		program: pkg.NewExecProgram(),
	}
}

func (l *LinuxOperations) Setup() error {
	systemd := pkg.NewSystemdClient()
	l.systemd = systemd
	fileSystem := pkg.NewFileSystem()
	l.fileSystem = fileSystem
	return nil
}

func (l *LinuxOperations) Status(serviceName string) (string, error) {
	name := utils.ChoiceNameService(serviceName)
	output, err := l.systemd.Status(name)
	if err != nil {
		return "", err
	}
	output = strings.ReplaceAll(output, "\"", "")
	return output, nil
}

// AlreadyInstalled verify if service installed
func (l *LinuxOperations) AlreadyInstalled(serviceName string) (bool, error) {
	name := utils.ChoiceNameService(serviceName)
	installed, err := l.systemd.AlreadyInstalledService(name)
	if err != nil {
		return false, err
	}
	return installed, nil
}

// DaemonReload execute daemon reload the service in systemd
func (l *LinuxOperations) DaemonReload() error {
	if err := l.program.Execute("sudo", []string{}, "systemctl", "daemon-reload"); err != nil {
		return err
	}
	return nil
}

// RestartService execute restart the service in systemd
func (l *LinuxOperations) RestartService(serviceName string) error {
	name := utils.ChoiceNameService(serviceName)
	if err := l.program.Execute("sudo", []string{}, "systemctl", "restart", name); err != nil {
		return err
	}
	return nil
}

// StopService execute stop the service in systemd
func (l *LinuxOperations) StopService(serviceName string) error {
	name := utils.ChoiceNameService(serviceName)
	if err := l.program.Execute("sudo", []string{}, "systemctl", "stop", name); err != nil {
		return err
	}
	return nil
}

// InstallAgent execute install the agent docp
func (l *LinuxOperations) InstallAgent(version string) error {
	if err := l.program.Execute("bash", []string{fmt.Sprintf("VERSION=%s", version)}, "-c", CURL_AGENT_LINUX_INSTALL_SH); err != nil {
		return err
	}
	return nil
}

// InstallUpdater execute install the updater docp
func (l *LinuxOperations) InstallUpdater(version string) error {
	workdir, err := utils.GetWorkDirPath()
	if err != nil {
		return err
	}

	//validate path version
	pathVersion := filepath.Join(workdir, "bin", "releases", version)
	if err := l.fileSystem.VerifyDirExistAndCreate(pathVersion); err != nil {
		return err
	}

	if err := l.program.Execute("bash", []string{fmt.Sprintf("VERSION=%s", version)}, "-c", CURL_UPDATER_LINUX_INSTALL_SH); err != nil {
		return err
	}
	return nil
}

// UninstallAgent execute uninstall the agent docp
func (l *LinuxOperations) UninstallAgent() error {
	if err := l.program.Execute("bash", []string{}, "-c", CURL_AGENT_LINUX_UNINSTALL_SH); err != nil {
		return err
	}
	return nil
}

// UninstallUpdater execute uninstall the updater docp
func (l *LinuxOperations) UninstallUpdater() error {
	if err := l.program.Execute("bash", []string{}, "-c", CURL_UPDATER_LINUX_UNINSTALL_SH); err != nil {
		return err
	}
	return nil
}

// UpdateAgent execute update the agent docp
func (l *LinuxOperations) UpdateAgent(version string) error {
	//install updater
	if err := l.InstallUpdater(version); err != nil {
		return err
	}

	return nil
}

// AutoUninstall execute auto uninstall the manager
func (l *LinuxOperations) AutoUninstall() error {
	if err := l.program.Execute("bash", []string{}, "-c", CURL_AGENT_LINUX_UNINSTALL_SH); err != nil {
		return err
	}
	if err := l.program.Execute("bash", []string{}, "-c", CURL_UPDATER_LINUX_UNINSTALL_SH); err != nil {
		return err
	}
	job := fmt.Sprintf("* * * * * %s; crontab -l | grep -v '%s' | crontab -", CURL_LINUX_AUTO_UNINSTALL_SH, CURL_LINUX_AUTO_UNINSTALL_SH)
	command := fmt.Sprintf("(crontab -l 2>/dev/null; echo \"%s\") | crontab -", job)
	if err := l.program.Execute("bash", []string{}, "-c", command); err != nil {
		return err
	}
	return nil
}

// Execute run handlers the operation
func (l *LinuxOperations) Execute() error {
	return nil
}
