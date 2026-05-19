package interfaces

import "github.com/OryaHub/agent-os-instance/libs/dto"

type IDatadogOperation interface {
	Setup() error
	InstallAgent(ddSite, ddApiKey, version string) error
	InstallAgentApmSingleStep(ddSite string, ddApiKey string, ddApmInstrumentationLibraries string) error
	InstallAgentApmTracingLibrary(languageName, pathTracer, version string) error
	UninstallAgent() error
	DiscoverDatadogConfigPath() (string, error)
	DatadogAddPermitionGroupFilePath(filePath string) error
	DatadogAddPermitionUser() error
	BackupConfigFileDatadog(filePath string, content []byte) error
	UpdateConfigFileDatadog(filePath string) error
	UpdateConfigFileDatadogContent(filePath string, config dto.DatadogConfigDTO) error
	UpdateRepository() error
	GetVersion() (string, error)
	GetLatestVersion() (string, error)
	UpdateVersion(version string) error
	RollbackVersion(version string) error
	DPKGConfigure() error
	GetInfos() ([]byte, error)
	WriteEnvironmentFile(ddSite, ddApiKey string) error
}
