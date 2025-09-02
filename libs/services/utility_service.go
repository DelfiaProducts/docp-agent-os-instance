package services

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/dto"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

// UtilityService provides utility functions for the application.
type UtilityService struct {
	client *http.Client
	logger interfaces.ILogger
}

// NewUtilityService creates a new instance of UtilityService.
func NewUtilityService(logger interfaces.ILogger) *UtilityService {
	return &UtilityService{
		logger: logger,
	}
}

// Setup configure the utility service.
func (u *UtilityService) Setup() error {
	client := &http.Client{
		Timeout: time.Second * 90,
	}
	u.client = client

	return nil
}

// FetchAgentVersions fetches the summary agent versions from the repository.
func (u *UtilityService) FetchAgentVersions() (dto.AgentVersions, error) {
	urlVersions, err := url.JoinPath(utils.GetBinariesRepositoryUrl(), utils.GetFileAgentVersionsName())
	if err != nil {
		return dto.AgentVersions{}, err
	}
	u.logger.Debug("fetch summary agent versions", "urlVersions", urlVersions)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlVersions, nil)
	if err != nil {
		u.logger.Error("error fetch summary agent versions", "error", err.Error())
		return dto.AgentVersions{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := u.client.Do(req)
	if err != nil {
		u.logger.Error("error in execute request", "error", err.Error())
		return dto.AgentVersions{}, err
	}
	defer res.Body.Close()

	u.logger.Debug("fetch summary agent versions", "statusCode", res.StatusCode)
	if res.StatusCode != http.StatusOK {
		return dto.AgentVersions{}, utils.ErrFailedGetAgentVersions()
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return dto.AgentVersions{}, err
	}
	u.logger.Debug("fetch summary agent versions", "body", string(body))

	var versions dto.AgentVersions
	if err := json.Unmarshal(body, &versions); err != nil {
		return dto.AgentVersions{}, err
	}

	return versions, nil
}
