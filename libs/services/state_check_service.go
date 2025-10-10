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

// StateCheckService is struct for state check service
type StateCheckService struct {
	logger        interfaces.ILogger
	stateCheckUrl string
	workDirPath   string
	fileSystem    *pkg.FileSystem
	hostStats     *pkg.HostStats
	ymlClient     *pkg.YmlClient
	client        *http.Client
	json          *pkg.JsonClient
}

// NewStateCheckService return instance of state check service
func NewStateCheckService(logger interfaces.ILogger) *StateCheckService {
	return &StateCheckService{
		logger: logger,
		json:   pkg.NewJsonClient(),
	}
}

// Setup configure state check
func (s *StateCheckService) Setup() error {
	client := &http.Client{
		Timeout: time.Second * 90,
	}
	s.client = client
	domainUrl, err := utils.GetDomainUrl()
	if err != nil {
		return err
	}
	s.stateCheckUrl = domainUrl
	workDirPath, err := utils.GetWorkDirPath()
	if err != nil {
		return err
	}
	s.workDirPath = workDirPath
	fileSystem := pkg.NewFileSystem()
	s.fileSystem = fileSystem
	hostStatus := pkg.NewHostStats()
	s.hostStats = hostStatus
	ymlClient := pkg.NewYmlClient()
	s.ymlClient = ymlClient
	return nil
}

// PreparePayload execute prepare the payload
func (s *StateCheckService) PreparePayload() ([]byte, string, error) {
	s.logger.Debug("prepare payload", "trace", "agent-os-instance.state_check_service.PreparePayload")
	content, err := s.getContentConfigFile()
	if err != nil {
		return nil, "", err
	}
	var config dto.ConfigAgent
	if err := s.ymlClient.Unmarshall(content, &config); err != nil {
		return nil, "", err
	}

	return nil, config.AccessToken, nil
}

// PreparePayloadStatus execute prepare the payload for status
func (s *StateCheckService) PreparePayloadStatus(transaction dto.TransactionStatus) ([]byte, string, error) {
	s.logger.Debug("prepare payload status", "trace", "agent-os-instance.state_check_service.PreparePayloadStatus", "transaction", transaction)
	content, err := s.getContentConfigFile()
	if err != nil {
		return nil, "", err
	}
	var config dto.ConfigAgent
	if err := s.ymlClient.Unmarshall(content, &config); err != nil {
		return nil, "", err
	}

	bStateCheckPayloadStatus, err := s.json.Marshall(&transaction)
	if err != nil {
		return nil, "", err
	}

	return bStateCheckPayloadStatus, config.AccessToken, nil
}

// GetState return state from state check api
func (s *StateCheckService) GetState() ([]byte, int, error) {
	s.logger.Debug("execute get state", "trace", "agent-os-instance.state_check_service.GetState")
	_, accessToken, err := s.PreparePayload()
	if err != nil {
		s.logger.Error("error in prepare payload", "trace", "agent-os-instance.state_check_service.GetState", "error", err.Error())
		return nil, 0, err
	}
	urlStateCheck := fmt.Sprintf("%s/compute/v1/status/info", s.stateCheckUrl)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStateCheck, nil)
	if err != nil {
		s.logger.Error("error in create request", "trace", "agent-os-instance.state_check_service.GetState", "error", err.Error())
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	res, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("error in execute request", "trace", "agent-os-instance.state_check_service.GetState", "error", err.Error())
		return nil, 0, err
	}
	defer res.Body.Close()
	respBytes, err := io.ReadAll(res.Body)
	if err != nil {
		s.logger.Error("error in read body response", "trace", "agent-os-instance.state_check_service.GetState", "error", err.Error())
		return nil, 0, err
	}
	return respBytes, res.StatusCode, nil
}

// SendStatus execute send status for state check api
func (s *StateCheckService) SendStatus(transaction dto.TransactionStatus) ([]byte, int, error) {
	s.logger.Debug("execute send status state check", "trace", "agent-os-instance.state_check_service.SendStatus", "transaction", transaction)
	payload, acessToken, err := s.PreparePayloadStatus(transaction)
	if err != nil {
		s.logger.Error("error in prepare payload", "trace", "agent-os-instance.state_check_service.SendStatus", "error", err.Error())
		return nil, 0, err
	}
	urlStateCheckStatus := fmt.Sprintf("%s/compute/transaction", s.stateCheckUrl)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStateCheckStatus, bytes.NewBuffer(payload))
	if err != nil {
		s.logger.Error("error in create request", "trace", "agent-os-instance.state_check_service.SendStatus", "error", err.Error())
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", acessToken))
	res, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("error in execute request", "trace", "agent-os-instance.state_check_service.SendStatus", "error", err.Error())
		return nil, 0, err
	}
	defer res.Body.Close()
	respBytes, err := io.ReadAll(res.Body)
	if err != nil {
		s.logger.Error("error in read body response", "trace", "agent-os-instance.state_check_service.SendStatus", "error", err.Error())
		return nil, 0, err
	}
	return respBytes, res.StatusCode, nil
}
