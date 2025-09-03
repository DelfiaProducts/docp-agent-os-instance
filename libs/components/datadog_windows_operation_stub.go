//go:build !windows

package components

import (
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/dto"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
)

type DatadogWindowsOperation struct{}

func NewDatadogWindowsOperation(logger interfaces.ILogger) *DatadogWindowsOperation {
	return &DatadogWindowsOperation{}
}

// prepareEnvs return envs the datadog
func (d *DatadogWindowsOperation) prepareEnvs(ddSite, ddApiKey string) []string {
	var envs []string
	return envs
}

// sendStatus execute send status for state check
func (d *DatadogWindowsOperation) sendStatus(status, message string) error {
	return nil
}

// getApmEnvVarsSingleStep get envs apm datadog in mode single step
func (d *DatadogWindowsOperation) getApmEnvVarsSingleStep(envs []dto.DatadogEnvVars) (string, string, string) {
	return "", "", ""
}

func (d *DatadogWindowsOperation) Setup() error {
	return nil
}

// InstallAgent execute install the agent in linux
func (d *DatadogWindowsOperation) InstallAgent(ddSite, ddApiKey string) error {
	return nil
}

// InstallAgentApmSingleStep execute install the agent in linux with apm tracer on mode single step
func (d *DatadogWindowsOperation) InstallAgentApmSingleStep(ddSite string, ddApiKey string, datadogEnvVars []dto.DatadogEnvVars) error {
	return nil
}

// InstallAgentApmTracingLibrary execute install the agent in linux with apm tracer on mode tracing library
func (d *DatadogWindowsOperation) InstallAgentApmTracingLibrary(languageName, pathTracer, version string) error {
	return nil
}

// UninstallAgent execute uninstall the agent in linux
func (d *DatadogWindowsOperation) UninstallAgent() error {
	return nil
}

// DiscoverDatadogConfigPath return file path the datadog config
func (d *DatadogWindowsOperation) DiscoverDatadogConfigPath() (string, error) {
	return "", nil
}

// DatadogAddPermitionGroupFilePath add permition for file path the datadog
func (d *DatadogWindowsOperation) DatadogAddPermitionGroupFilePath(filePath string) error {
	return nil
}

// DatadogAddPermitionUser add permition for directory the datadog
func (d *DatadogWindowsOperation) DatadogAddPermitionUser() error {
	return nil
}

// BackupConfigFileDatadog execute backup the current config file datadog
func (d *DatadogWindowsOperation) BackupConfigFileDatadog(filePath string, content []byte) error {
	return nil
}

// UpdateConfigFileDatadog execute update the config file datadog
func (d *DatadogWindowsOperation) UpdateConfigFileDatadog(filePath string) error {
	return nil
}

// UpdateRepository execute update repository local
func (d *DatadogWindowsOperation) UpdateRepository() error {
	return nil
}

// GetVersion return the version of the datadog agent
func (d *DatadogWindowsOperation) GetVersion() (string, error) {
	return "", nil
}

// DPKGConfigure execute configure dpkg
func (d *DatadogWindowsOperation) DPKGConfigure() error {
	// TODO: not implemented windows
	return nil
}
