package utils

import (
	"runtime"

	"github.com/OryaHub/agent-os-instance/libs/pkg"
)

// ChoiceNameService return name os service
func ChoiceNameService(serviceName string) string {
	name := ""
	switch serviceName {
	case "agent":
		name = "orya-agent.service"
	case "manager":
		name = "orya-manager.service"
	case "datadog":
		name = "datadog-agent.service"
	}
	return name
}

// ChoiceNameServiceWindows return name windows os service
func ChoiceNameServiceWindows(serviceName string) string {
	name := ""
	switch serviceName {
	case "agent":
		name = "OryaAgent"
	case "manager":
		name = "OryaManager"
	case "datadog":
		name = "DatadogAgent"
	}
	return name
}

// GetNameForProcess return name for process
func GetNameForProcess(serviceName string) string {
	name := ""
	switch serviceName {
	case "agent":
		name = "orya-agent/agent"
	case "manager":
		name = "orya-agent/manager"
	case "datadog":
		name = "datadog-agent"
	}
	return name
}

// GetNameForLaunchd return name for launchd
func GetNameForLaunchd(serviceName string) string {
	name := ""
	switch serviceName {
	case "agent":
		name = "com.orya.agent"
	case "manager":
		name = "com.orya.manager"
	case "datadog":
		name = "com.datadoghq.agent"
	}
	return name
}

// RemoveItemFromSlice remove the item from slice
func RemoveItemFromSlice(items []string, item string) []string {
	newSlc := []string{}
	for _, i := range items {
		if i != item {
			newSlc = append(newSlc, i)
		}
	}
	return newSlc
}

// GetBinariesRepositoryUrl return url the repository
func GetBinariesRepositoryUrl() string {
	return pkg.URL_RELEASE
}

// GetBucketUrl return the bucket url
func GetBucketUrl() string {
	return pkg.BUCKET_URL
}

// GetFileAgentVersionsName return the file name for agent versions
func GetFileAgentVersionsName() string {
	return pkg.ORYA_FILE_AGENT_VERSIONS_NAME
}

// GetDatadogAgentUrlWindows return the url for download datadog agent windows
func GetDatadogAgentUrlWindows() string {
	return pkg.URL_DATADOG_AGENT_WINDOWS
}

// GetDatadogGithubUrlVersions return the url for datadog github versions
func GetDatadogGithubUrlVersions() string {
	return pkg.URL_DATADOG_GITHUB_VERSIONS
}

// GetRuntimeArch return runtime arch
func GetRuntimeArch() string {
	return runtime.GOARCH
}

// GetOSSystem return the operating system
func GetOSSystem() string {
	return runtime.GOOS
}

// ErrAuthTokenClaimsInvalid return error the invalid claims token
func ErrAuthTokenClaimsInvalid() error {
	return pkg.ErrAuthTokenClaimsInvalid
}

// ErrSignalAlreadyExists return error the signal already exists
func ErrSignalAlreadyExists() error {
	return pkg.ErrSignalAlreadyExists
}

// ErrFailedGetAgentVersions return error the failed get agent versions
func ErrFailedGetAgentVersions() error {
	return pkg.ErrFailedGetAgentVersions
}

// ErrAgentVersionNotFound return error the agent version not found
func ErrAgentVersionNotFound() error {
	return pkg.ErrAgentVersionNotFound
}

// ErrDatadogVersionNotFound return error the datadog version not found
func ErrDatadogVersionNotFound() error {
	return pkg.ErrDatadogVersionNotFound
}

// ErrContextExpired return error the context expired
func ErrContextExpired() error {
	return pkg.ErrContextExpired
}

// ErrValidateUrlExists return error the url not found
func ErrValidateUrlExists() error {
	return pkg.ErrValidateUrlExists
}

// ErrFailedGetLatestVersionDatadog return error the failed get latest version datadog
func ErrFailedGetLatestVersionDatadog() error {
	return pkg.ErrFailedGetLatestVersionDatadog
}

// ErrDatadogVersionInvalidFormat return error the datadog version invalid format
func ErrDatadogVersionInvalidFormat() error {
	return pkg.ErrDatadogVersionInvalidFormat
}

// ErrFailedCheckUsageLimit return error the failed check usage limit
func ErrFailedCheckUsageLimit() error {
	return pkg.ErrFailedCheckUsageLimit
}

// ErrApiKeyInvalid return error for invalid API Key or wrong Orya Site
func ErrApiKeyInvalid() error {
	return pkg.ErrApiKeyInvalid
}

// ErrNetworkError return error for network failure
func ErrNetworkError() error {
	return pkg.ErrNetworkError
}

// ErrBackendError return error for backend server error
func ErrBackendError() error {
	return pkg.ErrBackendError
}

// ErrUsageLimitExceeded return error for exceeded usage limit
func ErrUsageLimitExceeded() error {
	return pkg.ErrUsageLimitExceeded
}
