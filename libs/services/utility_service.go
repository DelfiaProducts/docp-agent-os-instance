package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/OryaHub/agent-os-instance/libs/dto"
	"github.com/OryaHub/agent-os-instance/libs/interfaces"
	"github.com/OryaHub/agent-os-instance/libs/utils"
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
	urlVersions, err := url.JoinPath(utils.GetBucketUrl(), utils.GetFileAgentVersionsName())
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

// ValidateUrlExists validates if a URL exists.
func (u *UtilityService) ValidateUrlExists(url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	res, err := u.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return utils.ErrValidateUrlExists()
	}

	return nil
}

// GetDatadogLastVersionFromGithub fetches the latest version of Datadog from GitHub.
func (u *UtilityService) GetDatadogLastVersionFromGithub() (string, error) {
	url := fmt.Sprintf("%s/latest", utils.GetDatadogGithubUrlVersions())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := u.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", utils.ErrFailedGetLatestVersionDatadog()
	}

	var datadogApiGithubVersions dto.DatadogApiGithubVersions
	if err := json.NewDecoder(res.Body).Decode(&datadogApiGithubVersions); err != nil {
		return "", err
	}

	return datadogApiGithubVersions.TagName, nil
}
