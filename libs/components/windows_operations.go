//go:build windows

package components

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/pkg"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	URL_AGENT_WINDOWS = "https://test-docp-agent-data.s3.amazonaws.com/installer/install_agent_windows.msi"
)

// WindowsOperations is instance of windows operations
type WindowsOperations struct {
	serviceName            string
	logger                 interfaces.ILogger
	filesystem             *pkg.FileSystem
	program                *pkg.ExecProgram
	ymlClient              *pkg.YmlClient
	scmManager             *SCMManager
	isProcessAutoUninstall bool
}

// NewWindowsOperations return instance of windows operations
func NewWindowsOperations(logger interfaces.ILogger) *WindowsOperations {
	var srvName string
	if os.Getenv("SCM") == "agent" {
		srvName = "DocpAgent"
	} else {
		srvName = "DocpManager"
	}
	return &WindowsOperations{
		logger:                 logger,
		program:                pkg.NewExecProgram(),
		filesystem:             pkg.NewFileSystem(),
		ymlClient:              pkg.NewYmlClient(),
		serviceName:            srvName,
		isProcessAutoUninstall: false,
	}
}

func (l *WindowsOperations) Setup() error {
	handlerSCM := NewHandlerSCM()
	scmManager := NewSCMManager(l.serviceName, handlerSCM)
	l.scmManager = scmManager
	return nil
}

func (l *WindowsOperations) Status(serviceName string) (string, error) {
	name := utils.ChoiceNameServiceWindows(serviceName)
	m, err := mgr.Connect()
	if err != nil {
		return "", err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			return "inactive", nil
		}
		return "", err
	}
	defer s.Close()
	status, err := s.Query()
	if err != nil {
		return "", err
	}
	if status.State == svc.Running {
		return "active", nil
	} else {
		return "inactive", nil
	}
}

// AlreadyInstalled verify if service installed
func (l *WindowsOperations) AlreadyInstalled(serviceName string) (bool, error) {
	name := utils.ChoiceNameServiceWindows(serviceName)
	m, err := mgr.Connect()
	if err != nil {
		return false, err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			return false, nil
		}
		return false, err
	}
	defer s.Close()
	return true, nil
}

// DaemonReload execute daemon reload the service in systemd
func (l *WindowsOperations) DaemonReload() error {
	return nil
}

// RestartService execute restart the service in systemd
func (l *WindowsOperations) RestartService(serviceName string) error {
	name := utils.ChoiceNameServiceWindows(serviceName)
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return err
	}
	defer s.Close()
	_, err = s.Control(svc.Stop)
	if err != nil {
		return err
	}
	if err := s.Start(); err != nil {
		return err
	}

	return nil
}

// StopService execute stop the service in systemd
func (l *WindowsOperations) StopService(serviceName string) error {
	name := utils.ChoiceNameServiceWindows(serviceName)
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return err
	}
	defer s.Close()
	_, err = s.Control(svc.Stop)
	if err != nil {
		return err
	}
	return nil
}

// InstallAgent execute install the agent docp
func (l *WindowsOperations) InstallAgent(version string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService("DocpAgent")
	if err != nil {
		if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			command := fmt.Sprintf(`Start-Process -Wait msiexec -ArgumentList '/qn /i %s'`, URL_AGENT_WINDOWS)
			out, err := l.program.ExecuteWithOutput("powershell", []string{}, "-Command", command)
			if err != nil {
				l.logger.Error("error in install docp agent start process", "error", err)
				return err
			}
			l.logger.Debug("install docp agent", "output", out)
			return nil
		}
		return err
	}
	defer s.Close()
	return nil
}

// InstallUpdater execute install the updater docp
func (l *WindowsOperations) InstallUpdater(version string) error {
	//TODO: add logic for install updater windows
	return nil
}

// UpdateAgent execute update the agent docp
func (l *WindowsOperations) UpdateAgent(version string) error {
	//TODO: add logic for update agent windows
	return nil
}

// UninstallAgent execute uninstall the agent docp
func (l *WindowsOperations) UninstallAgent() error {
	wmiCmd := `(Get-Package -Name "DocpAgent").Metadata['ProductCode']`
	out, err := l.program.ExecuteWithOutput("powershell", []string{}, "-Command", wmiCmd)
	if err != nil {
		l.logger.Error("error in get wmi docp agent", "error", err)
		return err
	}
	l.logger.Debug("output wmi docp agent", "output", out)
	re := regexp.MustCompile(`\{[A-Fa-f0-9\-]+\}`)
	identifyingNumber := re.FindString(string(out))

	if identifyingNumber == "" {
		return fmt.Errorf("not found identify number for docp agent")
	}

	outUnistall, err := l.program.ExecuteWithOutput("powershell", []string{}, "-Command", fmt.Sprintf(`start-process msiexec -Wait -ArgumentList ('/log', 'C:\uninst.log', '/norestart', '/q', '/x', '%s')`, identifyingNumber))
	if err != nil {
		l.logger.Error("error in uninstall docp agent", "error", err)
		return err
	}
	l.logger.Debug("output uninstall docp agent", "output", outUnistall)
	return nil
}

// schedulerStopDocpManager execute stop the docp manager
func (l *WindowsOperations) schedulerStopDocpManager() error {
	taskNameStop := "DocpManagerStop"
	runAtStop := time.Now().Add(time.Duration(1) * time.Minute)
	runStopTime := runAtStop.Format("15:04")

	cmdStop := fmt.Sprintf(`"sc.exe stop DocpManager"`)

	outputStop, err := l.program.ExecuteWithOutput("powershell", []string{}, "-Command", fmt.Sprintf(`start-process schtasks -Wait -ArgumentList ('/Create', '/SC', 'ONCE', '/TN','%s', '/TR', '%s','/ST','%s','/RU','SYSTEM','/RL','HIGHEST','/F')`, taskNameStop, cmdStop, runStopTime))
	if err != nil {
		l.logger.Error("auto uninstall manager create scheduler stop error", "error", err.Error())
		return err
	}
	l.logger.Debug("auto uninstall manager stop", "outputStop", outputStop)
	return nil
}

// schedulerAutoUninstall execute scheduler for auto remove manager
func (l *WindowsOperations) schedulerAutoUninstall() error {
	wmiCmd := `(Get-Package -Name "DocpManager").Metadata['ProductCode']`
	out, err := l.program.ExecuteWithOutput("powershell", []string{}, "-Command", wmiCmd)
	if err != nil {
		l.logger.Error("error in get wmi docp agent", "error", err)
		return err
	}
	l.logger.Debug("output wmi docp manager", "output", out)
	re := regexp.MustCompile(`\{[A-Fa-f0-9\-]+\}`)
	identifyingNumber := re.FindString(string(out))

	if identifyingNumber == "" {
		return fmt.Errorf("not found identify number for docp agent")
	}

	taskName := "DocpManagerAutoRemove"
	runAt := time.Now().Add(time.Duration(3) * time.Minute)
	runTime := runAt.Format("15:04")

	cmd := fmt.Sprintf(`"msiexec /x %s /quiet /norestart"`, identifyingNumber)

	outputAutoRemove, err := l.program.ExecuteWithOutput("powershell", []string{}, "-Command", fmt.Sprintf(`start-process schtasks -Wait -ArgumentList ('/Create', '/SC', 'ONCE', '/TN','%s', '/TR', '%s','/ST','%s','/RU','SYSTEM','/RL','HIGHEST','/F')`, taskName, cmd, runTime))
	if err != nil {
		l.logger.Error("auto uninstall manager create scheduler uninstall error", "error", err.Error())
		return err
	}
	l.logger.Debug("auto uninstall manager", "outputAutoRemove", outputAutoRemove)
	return nil
}

// AutoUninstall execute auto uninstall the manager
func (l *WindowsOperations) AutoUninstall() error {
	if !l.isProcessAutoUninstall {
		if err := l.schedulerAutoUninstall(); err != nil {
			return err
		}
		if err := l.schedulerStopDocpManager(); err != nil {
			return err
		}
	}
	l.isProcessAutoUninstall = true
	return nil
}

// Execute run handler functions the operation
func (l *WindowsOperations) Execute() error {
	if err := l.scmManager.Run(); err != nil {
		return err
	}
	return nil
}
