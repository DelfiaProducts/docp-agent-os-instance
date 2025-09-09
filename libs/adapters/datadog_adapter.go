package adapters

import (
	"strings"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/components"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/dto"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/pkg"
)

// DatadogAdapter is struct for adapter the datadog
type DatadogAdapter struct {
	base64Client     *pkg.Base64Client
	datadogOperation interfaces.IDatadogOperation
	logger           interfaces.ILogger
	osOperation      interfaces.IOSOperation
}

// NewDatadogAdapter return instance of datadog adapter
func NewDatadogAdapter(logger interfaces.ILogger) *DatadogAdapter {
	return &DatadogAdapter{
		logger: logger,
	}
}

// Setup execute configuration the adapter
func (d *DatadogAdapter) Setup() error {
	osOperation, err := components.SystemOperation(d.logger)
	if err != nil {
		return err
	}
	if err := osOperation.Setup(); err != nil {
		return err
	}
	datadogOperation, err := components.DatadogOperation(d.logger)
	if err != nil {
		return err
	}
	if err := datadogOperation.Setup(); err != nil {
		return err
	}
	d.datadogOperation = datadogOperation
	base64Client := pkg.NewBase64Client()
	d.base64Client = base64Client
	return nil
}

// IsActive execute install the agent in linux
func (d *DatadogAdapter) IsActive() (bool, error) {
	d.logger.Debug("is active", "trace", "docp-agent-os-instance.datadog_linux_adapter.IsActive")
	output, err := d.osOperation.Status("datadog")
	if err != nil {
		return false, err
	}
	output = strings.ReplaceAll(output, "\"", "")
	if output == "active" {
		return true, nil
	}
	return false, nil
}

// InstallAgent execute install the agent in linux
func (d *DatadogAdapter) InstallAgent(ddSite, ddApiKey string) error {
	d.logger.Debug("install agent", "trace", "docp-agent-os-instance.datadog_linux_adapter.InstallAgent", "ddSite", ddSite, "ddApiKey", ddApiKey)
	if err := d.datadogOperation.InstallAgent(ddSite, ddApiKey); err != nil {
		return err
	}

	return nil
}

// GetApmEnvVarsTracingLibrary get envs apm datadog in mode tracing library
func (d *DatadogAdapter) GetApmEnvVarsTracingLibrary(envs []dto.DatadogEnvVars) (string, string, string) {
	d.logger.Debug("get apm envs vars tracing library", "trace", "docp-agent-os-instance.datadog_linux_adapter.GetApmEnvVarsTracingLibrary", "envs", envs)
	var language, pathTracer, version string
	for _, env := range envs {
		if env.Name == "language" {
			language = env.Value
		}
		if env.Name == "path_tracer" {
			pathTracer = env.Value
		}
		if env.Name == "version" {
			version = env.Value
		}
	}
	return language, pathTracer, version
}

// InstallAgentApmSingleStep execute install the agent in linux with apm tracer on mode single step
func (d *DatadogAdapter) InstallAgentApmSingleStep(ddSite string, ddApiKey string, datadogEnvVars []dto.DatadogEnvVars) error {
	d.logger.Debug("install agent apm single step", "trace", "docp-agent-os-instance.datadog_linux_adapter.InstallAgentApmSingleStep", "ddSite", ddSite, "ddApiKey", ddApiKey, "datadogEnvVars", datadogEnvVars)
	if err := d.datadogOperation.InstallAgentApmSingleStep(ddSite, ddApiKey, datadogEnvVars); err != nil {
		return err
	}
	return nil
}

// InstallAgentApmTracingLibrary execute install the agent in linux with apm tracer on mode tracing library
func (d *DatadogAdapter) InstallAgentApmTracingLibrary(languageName, pathTracer, version string) error {
	d.logger.Debug("install agent apm tracing library", "trace", "docp-agent-os-instance.datadog_linux_adapter.InstallAgentApmTracingLibrary", "languageName", languageName, "pathTracer", pathTracer, "version", version)
	if err := d.datadogOperation.InstallAgentApmTracingLibrary(languageName, pathTracer, version); err != nil {
		return err
	}

	return nil
}

// UninstallAgent execute uninstall the agent in linux
func (d *DatadogAdapter) UninstallAgent() error {
	d.logger.Debug("uninstall agent", "trace", "docp-agent-os-instance.datadog_linux_adapter.UninstallAgent")
	if err := d.datadogOperation.UninstallAgent(); err != nil {
		return err
	}
	return nil
}

// DiscoverDatadogConfigPath return file path the datadog config
func (d *DatadogAdapter) DiscoverDatadogConfigPath() (string, error) {
	d.logger.Debug("discover datadog config path", "trace", "docp-agent-os-instance.datadog_linux_adapter.DiscoverDatadogConfigPath")
	datadogPath, err := d.datadogOperation.DiscoverDatadogConfigPath()
	return datadogPath, err
}

// DecodeBase64 decode base64 content
func (d *DatadogAdapter) DecodeBase64(base64Content string) ([]byte, error) {
	d.logger.Debug("decode base64", "trace", "docp-agent-os-instance.datadog_linux_adapter.DecodeBase64", "base64Content", base64Content)
	content, err := d.base64Client.Decode(base64Content)
	if err != nil {
		return nil, err
	}
	return content, nil
}

// DatadogAddPermitionGroupFilePath add permition for file path the datadog
func (d *DatadogAdapter) DatadogAddPermitionGroupFilePath(filePath string) error {
	d.logger.Debug("datadog add permition group file path", "trace", "docp-agent-os-instance.datadog_linux_adapter.DatadogAddPermitionGroupFilePath")
	if err := d.datadogOperation.DatadogAddPermitionGroupFilePath(filePath); err != nil {
		return err
	}
	return nil
}

// DatadogAddPermitionUser add permition for directory the datadog
func (d *DatadogAdapter) DatadogAddPermitionUser() error {
	d.logger.Debug("datadog add permition user", "trace", "docp-agent-os-instance.datadog_linux_adapter.DatadogAddPermitionUser")
	if err := d.datadogOperation.DatadogAddPermitionUser(); err != nil {
		return err
	}
	return nil
}

// BackupConfigFileDatadog execute backup the current config file datadog
func (d *DatadogAdapter) BackupConfigFileDatadog(filePath string, content []byte) error {
	d.logger.Debug("backup config file datadog", "trace", "docp-agent-os-instance.datadog_linux_adapter.BackupConfigFileDatadog", "filePath", filePath, "content", string(content))
	if err := d.datadogOperation.BackupConfigFileDatadog(filePath, content); err != nil {
		return err
	}
	return nil
}

// UpdateConfigFileDatadog execute update the config file datadog
func (d *DatadogAdapter) UpdateConfigFileDatadog(filePath string) error {
	d.logger.Debug("update config file datadog", "trace", "docp-agent-os-instance.datadog_linux_adapter.UpdateConfigFileDatadog", "filePath", filePath)
	if err := d.datadogOperation.UpdateConfigFileDatadog(filePath); err != nil {
		return err
	}
	return nil
}

// UpdateRepository execute update repository local
func (d *DatadogAdapter) UpdateRepository() error {
	d.logger.Debug("update repository", "trace", "docp-agent-os-instance.datadog_linux_adapter.UpdateRepository")
	if err := d.datadogOperation.UpdateRepository(); err != nil {
		return err
	}
	return nil
}

// GetVersion return version installed datadog
func (d *DatadogAdapter) GetVersion() (string, error) {
	d.logger.Debug("get version", "trace", "docp-agent-os-instance.datadog_linux_adapter.GetVersion")
	version, err := d.datadogOperation.GetVersion()
	if err != nil {
		return "", err
	}
	return version, nil
}

// GetLatestVersion return latest version datadog
func (d *DatadogAdapter) GetLatestVersion() (string, error) {
	d.logger.Debug("get latest version", "trace", "docp-agent-os-instance.datadog_linux_adapter.GetLatestVersion")
	version, err := d.datadogOperation.GetLatestVersion()
	if err != nil {
		return "", err
	}
	return version, nil
}

// UpdateVersion execute update the version of the datadog agent
func (d *DatadogAdapter) UpdateVersion(version string) error {
	d.logger.Debug("update version", "trace", "docp-agent-os-instance.datadog_linux_adapter.UpdateVersion", "version", version)
	if err := d.datadogOperation.UpdateVersion(version); err != nil {
		return err
	}
	return nil
}

// RollbackVersion execute rollback the version of the datadog agent
func (d *DatadogAdapter) RollbackVersion(version string) error {
	d.logger.Debug("rollback version", "trace", "docp-agent-os-instance.datadog_linux_adapter.RollbackVersion", "version", version)
	if err := d.datadogOperation.RollbackVersion(version); err != nil {
		return err
	}
	return nil
}

// DPKGConfigure execute configure dpkg
func (d *DatadogAdapter) DPKGConfigure() error {
	d.logger.Debug("dpkg configure", "trace", "docp-agent-os-instance.datadog_linux_adapter.DPKGConfigure")
	if err := d.datadogOperation.DPKGConfigure(); err != nil {
		return err
	}
	return nil
}
