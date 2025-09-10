package utils

import (
	"runtime"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/pkg"
)

// ChoiceNameService return name os service
func ChoiceNameService(serviceName string) string {
	name := ""
	switch serviceName {
	case "agent":
		name = "docp-agent.service"
	case "manager":
		name = "docp-manager.service"
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
		name = "DocpAgent"
	case "manager":
		name = "DocpManager"
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
		name = "docp-agent/agent"
	case "manager":
		name = "docp-agent/manager"
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
		name = "com.docp.agent"
	case "manager":
		name = "com.docp.manager"
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
	return pkg.DOCP_FILE_AGENT_VERSIONS_NAME
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
