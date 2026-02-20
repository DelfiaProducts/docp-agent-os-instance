package adapters

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/OryaHub/agent-os-instance/libs/dto"
	"github.com/OryaHub/agent-os-instance/libs/utils"
)

// Collect execute collect the metrics the host
func (l *ManagerAdapter) Collect() <-chan []byte {
	l.logger.Info("collect metadata", "timestamp", time.Now().UTC())
	l.logger.Debug("collect metadata", "trace", "agent-os-instance.manager_adapter.Collect")
	l.wg.Add(2)
	go l.firstGetInfos()
	go l.start()
	return l.chanMetadata
}

// GetAgentVersion return version installed agent
func (l *ManagerAdapter) GetAgentVersion() (string, error) {
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
func (l *ManagerAdapter) GetAgentRollbackVersion() (string, error) {
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

// GetAgentVersionFromSignalBytes return version from signal bytes
func (l *ManagerAdapter) GetAgentVersionFromSignalBytes(data []byte) (string, error) {
	var stateCheckResponse dto.StateCheckResponse
	var version string
	if len(data) > 0 {
		if err := l.json.Unmarshall(data, &stateCheckResponse); err != nil {
			return version, err
		}

		if len(stateCheckResponse.Signal.Agents.OryaAgent.Version) > 0 {
			version = stateCheckResponse.Signal.Agents.OryaAgent.Version
		}
	}
	return version, nil
}

// GetAgentVersionDatadogFromSignalBytes return version the datadog from signal bytes
func (l *ManagerAdapter) GetAgentVersionDatadogFromSignalBytes(data []byte) (string, error) {
	var stateCheckResponse dto.StateCheckResponse
	var version string
	if len(data) > 0 {
		if err := l.json.Unmarshall(data, &stateCheckResponse); err != nil {
			return version, err
		}

		if len(stateCheckResponse.Signal.Agents.DatadogAgent.Version) > 0 {
			version = stateCheckResponse.Signal.Agents.DatadogAgent.Version
		}
	}
	return version, nil
}

// FetchAgentVersions fetches the available agent versions
func (l *ManagerAdapter) FetchAgentVersions() (dto.AgentVersions, error) {
	l.logger.Info("fetch agent versions", "timestamp", time.Now().UTC())
	l.logger.Debug("fetch agent versions", "trace", "agent-os-instance.manager_adapter.FetchAgentVersions")
	agentVersions, err := l.utilityService.FetchAgentVersions()
	if err != nil {
		return dto.AgentVersions{}, err
	}
	return agentVersions, nil
}

// GetStateReceived return state received from state check service
func (l *ManagerAdapter) GetStateReceived() ([]byte, error) {
	l.logger.Debug("get state received", "trace", "agent-os-instance.manager_adapter.GetStateReceived")
	res, err := l.fileSystem.GetFileContent(filepath.Join(l.agentWorkDir, "state", "received"))
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		res = []byte("{}")
	}
	return res, nil
}

// GetStateCurrent return state current from file
func (l *ManagerAdapter) GetStateCurrent() ([]byte, error) {
	l.logger.Debug("get state current", "trace", "agent-os-instance.manager_adapter.GetStateCurrent")
	res, err := l.fileSystem.GetFileContent(filepath.Join(l.agentWorkDir, "state", "current"))
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		res = []byte("{}")
	}
	return res, nil
}

// GetRemoveOtherVendors return other vendors from signal
func (l *ManagerAdapter) GetRemoveOtherVendors() ([]string, error) {
	var state dto.StateCheckResponse
	received, err := l.GetStateReceived()
	if err != nil {
		return nil, err
	}

	if err := l.json.Unmarshall(received, &state); err != nil {
		return nil, err
	}

	otherVendors := state.Signal.RemoveOtherVendors

	return otherVendors, nil
}

// GetActions return actions for agents
func (l *ManagerAdapter) GetActions(stateCheckResponse *dto.StateCheckResponse) ([]dto.StateAction, error) {
	l.logger.Debug("get actions", "trace", "agent-os-instance.manager_adapter.GetActions", "stateCheckResponse", stateCheckResponse)
	var arrStateActions []dto.StateAction

	// prepare orya agent action
	oryaAgentAction := l.prepareOryaAgentAction(stateCheckResponse.Signal)
	lastOryaAgentActionHash := l.GetStore("action.orya.state")
	oryaAgentActionBytes, err := l.json.Marshall(&oryaAgentAction)
	if err != nil {
		return nil, err
	}

	newOryaAgentActionHash := utils.GenerateMd5Hash(oryaAgentActionBytes)
	if newOryaAgentActionHash != lastOryaAgentActionHash {
		arrStateActions = append(arrStateActions, oryaAgentAction)
		if err := l.SetStore("action.orya.state", newOryaAgentActionHash); err != nil {
			return nil, err
		}
	}
	// validate datadog installed
	alreadyInstalled, err := l.AlreadyInstalled("datadog")
	if err != nil {
		return nil, err
	}

	// prepare datadog agent action
	agentDatadogAction := l.prepareAgentDatadogAction(stateCheckResponse.Signal)
	lastAgentDatadogActionHash := l.GetStore("action.datadog.state")
	datadogAgentActionBytes, err := l.json.Marshall(&agentDatadogAction)
	if err != nil {
		return nil, err
	}

	newAgentDatadogActionHash := fmt.Sprintf("%s.%v", utils.GenerateMd5Hash(datadogAgentActionBytes), alreadyInstalled)
	if newAgentDatadogActionHash != lastAgentDatadogActionHash {
		arrStateActions = append(arrStateActions, agentDatadogAction)
		if err := l.SetStore("action.datadog.state", newAgentDatadogActionHash); err != nil {
			return nil, err
		}
	}

	// prepare datadog update agent action
	agentDatadogUpdateAction := l.prepareAgentDatadogUpdateAction(stateCheckResponse.Signal)
	lastAgentDatadogUpdateActionHash := l.GetStore("action.datadog.update")
	datadogUpdateAgentActionBytes, err := l.json.Marshall(&agentDatadogUpdateAction)
	if err != nil {
		return nil, err
	}

	newAgentDatadogUpdateActionHash := utils.GenerateMd5Hash(datadogUpdateAgentActionBytes)
	if newAgentDatadogUpdateActionHash != lastAgentDatadogUpdateActionHash {
		if alreadyInstalled {
			arrStateActions = append(arrStateActions, agentDatadogUpdateAction)
			if err := l.SetStore("action.datadog.update", newAgentDatadogUpdateActionHash); err != nil {
				return nil, err
			}
		}
	}

	// prepare datadog tracer library action
	tracerDatadogLibraryAction := l.prepareTracerDatadogLibraryAction(stateCheckResponse.Signal)
	lastTracerDatadogLibraryActionHash := l.GetStore("action.datadog.tracer.library")
	tracerDatadogLibraryActionBytes, err := l.json.Marshall(&tracerDatadogLibraryAction)
	if err != nil {
		return nil, err
	}

	newTracerDatadogLibraryActionHash := utils.GenerateMd5Hash(tracerDatadogLibraryActionBytes)
	if newTracerDatadogLibraryActionHash != lastTracerDatadogLibraryActionHash {
		arrStateActions = append(arrStateActions, tracerDatadogLibraryAction)
		if err := l.SetStore("action.datadog.tracer.library", newTracerDatadogLibraryActionHash); err != nil {
			return nil, err
		}
	}

	// prepare datadog tracer single step action
	tracerDatadogSingleStepAction := l.prepareTracerDatadogSingleStepAction(stateCheckResponse.Signal)
	lastTracerDatadogSingleStepActionHash := l.GetStore("action.datadog.tracer.single.step")
	tracerDatadogSingleStepActionBytes, err := l.json.Marshall(&tracerDatadogSingleStepAction)
	if err != nil {
		return nil, err
	}

	newTracerDatadogSingleStepActionHash := utils.GenerateMd5Hash(tracerDatadogSingleStepActionBytes)
	if newTracerDatadogSingleStepActionHash != lastTracerDatadogSingleStepActionHash {
		arrStateActions = append(arrStateActions, tracerDatadogSingleStepAction)
		if err := l.SetStore("action.datadog.tracer.single.step", newTracerDatadogSingleStepActionHash); err != nil {
			return nil, err
		}
	}

	//prepare orya standby agent action
	oryaStandbyAgentAction := l.prepareOryaStandbyAgentAction(stateCheckResponse.Signal)
	lastOryaStandbyAgentActionHash := l.GetStore("action.orya.standby")
	oryaStandbyAgentActionBytes, err := l.json.Marshall(&oryaStandbyAgentAction)
	if err != nil {
		return nil, err
	}

	newOryaStandbyAgentActionHash := utils.GenerateMd5Hash(oryaStandbyAgentActionBytes)
	if newOryaStandbyAgentActionHash != lastOryaStandbyAgentActionHash {
		arrStateActions = append(arrStateActions, oryaStandbyAgentAction)
		if err := l.SetStore("action.orya.standby", newOryaStandbyAgentActionHash); err != nil {
			return nil, err
		}
	}

	arrStateActionsFiltered := l.removeAgentDatadogIfTracerSingleStepExists(arrStateActions)
	l.logger.Debug("get actions", "trace", "agent-os-instance.manager_adapter.GetActions", "oryaAgentAction", oryaAgentAction, "agentDatadogAction", agentDatadogAction, "agentDatadogUpdateAction", agentDatadogUpdateAction, "tracerDatadogLibraryAction", tracerDatadogLibraryAction, "tracerDatadogSingleStepAction", tracerDatadogSingleStepAction)
	l.logger.Debug("get actions", "trace", "agent-os-instance.manager_adapter.GetActions", "arrStateActions", arrStateActions)

	return arrStateActionsFiltered, nil
}

// GetState get state from state check
func (l *ManagerAdapter) GetState() ([]byte, error) {
	l.logger.Debug("get state", "trace", "agent-os-instance.manager_adapter.GetState")
	var stateData dto.StateCheckResponse

	// get state received
	pathStateReceived := filepath.Join(l.agentWorkDir, "state", "received")
	content, err := l.fileSystem.GetFileContent(pathStateReceived)
	if err != nil {
		return nil, err
	}

	if err := l.json.Unmarshall(content, &stateData); err != nil {
		return nil, err
	}

	stateActions, err := l.GetActions(&stateData)
	if err != nil {
		return nil, err
	}

	bStateActions, err := l.json.Marshall(&stateActions)
	if err != nil {
		return nil, err
	}

	return bStateActions, nil
}

// GetAlreadyTracer return already tracer from config file
func (l *ManagerAdapter) GetAlreadyTracer() (bool, error) {
	l.logger.Debug("get already tracer", "trace", "agent-os-instance.manager_adapter.GetAlreadyTracer")
	var configAgent dto.ConfigAgent
	pathConfigFile := filepath.Join(l.agentWorkDir, "config.yml")
	contentConfigAgent, err := l.fileSystem.GetFileContent(pathConfigFile)
	if err != nil {
		return false, err
	}
	if err := l.ymlClient.Unmarshall(contentConfigAgent, &configAgent); err != nil {
		return false, err
	}

	return configAgent.AlreadyTracer, nil
}

// GetAlreadyTracer return already created agent from config file
func (l *ManagerAdapter) GetAlreadyCreated() (bool, error) {
	l.logger.Debug("get already tracer", "trace", "agent-os-instance.manager_adapter.GetAlreadyCreated")
	var configAgent dto.ConfigAgent
	pathConfigFile := filepath.Join(l.agentWorkDir, "config.yml")
	contentConfigAgent, err := l.fileSystem.GetFileContent(pathConfigFile)
	if err != nil {
		return false, err
	}
	if err := l.ymlClient.Unmarshall(contentConfigAgent, &configAgent); err != nil {
		return false, err
	}

	return configAgent.AlreadyCreated, nil
}

// GetConfigAgent get config agent from file
func (l *ManagerAdapter) GetConfigAgent() (dto.ConfigAgent, error) {
	l.logger.Debug("get config agent", "trace", "agent-os-instance.manager_adapter.GetConfigAgent")
	var configAgent dto.ConfigAgent
	pathConfigFile := filepath.Join(l.agentWorkDir, "config.yml")
	contentConfigAgent, err := l.fileSystem.GetFileContent(pathConfigFile)
	if err != nil {
		return dto.ConfigAgent{}, err
	}
	if err := l.ymlClient.Unmarshall(contentConfigAgent, &configAgent); err != nil {
		return dto.ConfigAgent{}, err
	}
	return configAgent, nil
}

// GetStore return value from store
func (l *ManagerAdapter) GetStore(key string) any {
	return l.store.Get(key)
}
