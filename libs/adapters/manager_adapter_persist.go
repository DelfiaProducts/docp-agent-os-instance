package adapters

import (
	"path/filepath"
	"time"

	"github.com/OryaHub/agent-os-instance/libs/dto"
	"github.com/OryaHub/agent-os-instance/libs/utils"
)

// SaveAgentVersion save version installed agent
func (l *ManagerAdapter) SaveAgentVersion(version string) error {
	var configAgent dto.ConfigAgent
	configPath, err := utils.GetConfigFilePath()
	if err != nil {
		return err
	}

	content, err := l.fileSystem.GetFileContent(configPath)
	if err != nil {
		return err
	}

	if err := l.ymlClient.Unmarshall(content, &configAgent); err != nil {
		return err
	}

	configAgent.Version = version

	ymlBytes, err := l.ymlClient.Marshall(&configAgent)
	if err != nil {
		return err
	}
	if err := l.fileSystem.WriteFileContent(configPath, ymlBytes); err != nil {
		return err
	}

	return nil
}

// SaveAgentRollbackVersion save rollback version installed agent
func (l *ManagerAdapter) SaveAgentRollbackVersion(version string) error {
	var configAgent dto.ConfigAgent
	configPath, err := utils.GetConfigFilePath()
	if err != nil {
		return err
	}

	content, err := l.fileSystem.GetFileContent(configPath)
	if err != nil {
		return err
	}

	if err := l.ymlClient.Unmarshall(content, &configAgent); err != nil {
		return err
	}

	configAgent.RollbackVersion = version

	ymlBytes, err := l.ymlClient.Marshall(&configAgent)
	if err != nil {
		return err
	}
	if err := l.fileSystem.WriteFileContent(configPath, ymlBytes); err != nil {
		return err
	}

	return nil
}

// SaveStateReceived execute save the state received
func (l *ManagerAdapter) SaveStateReceived(stateData []byte) error {
	l.logger.Debug("save state received", "trace", "agent-os-instance.manager_adapter.SaveStateReceived", "stateData", string(stateData))
	if err := l.fileSystem.WriteFileContent(filepath.Join(l.agentWorkDir, "state", "received"), stateData); err != nil {
		return err
	}
	return nil
}

// SaveStateCurrent execute save the state current
func (l *ManagerAdapter) SaveStateCurrent(stateData []byte) error {
	l.logger.Debug("save state current", "trace", "agent-os-instance.manager_adapter.SaveStateCurrent", "stateData", string(stateData))
	if err := l.fileSystem.WriteFileContent(filepath.Join(l.agentWorkDir, "state", "current"), stateData); err != nil {
		return err
	}
	return nil
}

// SaveState save state from state check
func (l *ManagerAdapter) SaveState(data []byte) error {
	l.logger.Debug("save state", "trace", "agent-os-instance.manager_adapter.SaveState", "data", string(data))
	if err := l.SaveStateReceived(data); err != nil {
		return err
	}
	return nil
}

// SaveAlreadyTracer save already tracer on config file
func (l *ManagerAdapter) SaveAlreadyTracer(value bool) error {
	l.logger.Debug("save already tracer", "trace", "agent-os-instance.manager_adapter.SaveAlreadyTracer")
	var configAgent dto.ConfigAgent
	pathConfigFile := filepath.Join(l.agentWorkDir, "config.yml")
	contentConfigAgent, err := l.fileSystem.GetFileContent(pathConfigFile)
	if err != nil {
		return err
	}
	if err := l.ymlClient.Unmarshall(contentConfigAgent, &configAgent); err != nil {
		return err
	}

	configAgent.AlreadyTracer = value

	newConfigAgentBytes, err := l.ymlClient.Marshall(&configAgent)
	if err != nil {
		return err
	}
	if err := l.fileSystem.WriteFileContent(pathConfigFile, newConfigAgentBytes); err != nil {
		return err
	}
	return nil
}

// SaveInitialConfigFromRegister save initial config from register in file
func (l *ManagerAdapter) SaveInitialConfigFromRegister(data []byte) error {
	l.logger.Debug("save initial config from register", "trace", "agent-os-instance.manager_adapter.SaveInitialConfigFromRegister", "data", string(data))
	var configAgent dto.ConfigAgent
	var registerDataResponseSuccess dto.AgentRegisterDataResponseSuccess
	pathConfigFile := filepath.Join(l.agentWorkDir, "config.yml")
	content, err := l.fileSystem.GetFileContent(pathConfigFile)
	if err != nil {
		return err
	}
	if err := l.ymlClient.Unmarshall(content, &configAgent); err != nil {
		return err
	}
	if err := l.json.Unmarshall(data, &registerDataResponseSuccess); err != nil {
		return err
	}
	token := registerDataResponseSuccess.AccessToken
	claims, err := utils.DecodeJwt(token)
	if err != nil {
		return err
	}
	configAgent.AccessToken = token
	configAgent.ComputeId = claims.ComputeId
	configAgent.OryaOrgId = claims.OryaOrgId
	configAgent.AlreadyCreated = true
	configAgent.AlreadyTracer = false

	newConfigAgentBytes, err := l.ymlClient.Marshall(&configAgent)
	if err != nil {
		return err
	}
	if err := l.fileSystem.WriteFileContent(pathConfigFile, newConfigAgentBytes); err != nil {
		return err
	}
	return nil
}

// AddTracerLanguage append tracer language in slice the config file
func (l *ManagerAdapter) AddTracerLanguage(language string) error {
	l.logger.Info("add tracer language", "language", language)
	l.logger.Debug("add tracer language", "trace", "agent-os-instance.manager_adapter.AddTracerLanguage", "language", language)
	var configAgent dto.ConfigAgent
	pathConfigFile := filepath.Join(l.agentWorkDir, "config.yml")
	contentConfigAgent, err := l.fileSystem.GetFileContent(pathConfigFile)
	if err != nil {
		return err
	}
	if err := l.ymlClient.Unmarshall(contentConfigAgent, &configAgent); err != nil {
		return err
	}
	slice := configAgent.TracerLanguages
	seen := make(map[string]struct{})

	for _, item := range slice {
		seen[item] = struct{}{}
	}

	if _, exists := seen[language]; !exists {
		slice = append(slice, language)
	}
	configAgent.TracerLanguages = slice

	bytYml, err := l.ymlClient.Marshall(&configAgent)
	if err != nil {
		return err
	}
	if err := l.fileSystem.WriteFileContent(pathConfigFile, bytYml); err != nil {
		return err
	}

	return nil
}

// ClearTracerLanguage append tracer language in slice the config file
func (l *ManagerAdapter) ClearTracerLanguage() error {
	l.logger.Info("clear tracer language", "timestamp", time.Now().UTC())
	l.logger.Debug("clear tracer language", "trace", "agent-os-instance.manager_adapter.ClearTracerLanguage")
	var configAgent dto.ConfigAgent
	pathConfigFile := filepath.Join(l.agentWorkDir, "config.yml")
	contentConfigAgent, err := l.fileSystem.GetFileContent(pathConfigFile)
	if err != nil {
		return err
	}
	if err := l.ymlClient.Unmarshall(contentConfigAgent, &configAgent); err != nil {
		return err
	}
	configAgent.TracerLanguages = []string{}

	bytYml, err := l.ymlClient.Marshall(&configAgent)
	if err != nil {
		return err
	}
	if err := l.fileSystem.WriteFileContent(pathConfigFile, bytYml); err != nil {
		return err
	}

	return nil
}

// UpdateConfigAgent update config agent for file
func (l *ManagerAdapter) UpdateConfigAgent(configAgent dto.ConfigAgent) error {
	l.logger.Debug("update config agent", "trace", "agent-os-instance.manager_adapter.UpdateConfigAgent")
	pathConfigFile := filepath.Join(l.agentWorkDir, "config.yml")

	configBytes, err := l.ymlClient.Marshall(&configAgent)
	if err != nil {
		return err
	}
	if err := l.fileSystem.WriteFileContent(pathConfigFile, configBytes); err != nil {
		return err
	}
	return nil
}

// SetStore execute set key and value to store
func (l *ManagerAdapter) SetStore(key string, val any) error {
	return l.store.Set(key, val)
}
