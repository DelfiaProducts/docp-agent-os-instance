package components

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/OryaHub/agent-os-instance/libs/interfaces"
	"github.com/OryaHub/agent-os-instance/libs/pkg"
	"github.com/OryaHub/agent-os-instance/libs/utils"
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

// InstallAgent execute install the agent orya
func (l *LinuxOperations) InstallAgent(version string) error {
	if err := l.program.Execute("bash", []string{fmt.Sprintf("VERSION=%s", version)}, "-c", utils.ChoiceInstallerOrUninstaller("linux", "agent", "install", version)); err != nil {
		return err
	}
	return nil
}

// InstallUpdater execute install the updater orya
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

	if err := l.program.Execute("bash", []string{fmt.Sprintf("VERSION=%s", version)}, "-c", utils.ChoiceInstallerOrUninstaller("linux", "updater", "install", version)); err != nil {
		return err
	}
	return nil
}

// UninstallAgent execute uninstall the agent orya
func (l *LinuxOperations) UninstallAgent(version string) error {
	if err := l.program.Execute("bash", []string{}, "-c", utils.ChoiceInstallerOrUninstaller("linux", "agent", "uninstall", version)); err != nil {
		return err
	}
	return nil
}

// UninstallUpdater execute uninstall the updater orya
func (l *LinuxOperations) UninstallUpdater(version string) error {
	if err := l.program.Execute("bash", []string{}, "-c", utils.ChoiceInstallerOrUninstaller("linux", "updater", "uninstall", version)); err != nil {
		return err
	}
	return nil
}

// UpdateAgent execute update the agent orya
func (l *LinuxOperations) UpdateAgent(version string) error {
	//install updater
	if err := l.InstallUpdater(version); err != nil {
		return err
	}

	return nil
}

// ExecuteUpdateVersion execute update the version
func (l *LinuxOperations) ExecuteUpdateVersion(version string) error {
	var managerUrl string
	var agentUrl string
	repoUrl := utils.GetBinariesRepositoryUrl()
	arch := utils.GetRuntimeArch()
	osSystem := utils.GetOSSystem()
	switch osSystem {
	case "linux":
		switch arch {
		case "amd64":
			managerUrl = fmt.Sprintf("%s/%s/manager-linux-amd64", repoUrl, version)
			agentUrl = fmt.Sprintf("%s/%s/agent-linux-amd64", repoUrl, version)
		case "arm64":
			managerUrl = fmt.Sprintf("%s/%s/manager-linux-arm64", repoUrl, version)
			agentUrl = fmt.Sprintf("%s/%s/agent-linux-arm64", repoUrl, version)
		}
	}

	statusManager, err := l.Status("manager")
	if err != nil {
		return err
	}

	statusAgent, err := l.Status("agent")
	if err != nil {
		return err
	}
	if statusAgent == "active" {
		if err := l.StopService("agent"); err != nil {
			return err
		}
	}
	if statusManager == "active" {
		if err := l.StopService("manager"); err != nil {
			return err
		}
	}

	respManager, _, err := utils.GetBinary(managerUrl)
	if err != nil {
		return err
	}

	respAgent, _, err := utils.GetBinary(agentUrl)
	if err != nil {
		return err
	}

	workdir, err := utils.GetWorkDirPath()
	if err != nil {
		return err
	}

	//validate path version
	pathVersion := filepath.Join(workdir, "bin", "releases", version)
	if err := l.fileSystem.VerifyDirExistAndCreate(pathVersion); err != nil {
		return err
	}

	pathCurrent := filepath.Join(workdir, "bin", "current")
	if err := l.fileSystem.VerifyDirExistAndCreate(pathVersion); err != nil {
		return err
	}

	pathManagerBinary := filepath.Join(pathVersion, "manager")
	pathAgentBinary := filepath.Join(pathVersion, "agent")

	pathCurrentManager := filepath.Join(pathCurrent, "manager")
	pathCurrentAgent := filepath.Join(pathCurrent, "agent")

	if err := l.fileSystem.WriteBinaryContent(pathManagerBinary, respManager); err != nil {
		return err
	}
	if err := l.fileSystem.WriteBinaryContent(pathAgentBinary, respAgent); err != nil {
		return err
	}

	err = os.Chmod(pathManagerBinary, 0755)
	if err != nil {
		return err
	}

	err = os.Chmod(pathAgentBinary, 0755)
	if err != nil {
		return err
	}
	//create symlink manager
	if err := l.fileSystem.CreateOrUpdateSymlink(pathManagerBinary, pathCurrentManager); err != nil {
		return err
	}

	//create symlink agent
	if err := l.fileSystem.CreateOrUpdateSymlink(pathAgentBinary, pathCurrentAgent); err != nil {
		return err
	}

	if err := l.RestartService("agent"); err != nil {
		return err
	}

	if err := l.RestartService("manager"); err != nil {
		return err
	}

	return nil
}

// ExecuteRollbackVersion execute rollback the version
func (l *LinuxOperations) ExecuteRollbackVersion(version string) error {
	workdir, err := utils.GetWorkDirPath()
	if err != nil {
		return err
	}
	//validate path version
	pathVersion := filepath.Join(workdir, "bin", "releases", version)
	if err := l.fileSystem.VerifyDirExistAndCreate(pathVersion); err != nil {
		return err
	}

	pathCurrent := filepath.Join(workdir, "bin", "current")
	if err := l.fileSystem.VerifyDirExistAndCreate(pathVersion); err != nil {
		return err
	}

	pathManagerBinary := filepath.Join(pathVersion, "manager")
	pathAgentBinary := filepath.Join(pathVersion, "agent")

	pathCurrentManager := filepath.Join(pathCurrent, "manager")
	pathCurrentAgent := filepath.Join(pathCurrent, "agent")

	if err := l.fileSystem.CreateOrUpdateSymlink(pathManagerBinary, pathCurrentManager); err != nil {
		return err
	}

	if err := l.fileSystem.CreateOrUpdateSymlink(pathAgentBinary, pathCurrentAgent); err != nil {
		return err
	}

	if err := l.RestartService("agent"); err != nil {
		return err
	}

	if err := l.RestartService("manager"); err != nil {
		return err
	}
	return nil
}

// UpdaterUninstall execute uninstall the updater agent
func (l *LinuxOperations) UpdaterUninstall(version string) error {
	job := fmt.Sprintf("* * * * * %s; crontab -l | grep -v '%s' | crontab -", utils.ChoiceInstallerOrUninstaller("linux", "updater", "uninstall", version), utils.ChoiceInstallerOrUninstaller("linux", "updater", "uninstall", version))
	command := fmt.Sprintf("(crontab -l 2>/dev/null; echo \"%s\") | crontab -", job)
	if err := l.program.Execute("bash", []string{}, "-c", command); err != nil {
		return err
	}
	return nil
}

// AutoUninstall execute auto uninstall the manager
func (l *LinuxOperations) AutoUninstall(version string) error {
	if err := l.program.Execute("bash", []string{}, "-c", utils.ChoiceInstallerOrUninstaller("linux", "agent", "uninstall", version)); err != nil {
		return err
	}
	if err := l.program.Execute("bash", []string{}, "-c", utils.ChoiceInstallerOrUninstaller("linux", "updater", "uninstall", version)); err != nil {
		return err
	}
	job := fmt.Sprintf("* * * * * %s; crontab -l | grep -v '%s' | crontab -", utils.ChoiceInstallerOrUninstaller("linux", "manager", "uninstall", version), utils.ChoiceInstallerOrUninstaller("linux", "manager", "uninstall", version))
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
