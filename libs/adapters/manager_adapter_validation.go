package adapters

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/OryaHub/agent-os-instance/libs/dto"
	"github.com/OryaHub/agent-os-instance/libs/utils"
)

// Validade execute validate states
func (l *ManagerAdapter) Validate() error {
	l.logger.Debug("validate", "trace", "agent-os-instance.manager_adapter.Validate")
	oryaAgentIsOK := false
	datadogAgentIsOK := false

	workdir, err := utils.GetWorkDirPath()
	if err != nil {
		return err
	}

	receivedFilePath := filepath.Join(workdir, "state", "received")

	if err := l.fileSystem.VerifyFileExist(receivedFilePath); err != nil {
		return err
	}

	content, err := l.fileSystem.GetFileContent(receivedFilePath)
	if err != nil {
		return err
	}

	var agentState dto.StateCheckResponse
	if err := json.Unmarshal(content, &agentState); err != nil {
		return err
	}
	signal := agentState.Signal
	agent := signal.Agents.OryaAgent
	datadog := signal.Agents.DatadogAgent

	// validate orya agent
	l.logger.Debug("validate", "trace", "agent-os-instance.manager_adapter.Validate", "oryaAgent", agent)
	if len(agent.Version) > 0 {
		status, err := l.osOperation.Status("agent")
		l.logger.Debug("validate", "trace", "agent-os-instance.manager_adapter.Validate", "orya agent status when enabled", status)
		if err != nil {
			return err
		}
		if strings.ReplaceAll(status, "\"", "") == "active" {
			oryaAgentIsOK = true
			l.logger.Debug("validate", "trace", "agent-os-instance.manager_adapter.Validate", "oryaAgentIsOK", oryaAgentIsOK)
		}

	}

	if len(agent.Version) == 0 {
		status, err := l.osOperation.Status("agent")
		l.logger.Debug("validate", "trace", "agent-os-instance.manager_adapter.Validate", "orya agent status not when enabled", status, "oryaAgentIsOk", oryaAgentIsOK)
		if err != nil {
			return err
		}
		if strings.ReplaceAll(status, "\"", "") != "active" {
			oryaAgentIsOK = true
		}

	}
	// validate datadog
	if len(datadog.Version) > 0 {
		status, err := l.osOperation.Status("datadog")
		if err != nil {
			return err
		}
		if strings.ReplaceAll(status, "\"", "") == "active" {
			datadogAgentIsOK = true
		}
	}
	if len(datadog.Version) == 0 {
		status, err := l.osOperation.Status("datadog")
		l.logger.Debug("validate", "trace", "agent-os-instance.manager_adapter.Validate", "datadog agent status", status)
		if err != nil {
			return err
		}
		if strings.ReplaceAll(status, "\"", "") != "active" {
			datadogAgentIsOK = true
		}
	}

	l.logger.Debug("validate", "trace", "agent-os-instance.manager_adapter.Validate", "oryaAgentIsOk", oryaAgentIsOK, "datadogAgentIsOk", datadogAgentIsOK)
	if oryaAgentIsOK && datadogAgentIsOK {
		l.logger.Debug("validate", "trace", "agent-os-instance.manager_adapter.Validate", "status", "validation success")
		if err := l.SaveStateCurrent(content); err != nil {
			return err
		}

		return nil
	}

	l.logger.Debug("validate", "trace", "agent-os-instance.manager_adapter.Validate", "status", "validation failed")
	return nil
}

// CompareState execute compare states between received and current
func (l *ManagerAdapter) CompareState() (bool, error) {
	l.logger.Debug("compare state", "trace", "agent-os-instance.manager_adapter.CompareState")
	receivedBytes, err := l.GetStateReceived()
	if err != nil {
		return false, err
	}
	currentBytes, err := l.GetStateCurrent()
	if err != nil {
		return false, err
	}
	md5ReceivedHash := utils.GenerateMd5Hash(receivedBytes)
	md5CurrentHash := utils.GenerateMd5Hash(currentBytes)
	l.logger.Debug("compare state", "trace", "agent-os-instance.manager_adapter.CompareState", "md5 received hash", md5ReceivedHash, "md5 current hash", md5CurrentHash)
	if md5CurrentHash == md5ReceivedHash {
		return true, nil
	} else {
		return false, nil
	}
}

// IsAlreadyCreated return is already created host
func (l *ManagerAdapter) IsAlreadyCreated() (bool, error) {
	l.logger.Debug("is already created", "trace", "agent-os-instance.manager_adapter.IsAlreadyCreated")
	var configAgent dto.ConfigAgent
	pathConfigFile := filepath.Join(l.agentWorkDir, "config.yml")
	content, err := l.fileSystem.GetFileContent(pathConfigFile)
	if err != nil {
		return false, err
	}
	if err := l.ymlClient.Unmarshall(content, &configAgent); err != nil {
		return false, err
	}
	if configAgent.AlreadyCreated {
		return true, nil
	}

	return false, nil
}

// ExisteOtherVendors verify if exist other vendors
func (l *ManagerAdapter) ExisteOtherVendors() (bool, error) {
	vendors, err := l.GetRemoveOtherVendors()
	if err != nil {
		return false, err
	}
	if vendors != nil && len(vendors) > 0 {
		return true, nil
	}
	return false, nil
}

// AlreadyInstalled return if service already installed
func (l *ManagerAdapter) AlreadyInstalled(serviceName string) (bool, error) {
	l.logger.Info("execute verify already installed service", "serviceName", serviceName)
	l.logger.Debug("execute verify already installed service", "trace", "agent-os-instance.manager_adapter.AlreadyInstalled", "serviceName", serviceName)
	installed, err := l.osOperation.AlreadyInstalled(serviceName)
	if err != nil {
		return false, err
	}
	l.logger.Debug("execute verify already installed service", "trace", "agent-os-instance.manager_adapter.AlreadyInstalled", "serviceName", serviceName, "installed", installed)
	return installed, nil
}

// Status return status from systemd api
func (l *ManagerAdapter) Status(serviceName string) (string, error) {
	l.logger.Info("execute verify status the service", "serviceName", serviceName)
	l.logger.Debug("execute verify status the service", "trace", "agent-os-instance.manager_adapter.Status", "serviceName", serviceName)
	output, err := l.osOperation.Status(serviceName)
	if err != nil {
		return "", err
	}
	output = strings.ReplaceAll(output, "\"", "")
	l.logger.Debug("execute verify status the service", "trace", "agent-os-instance.manager_adapter.Status", "serviceName", serviceName, "output", output)
	return output, nil
}

// IsLocked return if operator locked for send transactions events
func (l *ManagerAdapter) IsLockedEvents() bool {
	return l.LockedEvents
}

// ExistTracerLanguage verify if tracer language exist on host
func (l *ManagerAdapter) ExistTracerLanguage(language string) (bool, error) {
	l.logger.Debug("exist tracer language", "trace", "agent-os-instance.manager_adapter.ExistTracerLanguage", "language", language)
	var configAgent dto.ConfigAgent
	pathConfigFile := filepath.Join(l.agentWorkDir, "config.yml")
	contentConfigAgent, err := l.fileSystem.GetFileContent(pathConfigFile)
	if err != nil {
		return false, err
	}
	if err := l.ymlClient.Unmarshall(contentConfigAgent, &configAgent); err != nil {
		return false, err
	}
	for _, lang := range configAgent.TracerLanguages {
		if lang == language {
			return true, nil
		}
	}

	return false, nil
}
