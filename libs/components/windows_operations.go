//go:build windows

package components

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"github.com/OryaHub/agent-os-instance/libs/interfaces"
	"github.com/OryaHub/agent-os-instance/libs/pkg"
	"github.com/OryaHub/agent-os-instance/libs/utils"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

// WindowsOperations is instance of windows operations
type WindowsOperations struct {
	serviceName string
	logger      interfaces.ILogger
	filesystem  *pkg.FileSystem
	program     *pkg.ExecProgram
	ymlClient   *pkg.YmlClient
	scmManager  *SCMManager
}

// NewWindowsOperations return instance of windows operations
func NewWindowsOperations(logger interfaces.ILogger) *WindowsOperations {
	var srvName string
	if os.Getenv("SCM") == "agent" {
		srvName = "OryaAgent"
	} else {
		srvName = "OryaManager"
	}
	return &WindowsOperations{
		logger:      logger,
		program:     pkg.NewExecProgram(),
		filesystem:  pkg.NewFileSystem(),
		ymlClient:   pkg.NewYmlClient(),
		serviceName: srvName,
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
	dependsSrv, err := s.ListDependentServices(svc.Active)
	if err != nil {
		return err
	}
	if len(dependsSrv) > 0 {
		for _, dep := range dependsSrv {
			depService, err := m.OpenService(dep)
			if err != nil {
				return err
			}
			defer depService.Close()
			_, errDep := depService.Control(svc.Stop)
			if errDep != nil {
				return errDep
			}
			for {
				time.Sleep(time.Second * 1)
				statusDep, errDep := depService.Query()
				if errDep != nil {
					return errDep
				}

				if statusDep.State == svc.Stopped {
					break
				}
			}
		}
	}
	status, err := s.Query()
	if err != nil {
		return err
	}
	if status.State == svc.Running {
		_, err := s.Control(svc.Stop)
		if err != nil {
			return err
		}
		for {
			time.Sleep(time.Second * 1)
			status, err = s.Query()
			if err != nil {
				return err
			}

			if status.State == svc.Stopped {
				break
			}
		}
		l.terminateServiceProcesses(name)
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
	status, err := s.Query()
	if err != nil {
		return err
	}
	if status.State == svc.Running {
		_, err = s.Control(svc.Stop)
		if err != nil {
			return err
		}
		for {
			time.Sleep(time.Second * 1)
			status, err = s.Query()
			if err != nil {
				return err
			}

			if status.State == svc.Stopped {
				break
			}
		}
		l.terminateServiceProcesses(name)
	}
	return nil
}

// terminateServiceProcesses attempts to kill any remaining processes belonging
// to the given SCM service name. This is a best-effort cleanup for processes
// (especially child processes) that may survive a service stop. Errors are
// logged but not returned, so the caller's service start is never blocked.
func (l *WindowsOperations) terminateServiceProcesses(scmName string) {
	out, err := l.program.ExecuteWithOutput("taskkill", []string{}, "/f", "/fi", fmt.Sprintf("SERVICES eq %s", scmName), "/t")
	if err != nil {
		// taskkill exits with code 128 when no process matches the filter,
		// which is expected when the service stopped cleanly. Log and move on.
		l.logger.Debug("taskkill for service processes (may be already stopped)", "service", scmName, "output", out)
	} else {
		l.logger.Debug("taskkill for service processes completed", "service", scmName, "output", out)
	}
}

// InstallAgent execute install the agent orya
func (l *WindowsOperations) InstallAgent(version string) error {
	name := utils.ChoiceNameServiceWindows("agent")
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			command := utils.ChoiceInstallerOrUninstaller("windows", "agent", "install", version)
			out, err := l.program.ExecuteWithOutput("powershell", []string{}, "-Command", command)
			if err != nil {
				l.logger.Error("error in install orya agent start process", "error", err)
				return err
			}
			l.logger.Debug("install orya agent", "output", out)
			return nil
		}
		return err
	}
	defer s.Close()
	return nil
}

// InstallUpdater execute install the updater orya
func (l *WindowsOperations) InstallUpdater(version string) error {
	repoUrl := utils.GetBinariesRepositoryUrl()
	updaterUrl := fmt.Sprintf("%s/%s/updater-windows-amd64.exe", repoUrl, version)

	respUpdater, _, err := utils.GetBinary(updaterUrl)
	if err != nil {
		return err
	}

	workdir, err := utils.GetWorkDirPath()
	if err != nil {
		return err
	}
	pathUpdater := filepath.Join(workdir, "bin", "releases", version)
	if err := l.filesystem.VerifyDirExistAndCreate(pathUpdater); err != nil {
		return err
	}
	pathUpdaterExe := filepath.Join(pathUpdater, "updater.exe")
	if err := l.filesystem.WriteBinaryContent(pathUpdaterExe, respUpdater); err != nil {
		return err
	}
	pathLogOut := filepath.Join(workdir, "logs", "updater.out.txt")
	pathLogErr := filepath.Join(workdir, "logs", "updater.err.txt")
	envVersion := fmt.Sprintf("$env:VERSION ='%s'", version)
	command := fmt.Sprintf(`%s; Start-Process -FilePath "%s" -RedirectStandardOutput "%s" -RedirectStandardError "%s" -NoNewWindow`, envVersion, pathUpdaterExe, pathLogOut, pathLogErr)
	out, err := l.program.ExecuteWithOutput("powershell", []string{}, "-Command", command)
	if err != nil {
		l.logger.Error("error in get wmi orya agent", "error", err)
		return err
	}

	l.logger.Debug("output wmi orya agent", "output", out)
	return nil
}

func (l *WindowsOperations) UninstallUpdater(version string) error {
	return nil
}

// UpdateAgent execute update the agent orya
func (l *WindowsOperations) UpdateAgent(version string) error {
	go l.InstallUpdater(version)
	return nil
}

// ExecuteUpdateVersion execute update version the agent orya
func (l *WindowsOperations) ExecuteUpdateVersion(version string) error {

	//get msi agent and manager
	repoUrl := utils.GetBinariesRepositoryUrl()
	managerUrl := fmt.Sprintf("%s/%s/install_manager_windows.msi", repoUrl, version)
	agentUrl := fmt.Sprintf("%s/%s/install_agent_windows.msi", repoUrl, version)
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
	if err := l.filesystem.VerifyDirExistAndCreate(pathVersion); err != nil {
		return err
	}

	pathManagerMsi := filepath.Join(pathVersion, "manager.msi")
	pathAgentMsi := filepath.Join(pathVersion, "agent.msi")

	if err := l.filesystem.WriteBinaryContent(pathManagerMsi, respManager); err != nil {
		return err
	}
	if err := l.filesystem.WriteBinaryContent(pathAgentMsi, respAgent); err != nil {
		return err
	}

	if err := l.StopService("agent"); err != nil {
		return err
	}

	if err := l.StopService("manager"); err != nil {
		return err
	}

	commandManager := fmt.Sprintf(`start-process -Wait msiexec -ArgumentList '/qn /i "%s"'`, pathManagerMsi)
	commandAgent := fmt.Sprintf(`start-process -Wait msiexec -ArgumentList '/qn /i "%s"'`, pathAgentMsi)

	envVersion := fmt.Sprintf("$env:VERSION ='%s'", version)

	l.logger.Debug("commandManager", "commandManager", commandManager)
	outManager, err := l.program.ExecuteWithOutput("powershell", []string{envVersion}, "-Command", commandManager)
	if err != nil {
		l.logger.Error("error in update manager start process", "error", err)
		return err
	}
	l.logger.Debug("update manager", "output", outManager)
	time.Sleep(10 * time.Second)
	l.logger.Debug("commandAgent", "commandAgent", commandAgent)
	outAgent, err := l.program.ExecuteWithOutput("powershell", []string{envVersion}, "-Command", commandAgent)
	if err != nil {
		l.logger.Error("error in update agent start process", "error", err)
		return err
	}
	l.logger.Debug("update agent", "output", outAgent)

	//restart service agent and manager
	if err := l.RestartService("agent"); err != nil {
		return err
	}

	if err := l.RestartService("manager"); err != nil {
		return err
	}

	return nil
}

// ExecuteRollbackVersion execute rollback the version
func (l *WindowsOperations) ExecuteRollbackVersion(version string) error {
	statusManager, err := l.Status("manager")
	if err != nil {
		return err
	}

	statusAgent, err := l.Status("agent")
	if err != nil {
		return err
	}
	if statusAgent != "active" {
		if err := l.RestartService("agent"); err != nil {
			return err
		}
	}
	if statusManager != "active" {
		if err := l.RestartService("manager"); err != nil {
			return err
		}
	}
	return nil
}

// UpdaterUninstall execute uninstall the updater orya
func (l *WindowsOperations) UpdaterUninstall(version string) error {
	workdir, err := utils.GetWorkDirPath()
	if err != nil {
		return err
	}
	// Nome do executável atual
	pathVersion := filepath.Join(workdir, "bin", "releases", version)
	pathUpdaterExe := filepath.Join(pathVersion, "updater.exe")

	// Caminho do diretório
	dir := filepath.Dir(pathUpdaterExe)
	l.logger.Debug("diretório do executável", "dir", dir)
	// Nome do arquivo de script temporário
	batFileName := fmt.Sprintf("del_%d.bat", time.Now().Unix())
	batPath := filepath.Join(dir, batFileName)
	l.logger.Debug("caminho do script .bat", "batPath", batPath)

	batContent := fmt.Sprintf(
		"@echo off\n"+
			"timeout /t 5 /nobreak > nul\n"+ // Espera 5 segundos para o processo principal sair
			"taskkill /f /im updater.exe > nul 2>&1\n"+ // Tenta matar o processo updater.exe, caso ainda esteja rodando
			"del /f /q \"%s\"\n"+ // Apaga o seu executável
			"del /f /q \"%s\"", // Apaga o próprio script .bat
		pathUpdaterExe,
		batPath,
	)

	// Cria o arquivo de script
	err = os.WriteFile(batPath, []byte(batContent), 0644)
	if err != nil {
		return err
	}

	// Executa o script e sai imediatamente
	// O uso de "cmd /C" garante que o script seja executado em um novo shell e não bloqueie o programa atual
	cmd := exec.Command("cmd", "/C", batPath)
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}

// UninstallAgent execute uninstall the agent orya
func (l *WindowsOperations) UninstallAgent(version string) error {
	identifier, err := l.GetIdentifierService("OryaAgent")
	if err != nil {
		return err
	}

	outUnistall, err := l.program.ExecuteWithOutput("powershell", []string{}, "-Command", fmt.Sprintf(`start-process msiexec -Wait -ArgumentList ('/log', 'C:\uninst.log', '/norestart', '/q', '/x', '%s')`, identifier))
	if err != nil {
		l.logger.Error("error in uninstall orya agent", "error", err)
		return err
	}
	l.logger.Debug("output uninstall orya agent", "output", outUnistall)
	return nil
}

// GetIdentifierService return the identifier number of service
func (l *WindowsOperations) GetIdentifierService(name string) (string, error) {
	wmiCmd := fmt.Sprintf(`(Get-Package -Name "%s").Metadata['ProductCode']`, name)
	out, err := l.program.ExecuteWithOutput("powershell", []string{}, "-Command", wmiCmd)
	if err != nil {
		l.logger.Error("error in get wmi orya agent", "error", err)
		return "", err
	}
	l.logger.Debug("output wmi orya manager", "output", out)
	re := regexp.MustCompile(`\{[A-Fa-f0-9\-]+\}`)
	identifyingNumber := re.FindString(string(out))

	if identifyingNumber == "" {
		return "", fmt.Errorf("not found identify number for orya agent")
	}
	return identifyingNumber, nil
}

// AutoUninstall execute auto uninstall the manager
func (l *WindowsOperations) AutoUninstall(version string) error {
	workdir, err := utils.GetWorkDirPath()
	if err != nil {
		return err
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {

	}

	batPath := filepath.Join(homeDir, "autouninstall.bat")

	identifierAgent, err := l.GetIdentifierService("OryaAgent")
	if err != nil {
		return err
	}
	identifierManager, err := l.GetIdentifierService("OryaManager")
	if err != nil {
		return err
	}
	batContent := fmt.Sprintf(
		"@echo off\n"+
			"timeout /t 5 /nobreak > nul\n"+
			"taskkill /f /im agent.exe > nul 2>&1\n"+
			"taskkill /f /im manager.exe > nul 2>&1\n"+
			"msiexec /x %s /quiet /norestart\n"+
			"msiexec /x %s /quiet /norestart\n"+
			"sc.exe delete OryaAgent > nul 2>&1\n"+
			"sc.exe delete OryaManager > nul 2>&1\n"+
			"del /f /q \"%s\"\n"+
			"rmdir /s /q \"%s\"",
		identifierAgent,
		identifierManager,
		batPath,
		workdir,
	)

	err = os.WriteFile(batPath, []byte(batContent), 0644)
	if err != nil {
		return err
	}

	cmd := exec.Command("cmd", "/C", batPath)
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil

}

// Execute run handler functions the operation
func (l *WindowsOperations) Execute() error {
	if err := l.scmManager.Run(); err != nil {
		return err
	}
	return nil
}
