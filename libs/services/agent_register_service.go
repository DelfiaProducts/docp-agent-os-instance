package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/OryaHub/agent-os-instance/libs/dto"
	"github.com/OryaHub/agent-os-instance/libs/interfaces"
	"github.com/OryaHub/agent-os-instance/libs/pkg"
	"github.com/OryaHub/agent-os-instance/libs/utils"
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
	json           *pkg.JsonClient
}

// NewAgentRegisterService return instance of agent the register service
func NewAgentRegisterService(logger interfaces.ILogger) *AgentRegisterService {
	return &AgentRegisterService{
		logger:    logger,
		json:      pkg.NewJsonClient(),
		ymlClient: pkg.NewYmlClient(),
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
	client := &http.Client{
		Timeout: time.Second * 90,
	}
	ag.client = client
	return nil
}

// GetConfigFileContent get config file content from local
func (ag *AgentRegisterService) GetConfigFileContent(filePath string) ([]byte, error) {
	ag.logger.Debug("execute get config file content", "trace", "agent-os-instance.agent_register_service.GetConfigFileContent", "filePath", filePath)
	if err := ag.fileSystem.VerifyFileExist(filePath); err != nil {
		ag.logger.Error("error in verify file exist", "trace", "agent-os-instance.agent_register_service.GetConfigFileContent", "error", err.Error())
		return nil, err
	}
	content, err := ag.fileSystem.GetFileContent(filePath)
	if err != nil {
		ag.logger.Error("error in get file content", "trace", "agent-os-instance.agent_register_service.GetConfigFileContent", "error", err.Error())
		return nil, err
	}
	return content, nil
}

// InjectClientInfo execute injection the client info in metadata content
func (ag *AgentRegisterService) InjectClientInfo(configFileContent []byte, linuxMetadataContent []byte, isCreate bool) ([]byte, string, error) {
	ag.logger.Debug("execute inject client info", "trace", "agent-os-instance.agent_register_service.InjectClientInfo", "configFileContent", string(configFileContent), "linuxMetadataContent", string(linuxMetadataContent))
	var configAgentDto dto.ConfigAgent
	var metadata dto.Metadata
	var agentRegisterDataCreateOrUpdate dto.AgentRegisterDataCreateOrUpdate

	if err := ag.ymlClient.Unmarshall(configFileContent, &configAgentDto); err != nil {
		ag.logger.Error("error in unmarshaller", "trace", "agent-os-instance.agent_register_service.InjectClientInfo", "error", err.Error())
		return nil, "", err
	}

	if err := ag.json.Unmarshall(linuxMetadataContent, &metadata); err != nil {
		ag.logger.Error("error in unmarshaller", "trace", "agent-os-instance.agent_register_service.InjectClientInfo", "error", err.Error())
		return nil, "", err
	}

	slcTags, err := utils.TransformMapToSlice(configAgentDto.Agent.Tags)
	if err != nil {
		return nil, "", err
	}

	var vmName string
	if configAgentDto.VMName != "" {
		vmName = configAgentDto.VMName
		// vm_name from config is authoritative — override compute_name
		// so both fields reflect the configured name on first registration.
		metadata.ComputeInfo.Computename = configAgentDto.VMName
	} else {
		vmName = metadata.ComputeInfo.Computename
		// Persist the collected compute_name as vm_name in config.yml so
		// all three (config, VMName payload, compute_name) stay aligned.
		configAgentDto.VMName = vmName
		if configBytes, err := ag.ymlClient.Marshall(&configAgentDto); err != nil {
			ag.logger.Error("error marshalling config to save vm_name", "trace", "agent-os-instance.agent_register_service.InjectClientInfo", "error", err.Error())
		} else if err := ag.fileSystem.WriteFileContent(ag.configFilePath, configBytes); err != nil {
			ag.logger.Error("error writing vm_name to config file", "trace", "agent-os-instance.agent_register_service.InjectClientInfo", "error", err.Error())
		}
	}

	oryaId, err := utils.GetOrCreateOryaID()
	if err != nil {
		ag.logger.Error("error getting orya_id", "trace", "agent-os-instance.agent_register_service.InjectClientInfo", "error", err.Error())
	}
	metadata.ComputeInfo.OryaId = oryaId

	agentRegisterDataCreateOrUpdate = dto.AgentRegisterDataCreateOrUpdate{
		Tags:     slcTags,
		Metadata: metadata,
		VMName:   vmName,
	}
	if isCreate {
		agentRegisterDataCreateOrUpdate.NoGroupAssociation = configAgentDto.NoGroupAssociation
	}

	registerDataBytes, err := ag.json.Marshall(agentRegisterDataCreateOrUpdate)
	if err != nil {
		ag.logger.Error("error in marshaller", "trace", "agent-os-instance.agent_register_service.InjectClientInfo", "error", err.Error())
		return nil, "", err
	}
	var apiKeyOrToken string
	if isCreate {
		apiKeyOrToken = configAgentDto.Agent.ApiKey
	} else {
		apiKeyOrToken = configAgentDto.AccessToken
	}

	return registerDataBytes, apiKeyOrToken, nil
}

// SendMetadataCreate execute send metadata to create initial host in
// register service
func (ag *AgentRegisterService) SendMetadataCreate(data []byte) ([]byte, int, error) {
	ag.logger.Debug("execute send metadata create", "trace", "agent-os-instance.agent_register_service.SendMetadataCreate", "data", string(data))

	injectedMetadata, apiKey, err := ag.prepareToSendCreate(data)
	if err != nil {
		ag.logger.Error("error in prepare to send", "trace", "agent-os-instance.agent_register_service.SendMetadata", "error", err.Error())
		return nil, 0, err
	}

	urlMetadataCreate := fmt.Sprintf("%s/compute/v1/docp", ag.urlRegister)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlMetadataCreate, bytes.NewBuffer(injectedMetadata))
	if err != nil {
		ag.logger.Error("error in create request", "trace", "agent-os-instance.agent_register_service.SendMetadata", "error", err.Error())
		return nil, 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("docp-api-key", apiKey)

	res, err := ag.client.Do(req)
	if err != nil {
		ag.logger.Error("error in execute request", "trace", "agent-os-instance.agent_register_service.SendMetadata", "error", err.Error())
		return nil, 0, err
	}

	defer res.Body.Close()

	respBytes, err := io.ReadAll(res.Body)
	if err != nil {
		ag.logger.Error("error in read body response", "trace", "agent-os-instance.agent_register_service.SendMetadata", "error", err.Error())
		return nil, 0, err
	}

	return respBytes, res.StatusCode, nil
}

// SendMetadataUpdate execute send metadata to update host in
// register service
func (ag *AgentRegisterService) SendMetadataUpdate(data []byte) ([]byte, int, error) {
	ag.logger.Debug("execute send metadata update", "trace", "agent-os-instance.agent_register_service.SendMetadataUpdate", "data", string(data))

	injectedMetadata, token, err := ag.prepareToSendUpdate(data)
	if err != nil {
		ag.logger.Error("error in prepare to send", "trace", "agent-os-instance.agent_register_service.SendMetadata", "error", err.Error())
		return nil, 0, err
	}

	urlMetadataUpdate := fmt.Sprintf("%s/compute/v1/docp", ag.urlRegister)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, urlMetadataUpdate, bytes.NewBuffer(injectedMetadata))
	if err != nil {
		ag.logger.Error("error in create request", "trace", "agent-os-instance.agent_register_service.SendMetadata", "error", err.Error())
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	res, err := ag.client.Do(req)
	if err != nil {
		ag.logger.Error("error in execute request", "trace", "agent-os-instance.agent_register_service.SendMetadata", "error", err.Error())
		return nil, 0, err
	}
	defer res.Body.Close()
	respBytes, err := io.ReadAll(res.Body)
	if err != nil {
		ag.logger.Error("error in read body response", "trace", "agent-os-instance.agent_register_service.SendMetadata", "error", err.Error())
		return nil, 0, err
	}
	return respBytes, res.StatusCode, nil
}
