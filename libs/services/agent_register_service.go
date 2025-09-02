package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/dto"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/pkg"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

// AgentRegisterService is struct for register service
type AgentRegisterService struct {
	urlRegister    string
	workDirPath    string
	fileSystem     *pkg.FileSystem
	ymlClient      *pkg.YmlClient
	configFilePath string
	logger         interfaces.ILogger
	client         *http.Client
}

// NewAgentRegisterService return instance of agent the register service
func NewAgentRegisterService(logger interfaces.ILogger) *AgentRegisterService {
	return &AgentRegisterService{
		logger: logger,
	}
}

// Setup execute configurations for register service
func (ag *AgentRegisterService) Setup() error {
	urlRegister, err := utils.GetDomainUrl()
	if err != nil {
		return err
	}
	ag.urlRegister = urlRegister
	configFilePath, err := utils.GetConfigFilePath()
	if err != nil {
		return err
	}

	workDirPath, err := utils.GetWorkDirPath()
	if err != nil {
		return err
	}

	ag.configFilePath = configFilePath
	ag.workDirPath = workDirPath
	fileSystem := pkg.NewFileSystem()
	ag.fileSystem = fileSystem
	yamlClient := pkg.NewYmlClient()
	ag.ymlClient = yamlClient
	client := &http.Client{
		Timeout: time.Second * 90,
	}
	ag.client = client
	return nil
}

// marshaller execute marshal the struct for slice the bytes
func (ag *AgentRegisterService) marshaller(inner any) ([]byte, error) {
	ag.logger.Debug("execute marshaller", "trace", "docp-agent-os-instance.agent_register_service.marshaller", "inner", inner)
	resBytes, err := json.Marshal(inner)
	if err != nil {
		ag.logger.Error("error in execute marshaller", "trace", "docp-agent-os-instance.agent_register_service.marshaller", "error", err.Error())
		return nil, err
	}
	return resBytes, nil
}

// unmarshaller execute unmarshal the content bytes
func (ag *AgentRegisterService) unmarshaller(content []byte, inner any) error {
	ag.logger.Debug("execute unmarshaller", "trace", "docp-agent-os-instance.agent_register_service.unmarshaller", "content", string(content), "inner", inner)
	if err := json.Unmarshal(content, inner); err != nil {
		ag.logger.Error("error in execute unmarshaller", "trace", "docp-agent-os-instance.agent_register_service.unmarshaller", "error", err.Error())
		return err
	}
	return nil
}

// unmarshalYml exeuct unmarshal the config file
func (ag *AgentRegisterService) unmarshalYml(content []byte, config *dto.ConfigAgent) error {
	ag.logger.Debug("unmarshall yml", "trace", "docp-agent-os-instance.agent_register_service.unmarshalYml", "content", string(content), "config", config)
	if err := ag.ymlClient.Unmarshall(content, config); err != nil {
		return err
	}
	return nil
}

// prepareToSend exeucte prepare for send data to register service
func (ag *AgentRegisterService) prepareToSendCreate(metadata []byte) ([]byte, string, error) {
	ag.logger.Debug("execute prepare to send", "trace", "docp-agent-os-instance.agent_register_service.prepareToSendCreate", "metadata", string(metadata))
	configFileBytes, err := ag.GetConfigFileContent(ag.configFilePath)
	if err != nil {
		ag.logger.Error("error in prepare to send", "trace", "docp-agent-os-instance.agent_register_service.prepareToSend", "error", err.Error())
		return nil, "", err
	}
	injectedMetadataBytes, apiKey, err := ag.InjectClientInfoCreate(configFileBytes, metadata)
	if err != nil {
		ag.logger.Error("error in prepare to send", "trace", "docp-agent-os-instance.agent_register_service.prepareToSend", "error", err.Error())
		return nil, "", err
	}
	return injectedMetadataBytes, apiKey, nil
}

// prepareToSend exeucte prepare for send data to register service
func (ag *AgentRegisterService) prepareToSendUpdate(metadata []byte) ([]byte, string, error) {
	ag.logger.Debug("execute prepare to send", "trace", "docp-agent-os-instance.agent_register_service.prepareToSendUpdate", "metadata", string(metadata))
	configFileBytes, err := ag.GetConfigFileContent(ag.configFilePath)
	if err != nil {
		ag.logger.Error("error in prepare to send", "trace", "docp-agent-os-instance.agent_register_service.prepareToSend", "error", err.Error())
		return nil, "", err
	}
	injectedMetadataBytes, apiKey, err := ag.InjectClientInfoUpdate(configFileBytes, metadata)
	if err != nil {
		ag.logger.Error("error in prepare to send", "trace", "docp-agent-os-instance.agent_register_service.prepareToSend", "error", err.Error())
		return nil, "", err
	}
	return injectedMetadataBytes, apiKey, nil
}

// GetConfigFileContent get config file content from local
func (ag *AgentRegisterService) GetConfigFileContent(filePath string) ([]byte, error) {
	ag.logger.Debug("execute get config file content", "trace", "docp-agent-os-instance.agent_register_service.GetConfigFileContent", "filePath", filePath)
	if err := ag.fileSystem.VerifyFileExist(filePath); err != nil {
		ag.logger.Error("error in verify file exist", "trace", "docp-agent-os-instance.agent_register_service.GetConfigFileContent", "error", err.Error())
		return nil, err
	}
	content, err := ag.fileSystem.GetFileContent(filePath)
	if err != nil {
		ag.logger.Error("error in get file content", "trace", "docp-agent-os-instance.agent_register_service.GetConfigFileContent", "error", err.Error())
		return nil, err
	}
	return content, nil
}

// InjectClientInfo execute injection the client info in metadata content
func (ag *AgentRegisterService) InjectClientInfoCreate(configFileContent, linuxMetadataContent []byte) ([]byte, string, error) {
	ag.logger.Debug("execute inject client info", "trace", "docp-agent-os-instance.agent_register_service.InjectClientInfoCreate", "configFileContent", string(configFileContent), "linuxMetadataContent", string(linuxMetadataContent))
	var configAgentDto dto.ConfigAgent
	var metadata dto.Metadata
	var agentRegisterDataCreate dto.AgentRegisterDataCreate

	if err := ag.unmarshalYml(configFileContent, &configAgentDto); err != nil {
		ag.logger.Error("error in unmarshaller", "trace", "docp-agent-os-instance.agent_register_service.InjectClientInfo", "error", err.Error())
		return nil, "", err
	}

	if err := ag.unmarshaller(linuxMetadataContent, &metadata); err != nil {
		ag.logger.Error("error in unmarshaller", "trace", "docp-agent-os-instance.agent_register_service.InjectClientInfo", "error", err.Error())
		return nil, "", err
	}

	slcTags, err := utils.TransformMapToSlice(configAgentDto.Agent.Tags)
	if err != nil {
		return nil, "", err
	}

	agentRegisterDataCreate = dto.AgentRegisterDataCreate{
		NoGroupAssociation: configAgentDto.NoGroupAssociation,
		Tags:               slcTags,
		Metadata:           metadata,
		VMName:             metadata.ComputeInfo.Computename,
	}

	registerDataBytes, err := ag.marshaller(agentRegisterDataCreate)
	if err != nil {
		ag.logger.Error("error in marshaller", "trace", "docp-agent-os-instance.agent_register_service.InjectClientInfo", "error", err.Error())
		return nil, "", err
	}

	return registerDataBytes, configAgentDto.Agent.ApiKey, nil
}

// InjectClientInfo execute injection the client info in metadata content
func (ag *AgentRegisterService) InjectClientInfoUpdate(configFileContent, linuxMetadataContent []byte) ([]byte, string, error) {
	ag.logger.Debug("execute inject client info", "trace", "docp-agent-os-instance.agent_register_service.InjectClientInfoUpdate", "configFileContent", string(configFileContent), "linuxMetadataContent", string(linuxMetadataContent))
	var configAgentDto dto.ConfigAgent
	var metadata dto.Metadata
	var agentRegisterDataUpdate dto.AgentRegisterDataUpdate

	if err := ag.unmarshalYml(configFileContent, &configAgentDto); err != nil {
		ag.logger.Error("error in unmarshaller", "trace", "docp-agent-os-instance.agent_register_service.InjectClientInfo", "error", err.Error())
		return nil, "", err
	}

	if err := ag.unmarshaller(linuxMetadataContent, &metadata); err != nil {
		ag.logger.Error("error in unmarshaller", "trace", "docp-agent-os-instance.agent_register_service.InjectClientInfo", "error", err.Error())
		return nil, "", err
	}

	slcTags, err := utils.TransformMapToSlice(configAgentDto.Agent.Tags)
	if err != nil {
		return nil, "", err
	}

	agentRegisterDataUpdate = dto.AgentRegisterDataUpdate{
		Tags:     slcTags,
		Metadata: metadata,
		VMName:   metadata.ComputeInfo.Computename,
	}

	registerDataBytes, err := ag.marshaller(agentRegisterDataUpdate)
	if err != nil {
		ag.logger.Error("error in marshaller", "trace", "docp-agent-os-instance.agent_register_service.InjectClientInfo", "error", err.Error())
		return nil, "", err
	}

	return registerDataBytes, configAgentDto.AccessToken, nil
}

// SendMetadataCreate execute send metadata to create initial host in
// register service
func (ag *AgentRegisterService) SendMetadataCreate(data []byte) ([]byte, int, error) {
	ag.logger.Debug("execute send metadata create", "trace", "docp-agent-os-instance.agent_register_service.SendMetadataCreate", "data", string(data))

	injectedMetadata, apiKey, err := ag.prepareToSendCreate(data)
	if err != nil {
		ag.logger.Error("error in prepare to send", "trace", "docp-agent-os-instance.agent_register_service.SendMetadata", "error", err.Error())
		return nil, 0, err
	}

	urlMetadataCreate := fmt.Sprintf("%s/compute/v1/docp", ag.urlRegister)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlMetadataCreate, bytes.NewBuffer(injectedMetadata))
	if err != nil {
		ag.logger.Error("error in create request", "trace", "docp-agent-os-instance.agent_register_service.SendMetadata", "error", err.Error())
		return nil, 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("docp-api-key", apiKey)

	res, err := ag.client.Do(req)
	if err != nil {
		ag.logger.Error("error in execute request", "trace", "docp-agent-os-instance.agent_register_service.SendMetadata", "error", err.Error())
		return nil, 0, err
	}

	defer res.Body.Close()

	respBytes, err := io.ReadAll(res.Body)
	if err != nil {
		ag.logger.Error("error in read body response", "trace", "docp-agent-os-instance.agent_register_service.SendMetadata", "error", err.Error())
		return nil, 0, err
	}

	return respBytes, res.StatusCode, nil
}

// SendMetadataUpdate execute send metadata to update host in
// register service
func (ag *AgentRegisterService) SendMetadataUpdate(data []byte) ([]byte, int, error) {
	ag.logger.Debug("execute send metadata update", "trace", "docp-agent-os-instance.agent_register_service.SendMetadataUpdate", "data", string(data))

	injectedMetadata, token, err := ag.prepareToSendUpdate(data)
	if err != nil {
		ag.logger.Error("error in prepare to send", "trace", "docp-agent-os-instance.agent_register_service.SendMetadata", "error", err.Error())
		return nil, 0, err
	}

	urlMetadataUpdate := fmt.Sprintf("%s/compute/v1/docp", ag.urlRegister)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, urlMetadataUpdate, bytes.NewBuffer(injectedMetadata))
	if err != nil {
		ag.logger.Error("error in create request", "trace", "docp-agent-os-instance.agent_register_service.SendMetadata", "error", err.Error())
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	res, err := ag.client.Do(req)
	if err != nil {
		ag.logger.Error("error in execute request", "trace", "docp-agent-os-instance.agent_register_service.SendMetadata", "error", err.Error())
		return nil, 0, err
	}
	defer res.Body.Close()
	respBytes, err := io.ReadAll(res.Body)
	if err != nil {
		ag.logger.Error("error in read body response", "trace", "docp-agent-os-instance.agent_register_service.SendMetadata", "error", err.Error())
		return nil, 0, err
	}
	return respBytes, res.StatusCode, nil
}
