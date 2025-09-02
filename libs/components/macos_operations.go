package components

import (
	"fmt"
	"strings"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/pkg"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

const (
	CURL_AGENT_MACOS_INSTALL_SH   = "curl -L https://test-docp-agent-data.s3.amazonaws.com/installer/install_agent_macos.sh | bash"
	CURL_AGENT_MACOS_UNINSTALL_SH = "curl -L https://test-docp-agent-data.s3.amazonaws.com/installer/uninstall_agent_macos.sh | bash"
	CURL_MACOS_AUTO_UNINSTALL_SH  = "curl -L https://test-docp-agent-data.s3.amazonaws.com/installer/uninstall_manager_macos.sh | bash"
)

// MacosOperations is instance of macos operations
type MacosOperations struct {
	// systemd *internal.SystemdClient
	logger    interfaces.ILogger
	hostStats *pkg.HostStats
	program   *pkg.ExecProgram
}

// NewMacosOperations return instance of macos operations
func NewMacosOperations(logger interfaces.ILogger) *MacosOperations {
	return &MacosOperations{
		logger:    logger,
		program:   pkg.NewExecProgram(),
		hostStats: pkg.NewHostStats(),
	}
}

func (l *MacosOperations) Setup() error {
	return nil
}

func (l *MacosOperations) Status(serviceName string) (string, error) {
	name := utils.GetNameForProcess(serviceName)
	processes, err := l.hostStats.ProcessInfo()
	if err != nil {
		return "", err
	}
	activeProcess := false
	for _, prc := range processes {
		if strings.Contains(prc.Name, name) {
			activeProcess = true
		}
	}
	if activeProcess {
		return "active", nil
	} else {
		return "inactive", nil
	}
}

func (l *MacosOperations) AlreadyInstalled(serviceName string) (bool, error) {
	// TODO: not implemented
	return false, nil
}

// DaemonReload execute daemon reload the service in systemd
func (l *MacosOperations) DaemonReload() error {
	return nil
}

// RestartService execute restart the service in systemd
func (l *MacosOperations) RestartService(serviceName string) error {
	name := utils.GetNameForLaunchd(serviceName)
	id, err := l.program.ExecuteWithOutput("bash", []string{}, "id", "-u")
	if err != nil {
		return err
	}
	command := fmt.Sprintf("gui/%s/%s", id, name)
	if err := l.program.Execute("bash", []string{}, "launchctl", "kickstart", "-k", command); err != nil {
		return err
	}
	return nil
}

// StopService execute stop the service in systemd
func (l *MacosOperations) StopService(serviceName string) error {
	name := utils.GetNameForLaunchd(serviceName)
	id, err := l.program.ExecuteWithOutput("bash", []string{}, "id", "-u")
	if err != nil {
		return err
	}
	command := fmt.Sprintf("gui/%s/%s", id, name)
	if err := l.program.Execute("bash", []string{}, "launchctl", "bootout", "-k", command); err != nil {
		return err
	}
	return nil
}

// InstallAgent execute install the agent docp
func (l *MacosOperations) InstallAgent(version string) error {
	if err := l.program.Execute("bash", []string{fmt.Sprintf("VERSION=%s", version)}, "-c", CURL_AGENT_MACOS_INSTALL_SH); err != nil {
		return err
	}
	return nil
}

// InstallUpdater execute install the updater docp
func (l *MacosOperations) InstallUpdater(version string) error {
	return nil
}

// UpdateAgent execute update the agent docp
func (l *MacosOperations) UpdateAgent(version string) error {
	//TODO: add logic for update agent macos
	return nil
}

// UninstallAgent execute uninstall the agent docp
func (l *MacosOperations) UninstallAgent() error {
	if err := l.program.Execute("bash", []string{}, "-c", CURL_AGENT_MACOS_UNINSTALL_SH); err != nil {
		return err
	}
	return nil
}

// AutoUninstall execute auto uninstall the manager
func (l *MacosOperations) AutoUninstall() error {
	if err := l.program.Execute("bash", []string{}, "-c", CURL_AGENT_MACOS_UNINSTALL_SH); err != nil {
		return err
	}
	job := fmt.Sprintf("* * * * * %s; crontab -l | grep -v '%s' | crontab -", CURL_MACOS_AUTO_UNINSTALL_SH, CURL_MACOS_AUTO_UNINSTALL_SH)
	command := fmt.Sprintf("(crontab -l 2>/dev/null; echo \"%s\") | crontab -", job)
	if err := l.program.Execute("bash", []string{}, "-c", command); err != nil {
		return err
	}
	return nil
}

// Execute run handlers the operation
func (l *MacosOperations) Execute() error {
	return nil
}
