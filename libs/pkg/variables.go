package pkg

import "errors"

var (
	ErrNotAuthorized                 = errors.New("not authorized")
	ErrAuthTokenClaimsInvalid        = errors.New("invalid token claims")
	ErrNotFound                      = errors.New("not found")
	ErrSignalAlreadyExists           = errors.New("signal already exists")
	ErrFailedGetAgentVersions        = errors.New("failed get agent versions")
	ErrFailedGetLatestVersionDatadog = errors.New("failed get latest version datadog")
	ErrValidateUrlExists             = errors.New("url not found")
	ErrAgentVersionNotFound          = errors.New("agent version not found")
	ErrDatadogVersionNotFound        = errors.New("datadog version not found")
	ErrDatadogVersionInvalidFormat   = errors.New("datadog version invalid format")
	ErrContextExpired                = errors.New("context expired")
	ErrFailedCheckUsageLimit         = errors.New("failed check usage limit")

	ErrApiKeyInvalid      = errors.New("invalid api key or wrong orya site")
	ErrNetworkError       = errors.New("network error")
	ErrBackendError       = errors.New("orya backend error")
	ErrUsageLimitExceeded = errors.New("usage limit exceeded")

	// transactions events
	TransactionEventOpen   = "open"
	TransactionEventUpdate = "update"
	TransactionEventClose  = "close"
)
