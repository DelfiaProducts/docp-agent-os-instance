//go:build windows

package components

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/dto"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/pkg"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/services"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	URL_DATADOG_AGENT = "https://s3.amazonaws.com/ddagent-windows-stable/datadog-agent-7-latest.amd64.msi"
)

type DatadogWindowsOperation struct {
	logger           interfaces.ILogger
	program          *pkg.ExecProgram
	hostStats        *pkg.HostStats
	stateCheck       *services.StateCheckService
	fileSystem       *pkg.FileSystem
	datadogApmTracer *DatadogWindowsAPMTracer
}

func NewDatadogWindowsOperation(logger interfaces.ILogger) *DatadogWindowsOperation {
	return &DatadogWindowsOperation{
		logger: logger,
	}
}

// prepareEnvs return envs the datadog
func (d *DatadogWindowsOperation) prepareEnvs(ddSite, ddApiKey string) []string {
	var envs []string
	envs = append(envs, fmt.Sprintf("DD_API_KEY=%s", ddApiKey))
	envs = append(envs, fmt.Sprintf("DD_SITE=%s", ddSite))
	return envs
}

// getApmEnvVarsSingleStep get envs apm datadog in mode single step
func (d *DatadogWindowsOperation) getApmEnvVarsSingleStep(envs []dto.DatadogEnvVars) (string, string, string) {
	var ddApmInstrumentationEnabled, ddEnv, ddApmInstrumentationLibraries string
	for _, env := range envs {
		if env.Name == "DD_APM_INSTRUMENTATION_ENABLED" {
			ddApmInstrumentationEnabled = env.Value
		}
		if env.Name == "DD_ENV" {
			ddEnv = env.Value
		}
		if env.Name == "DD_APM_INSTRUMENTATION_LIBRARIES" {
			ddApmInstrumentationLibraries = env.Value
		}
	}
	return ddApmInstrumentationEnabled, ddEnv, ddApmInstrumentationLibraries
}

func (d *DatadogWindowsOperation) Setup() error {
	execProgram := pkg.NewExecProgram()
	d.program = execProgram
	hostStats := pkg.NewHostStats()
	d.hostStats = hostStats
	stateCheck := services.NewStateCheckService(d.logger)
	if err := stateCheck.Setup(); err != nil {
		return err
	}
	d.stateCheck = stateCheck
	datadogWindowsApmTracer := NewDatadogWindowsAPMTracer()
	d.datadogApmTracer = datadogWindowsApmTracer
	fileSystem := pkg.NewFileSystem()
	d.fileSystem = fileSystem
	return nil
}

// InstallAgent execute install the agent in linux
func (d *DatadogWindowsOperation) InstallAgent(ddSite, ddApiKey string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService("DatadogAgent")
	if err != nil {
		d.logger.Error("error in install datadog agent", "error", err)
		if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			command := fmt.Sprintf(`Start-Process -Wait msiexec -ArgumentList '/qn /i %s APIKEY="%s" SITE="%s"'`, URL_DATADOG_AGENT, ddApiKey, ddSite)
			out, err := d.program.ExecuteWithOutput("powershell", []string{}, "-Command", command)
			if err != nil {
				d.logger.Error("error in install datadog agent start process", "error", err)
				return err
			}
			d.logger.Debug("install agent datadog", "output", out)
			return nil
		}
		return err
	}
	defer s.Close()
	return nil
}

// InstallAgentApmSingleStep execute install the agent in linux with apm tracer on mode single step
func (d *DatadogWindowsOperation) InstallAgentApmSingleStep(ddSite string, ddApiKey string, datadogEnvVars []dto.DatadogEnvVars) error {
	d.logger.Debug("install agent apm single step", "trace", "docp-agent-os-instance.datadog_windows_operations.InstallAgentApmSingleStep")
	return nil
}

// InstallAgentApmTracingLibrary execute install the agent in linux with apm tracer on mode tracing library
func (d *DatadogWindowsOperation) InstallAgentApmTracingLibrary(languageName, pathTracer, version string) error {
	if err := d.datadogApmTracer.InstallLibrary(languageName, pathTracer, version); err != nil {
		return err
	}

	return nil
}

// UninstallAgent execute uninstall the agent in linux
func (d *DatadogWindowsOperation) UninstallAgent() error {
	wmiCmd := `(Get-Package -Name "Datadog Agent").Metadata['ProductCode']`
	out, err := d.program.ExecuteWithOutput("powershell", []string{}, "-Command", wmiCmd)
	if err != nil {
		d.logger.Error("error in get wmi datadog agent", "error", err)
		return err
	}
	d.logger.Debug("output wmi datadog agent", "output", out)
	re := regexp.MustCompile(`\{[A-Fa-f0-9\-]+\}`)
	identifyingNumber := re.FindString(string(out))

	if identifyingNumber == "" {
		return fmt.Errorf("not found identify number for datadog agent")
	}

	outUnistall, err := d.program.ExecuteWithOutput("powershell", []string{}, "-Command", fmt.Sprintf(`start-process msiexec -Wait -ArgumentList ('/log', 'C:\uninst.log', '/norestart', '/q', '/x', '%s')`, identifyingNumber))
	if err != nil {
		d.logger.Error("error in uninstall datadog agent", "error", err)
		return err
	}
	d.logger.Debug("output uninstall agent datadog", "output", outUnistall)
	return nil
}

// DiscoverDatadogConfigPath return file path the datadog config
func (d *DatadogWindowsOperation) DiscoverDatadogConfigPath() (string, error) {
	ddConfPathEnv := os.Getenv("DD_CONF_PATH")
	if len(ddConfPathEnv) > 0 {
		return ddConfPathEnv, nil
	}
	programData := os.Getenv("ProgramData")

	datadogPath := filepath.Join(programData, "Datadog")
	return datadogPath, nil
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
	docpFilePath, err := utils.GetWorkDirPath()
	if err != nil {
		return err
	}
	programData := os.Getenv("ProgramData")
	basePath := filepath.Join(programData, "Datadog")
	filteredPath := strings.TrimPrefix(filePath, basePath)
	filePathState := filepath.Join(docpFilePath, "state", "Datadog", filteredPath)
	if err := d.fileSystem.VerifyFileExist(filePathState); err != nil {
		if errCrt := d.fileSystem.CreatePathCompleted(filePathState); errCrt != nil {
			return errCrt
		}
	}
	if err := d.fileSystem.WriteFileContent(filePathState, content); err != nil {
		return err
	}
	return nil
}

// UpdateConfigFileDatadog execute update the config file datadog
func (d *DatadogWindowsOperation) UpdateConfigFileDatadog(filePath string) error {
	docpFilePath, err := utils.GetWorkDirPath()
	if err != nil {
		return err
	}
	programData := os.Getenv("ProgramData")
	basePath := filepath.Join(programData, "Datadog")
	filteredPath := strings.TrimPrefix(filePath, basePath)

	datadogFilePathDir := filepath.Dir(filePath)
	docpStateDatadogPath := filepath.Join(docpFilePath, "state", "Datadog", filteredPath)

	if err := os.MkdirAll(datadogFilePathDir, os.ModePerm); err != nil {
		return err
	}

	content, err := os.ReadFile(docpStateDatadogPath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		return err
	}
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

// GetLatestVersion return the latest version of the datadog agent
func (d *DatadogWindowsOperation) GetLatestVersion() (string, error) {
	return "", nil
}

// UpdateVersion execute update the version of the datadog agent
func (d *DatadogWindowsOperation) UpdateVersion(version string) error {
	return nil
}

// RollbackVersion execute rollback the version of the datadog agent
func (d *DatadogWindowsOperation) RollbackVersion(version string) error {
	return nil
}

// DPKGConfigure execute configure dpkg
func (d *DatadogWindowsOperation) DPKGConfigure() error {
	// TODO: not implemented windows
	return nil
}
