package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/OryaHub/agent-os-instance/libs/dto"
	"github.com/OryaHub/agent-os-instance/libs/interfaces"
	"github.com/OryaHub/agent-os-instance/libs/pkg"
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

// ValidateOryaSite checks if the Orya Site is reachable (DNS + connectivity).
// Scenario 3: non-existent domain / network -> returns ErrNetworkError
func (u *UtilityService) ValidateOryaSite(oryaSite string) error {
	u.logger.Debug("checking access to Orya Site", "url", oryaSite)

	parsedURL, err := url.Parse(oryaSite)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// 1. DNS validation
	hostname := parsedURL.Hostname()
	u.logger.Debug("resolving DNS", "hostname", hostname)
	_, err = net.LookupHost(hostname)
	if err != nil {
		u.logger.Error("DNS resolution failed", "hostname", hostname, "error", err.Error())
		return fmt.Errorf("%w: domain not found - %s", pkg.ErrNetworkError, hostname)
	}

	// 2. HTTP connectivity validation
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, oryaSite, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	res, err := u.client.Do(req)
	if err != nil {
		u.logger.Error("connection to Orya Site failed", "error", err.Error())
		return fmt.Errorf("%w: could not connect to %s", pkg.ErrNetworkError, hostname)
	}
	res.Body.Close()

	u.logger.Info("Orya Site reachable")
	return nil
}

// errorDetail represents the error body returned by the API
type errorDetail struct {
	Detail struct {
		ErrorID string `json:"error_id"`
		Message string `json:"message"`
	} `json:"detail"`
}

// FetchAgentVersions fetches the summary agent versions from the repository.
func (u *UtilityService) FetchAgentVersions() (dto.AgentVersions, error) {
	urlVersions, err := url.JoinPath(utils.GetBucketUrl(), utils.GetFileAgentVersionsName())
	if err != nil {
		return dto.AgentVersions{}, err
	}
	u.logger.Debug("fetching agent versions", "urlVersions", urlVersions)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlVersions, nil)
	if err != nil {
		u.logger.Error("error fetching agent versions", "error", err.Error())
		return dto.AgentVersions{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := u.client.Do(req)
	if err != nil {
		u.logger.Error("error executing request", "error", err.Error())
		return dto.AgentVersions{}, err
	}
	defer res.Body.Close()

	u.logger.Debug("version fetch response", "statusCode", res.StatusCode)
	if res.StatusCode != http.StatusOK {
		return dto.AgentVersions{}, utils.ErrFailedGetAgentVersions()
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return dto.AgentVersions{}, err
	}
	u.logger.Debug("version response body", "body", string(body))

	var versions dto.AgentVersions
	if err := json.Unmarshal(body, &versions); err != nil {
		return dto.AgentVersions{}, err
	}

	return versions, nil
}

// ValidateUsageLimit checks if the usage limit for installing the agent has been exceeded.
// Scenario 1: invalid API Key -> returns ErrApiKeyInvalid (identified by HTTP 404/401/403 body)
// Scenario 2: limit exceeded -> returns ErrUsageLimitExceeded (has_limit=false)
// Scenario 3: network/DNS -> returns ErrNetworkError
// Scenario 4: backend error -> returns ErrBackendError
func (u *UtilityService) ValidateUsageLimit(apiKey, oryaSite string) (dto.UsageLimitResponse, error) {
	url := fmt.Sprintf("%s/usage-tracking/usage-limit/orya/check", oryaSite)
	u.logger.Debug("checking usage limit", "url", url)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return dto.UsageLimitResponse{}, err
	}
	req.Header.Set("docp-api-key", apiKey)
	res, err := u.client.Do(req)
	if err != nil {
		u.logger.Error("network error checking usage limit", "error", err.Error())
		return dto.UsageLimitResponse{}, fmt.Errorf("%w: %s", pkg.ErrNetworkError, err.Error())
	}
	defer res.Body.Close()

	// Read body even on error to analyze the cause
	bodyBytes, _ := io.ReadAll(res.Body)

	// Scenario 4: internal server error
	if res.StatusCode >= http.StatusInternalServerError {
		return dto.UsageLimitResponse{}, pkg.ErrBackendError
	}

	// Scenario 1: invalid API Key — API returns 404/401/403 with {"detail":{"message":"Api key ... not found"}}
	if res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		// Try to extract error message from body
		var errResp errorDetail
		if json.Unmarshal(bodyBytes, &errResp) == nil && strings.Contains(errResp.Detail.Message, "not found") {
			return dto.UsageLimitResponse{}, pkg.ErrApiKeyInvalid
		}
		return dto.UsageLimitResponse{}, pkg.ErrApiKeyInvalid
	}

	if res.StatusCode != http.StatusOK {
		return dto.UsageLimitResponse{}, fmt.Errorf("%w: unexpected HTTP %d", pkg.ErrBackendError, res.StatusCode)
	}

	var usageLimitResponse dto.UsageLimitResponse
	if err := json.Unmarshal(bodyBytes, &usageLimitResponse); err != nil {
		return dto.UsageLimitResponse{}, err
	}

	if !usageLimitResponse.HasLimit {
		return usageLimitResponse, pkg.ErrUsageLimitExceeded
	}

	return usageLimitResponse, nil
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
