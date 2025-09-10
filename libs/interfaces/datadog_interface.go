package interfaces

import (
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/dto"
)

type IDatadogOperation interface {
	Setup() error
	InstallAgent(ddSite, ddApiKey string) error
	InstallAgentApmSingleStep(ddSite string, ddApiKey string, datadogEnvVars []dto.DatadogEnvVars) error
	InstallAgentApmTracingLibrary(languageName, pathTracer, version string) error
	UninstallAgent() error
	DiscoverDatadogConfigPath() (string, error)
	DatadogAddPermitionGroupFilePath(filePath string) error
	DatadogAddPermitionUser() error
	BackupConfigFileDatadog(filePath string, content []byte) error
	UpdateConfigFileDatadog(filePath string) error
	UpdateRepository() error
	GetVersion() (string, error)
	GetLatestVersion() (string, error)
	UpdateVersion(version string) error
	RollbackVersion(version string) error
	DPKGConfigure() error
}
