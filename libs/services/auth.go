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

// AuthService is struct for auth service
type AuthService struct {
	urlAuth string
	logger  interfaces.ILogger
	client  *http.Client
	json    *pkg.JsonClient
}

// NewAuthService return instance the auth service
func NewAuthService(logger interfaces.ILogger) *AuthService {
	return &AuthService{
		logger: logger,
		json:   pkg.NewJsonClient(),
	}
}

// Setup execute configurations for auth service
func (as *AuthService) Setup() error {
	urlDomain, err := utils.GetDomainUrl()
	if err != nil {
		return err
	}
	as.urlAuth = urlDomain
	client := &http.Client{
		Timeout: time.Second * 90,
	}
	as.client = client
	return nil
}

// AuthCall execute call for auth service
func (as *AuthService) AuthCall(payload dto.AuthPayload) ([]byte, int, error) {
	url := fmt.Sprintf("%s/agents/auth/api_key/token", as.urlAuth)
	apiKey := payload.ApiKey

	payloadBytes, err := as.json.Marshall(&payload)
	if err != nil {
		return nil, 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("docp-api-key", apiKey)

	res, err := as.client.Do(req)
	if err != nil {
		return nil, 0, err
	}

	defer res.Body.Close()

	respBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, 0, err
	}

	return respBytes, res.StatusCode, nil
}
