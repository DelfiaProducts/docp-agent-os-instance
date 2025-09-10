package adapters

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/components"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/dto"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/pkg"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/services"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

var (
	CURL_UPDATER_LINUX_UNINSTALL_SH = "curl -L https://test-docp-agent-data.s3.amazonaws.com/installer/uninstall_updater_linux.sh | bash"
)

// UpdaterAdapter is struct for updater adapter
type UpdaterAdapter struct {
	chanClose      chan struct{}
	agentWorkDir   string
	isClosed       bool
	wg             *sync.WaitGroup
	logger         interfaces.ILogger
	osOperation    interfaces.IOSOperation
	fileSystem     *pkg.FileSystem
	program        *pkg.ExecProgram
	ymlClient      *pkg.YmlClient
	utilityService *services.UtilityService
	client         *http.Client
	delay          time.Duration
}

// NewUpdaterAdapter return instance of linux updater adapter
func NewUpdaterAdapter(logger interfaces.ILogger) *UpdaterAdapter {
	return &UpdaterAdapter{
		logger:  logger,
		program: pkg.NewExecProgram(),
		delay:   time.Second * 1,
	}
}

// Prepare configure manager adapter
func (l *UpdaterAdapter) Prepare() error {
	ymlClient := pkg.NewYmlClient()
	chanClose := make(chan struct{}, 1)
	wg := &sync.WaitGroup{}
	l.ymlClient = ymlClient
	l.chanClose = chanClose
	l.isClosed = false
	l.wg = wg
	osOperation, err := components.SystemOperation(l.logger)
	if err != nil {
		return err
	}
	if err := osOperation.Setup(); err != nil {
		return err
	}
	l.osOperation = osOperation
	fileSystem := pkg.NewFileSystem()
	l.fileSystem = fileSystem

	agentWorkDir, err := utils.GetWorkDirPath()
	if err != nil {
		return err
	}
	l.agentWorkDir = agentWorkDir
	utilityService := services.NewUtilityService(l.logger)
	if err := utilityService.Setup(); err != nil {
		return err
	}
	l.utilityService = utilityService
	client := &http.Client{
		Timeout: time.Second * 30,
	}
	l.client = client
	return nil
}

// Close closing collect loop
func (l *UpdaterAdapter) Close() error {
	l.logger.Debug("execute close", "trace", "docp-agent-os-instance.manager_adapter.Close")
	l.chanClose <- struct{}{}
	l.wg.Add(1)
	go l.closeChannels()
	return nil
}

// closeChannels closing channels
func (l *UpdaterAdapter) closeChannels() {
	l.logger.Debug("close channels", "trace", "docp-agent-os-instance.manager_adapter.closeChannels")
	close(l.chanClose)
	l.isClosed = true
	l.wg.Done()
}

// Status return status from systemd api
func (l *UpdaterAdapter) Status(serviceName string) (string, error) {
	l.logger.Debug("execute verify status the service", "trace", "docp-agent-os-instance.updater_adapter.Status", "serviceName", serviceName)
	output, err := l.osOperation.Status(serviceName)
	if err != nil {
		return "", err
	}
	output = strings.ReplaceAll(output, "\"", "")
	l.logger.Debug("execute verify status the service", "trace", "docp-agent-os-instance.updater_adapter.Status", "serviceName", serviceName, "output", output)
	return output, nil
}

// DaemonReload execute daemon reload the service in systemd
func (l *UpdaterAdapter) DaemonReload() error {
	l.logger.Debug("daemon reload", "trace", "docp-agent-os-instance.updater_adapter.DaemonReload")
	if err := l.osOperation.DaemonReload(); err != nil {
		return err
	}
	return nil
}

// RestartService execute restart the service in systemd
func (l *UpdaterAdapter) RestartService(serviceName string) error {
	l.logger.Debug("restart service", "trace", "docp-agent-os-instance.manager_adapter.RestartService")
	if err := l.osOperation.RestartService(serviceName); err != nil {
		return err
	}
	return nil
}

// StopService execute stop the service in systemd
func (l *UpdaterAdapter) StopService(serviceName string) error {
	l.logger.Debug("stop service", "trace", "docp-agent-os-instance.manager_adapter.StopService")
	if err := l.osOperation.StopService(serviceName); err != nil {
		return err
	}
	return nil
}

// FetchAgentVersions fetches the available agent versions
func (l *UpdaterAdapter) FetchAgentVersions() (dto.AgentVersions, error) {
	l.logger.Debug("fetch agent versions", "trace", "docp-agent-os-instance.manager_adapter.FetchAgentVersions")
	agentVersions, err := l.utilityService.FetchAgentVersions()
	if err != nil {
		return dto.AgentVersions{}, err
	}
	return agentVersions, nil
}

// ExecuteUpdateVersion execute update the version
func (l *UpdaterAdapter) ExecuteUpdateVersion(version string) error {
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
	case "darwin":
		switch arch {
		case "amd64":
			managerUrl = fmt.Sprintf("%s/%s/manager-macos-amd64", repoUrl, version)
			agentUrl = fmt.Sprintf("%s/%s/agent-macos-amd64", repoUrl, version)
		case "arm64":
			managerUrl = fmt.Sprintf("%s/%s/manager-macos-arm64", repoUrl, version)
			agentUrl = fmt.Sprintf("%s/%s/agent-macos-arm64", repoUrl, version)
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

// GetContentReceived return content received from update
func (l *UpdaterAdapter) GetContentReceived() ([]byte, error) {
	workdir, err := utils.GetWorkDirPath()
	if err != nil {
		return nil, err
	}
	res, err := l.fileSystem.GetFileContent(filepath.Join(workdir, "state", "received"))
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetAgentVersionFromSignal return agent version from signal
func (l *UpdaterAdapter) GetAgentVersionFromSignal(response []byte) (string, error) {
	var signal dto.StateCheckResponse
	if err := json.Unmarshal(response, &signal); err != nil {
		return "", err
	}
	agent := signal.Signal.Agents.DocpAgent
	if len(agent.Version) > 0 {
		return agent.Version, nil
	}
	return "", utils.ErrAgentVersionNotFound()
}

// GetAgentVersion return rollback version installed agent
func (l *UpdaterAdapter) GetAgentVersion() (string, error) {
	var configAgent dto.ConfigAgent
	configPath, err := utils.GetConfigFilePath()
	if err != nil {
		return "", err
	}

	content, err := l.fileSystem.GetFileContent(configPath)
	if err != nil {
		return "", err
	}

	if err := l.ymlClient.Unmarshall(content, &configAgent); err != nil {
		return "", err
	}

	return configAgent.Version, nil
}

// GetAgentRollbackVersion return rollback version installed agent
func (l *UpdaterAdapter) GetAgentRollbackVersion() (string, error) {
	var configAgent dto.ConfigAgent
	configPath, err := utils.GetConfigFilePath()
	if err != nil {
		return "", err
	}

	content, err := l.fileSystem.GetFileContent(configPath)
	if err != nil {
		return "", err
	}

	if err := l.ymlClient.Unmarshall(content, &configAgent); err != nil {
		return "", err
	}

	return configAgent.RollbackVersion, nil
}

// ValidateSuccessUpdated validate if update was successful
func (l *UpdaterAdapter) ValidateSuccessUpdated() (bool, error) {
	activeManager, err := l.osOperation.Status("manager")
	if err != nil {
		return false, err
	}
	activeAgent, err := l.osOperation.Status("agent")
	if err != nil {
		return false, err
	}
	return activeManager == "active" && activeAgent == "active", nil
}

// ExecuteRollbackVersion execute rollback to previous version
func (l *UpdaterAdapter) ExecuteRollbackVersion(version string) error {
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

// UpdaterUninstall execute uninstall the updater
func (l *UpdaterAdapter) UpdaterUninstall() error {
	job := fmt.Sprintf("* * * * * %s; crontab -l | grep -v '%s' | crontab -", CURL_UPDATER_LINUX_UNINSTALL_SH, CURL_UPDATER_LINUX_UNINSTALL_SH)
	command := fmt.Sprintf("(crontab -l 2>/dev/null; echo \"%s\") | crontab -", job)
	if err := l.program.Execute("bash", []string{}, "-c", command); err != nil {
		return err
	}
	return nil
}

// HandlerSCMManager execute handler for scm manager
func (l *UpdaterAdapter) HandlerSCMManager() error {
	if err := l.osOperation.Execute(); err != nil {
		return err
	}
	return nil
}
