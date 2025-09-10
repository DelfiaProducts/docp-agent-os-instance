package pkg

import "errors"

var (
	ErrNotAuthorized          = errors.New("not authorized")
	ErrAuthTokenClaimsInvalid = errors.New("invalid token claims")
	ErrNotFound               = errors.New("not found")
	ErrSignalAlreadyExists    = errors.New("signal already exists")
	ErrFailedGetAgentVersions = errors.New("failed get agent versions")
	ErrAgentVersionNotFound   = errors.New("agent version not found")
	ErrDatadogVersionNotFound = errors.New("datadog version not found")
	ErrContextExpired         = errors.New("context expired")

	// transactions events
	TransactionEventOpen   = "open"
	TransactionEventUpdate = "update"
	TransactionEventClose  = "close"
)
