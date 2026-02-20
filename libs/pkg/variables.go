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

	// transactions events
	TransactionEventOpen   = "open"
	TransactionEventUpdate = "update"
	TransactionEventClose  = "close"
)
