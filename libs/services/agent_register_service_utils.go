package services

// prepareToSend exeucte prepare for send data to register service
func (ag *AgentRegisterService) prepareToSendCreate(metadata []byte) ([]byte, string, error) {
	ag.logger.Debug("execute prepare to send", "trace", "agent-os-instance.agent_register_service.prepareToSendCreate", "metadata", string(metadata))
	configFileBytes, err := ag.GetConfigFileContent(ag.configFilePath)
	if err != nil {
		ag.logger.Error("error in prepare to send", "trace", "agent-os-instance.agent_register_service.prepareToSend", "error", err.Error())
		return nil, "", err
	}
	injectedMetadataBytes, apiKey, err := ag.InjectClientInfoCreate(configFileBytes, metadata)
	if err != nil {
		ag.logger.Error("error in prepare to send", "trace", "agent-os-instance.agent_register_service.prepareToSend", "error", err.Error())
		return nil, "", err
	}
	return injectedMetadataBytes, apiKey, nil
}

// prepareToSend exeucte prepare for send data to register service
func (ag *AgentRegisterService) prepareToSendUpdate(metadata []byte) ([]byte, string, error) {
	ag.logger.Debug("execute prepare to send", "trace", "agent-os-instance.agent_register_service.prepareToSendUpdate", "metadata", string(metadata))
	configFileBytes, err := ag.GetConfigFileContent(ag.configFilePath)
	if err != nil {
		ag.logger.Error("error in prepare to send", "trace", "agent-os-instance.agent_register_service.prepareToSend", "error", err.Error())
		return nil, "", err
	}
	injectedMetadataBytes, apiKey, err := ag.InjectClientInfoUpdate(configFileBytes, metadata)
	if err != nil {
		ag.logger.Error("error in prepare to send", "trace", "agent-os-instance.agent_register_service.prepareToSend", "error", err.Error())
		return nil, "", err
	}
	return injectedMetadataBytes, apiKey, nil
}
