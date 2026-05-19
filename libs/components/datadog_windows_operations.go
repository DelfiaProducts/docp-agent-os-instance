//go:build windows

package components

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/OryaHub/agent-os-instance/libs/dto"
	"github.com/OryaHub/agent-os-instance/libs/interfaces"
	"github.com/OryaHub/agent-os-instance/libs/pkg"
	"github.com/OryaHub/agent-os-instance/libs/services"
	"github.com/OryaHub/agent-os-instance/libs/utils"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

type DatadogWindowsOperation struct {
	logger           interfaces.ILogger
	program          *pkg.ExecProgram
	hostStats        *pkg.HostStats
	stateCheck       *services.StateCheckService
	utilityService   *services.UtilityService
	fileSystem       *pkg.FileSystem
	ymlClient        *pkg.YmlClient
	json             *pkg.JsonClient
	datadogApmTracer *DatadogWindowsAPMTracer
}

func NewDatadogWindowsOperation(logger interfaces.ILogger) *DatadogWindowsOperation {
	return &DatadogWindowsOperation{
		logger: logger,
	}
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
	utilityService := services.NewUtilityService(d.logger)
	if err := utilityService.Setup(); err != nil {
		return err
	}
	d.utilityService = utilityService
	ymlClient := pkg.NewYmlClient()
	d.ymlClient = ymlClient
	return nil
}

// InstallAgent execute install the agent in linux
func (d *DatadogWindowsOperation) InstallAgent(ddSite, ddApiKey, version string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService("DatadogAgent")
	if err != nil {
		d.logger.Warn("error in install datadog agent", "error", err)
		if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			fileVersionUrl := utils.ChoiceMsiWindowsInstallerFileUrl(version)
			command := fmt.Sprintf(`Start-Process -Wait msiexec -ArgumentList '/qn /i %s APIKEY="%s" SITE="%s"'`, fileVersionUrl, ddApiKey, ddSite)
			out, err := d.program.ExecuteWithOutput("powershell", []string{}, "-Command", command)
			if err != nil {
				d.logger.Error("error in install datadog agent start process", "error", err)
				return err
			}
			d.logger.Debug("install agent datadog", "output", out)
			if err := d.WriteEnvironmentFile(ddSite, ddApiKey); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	defer s.Close()
	return nil
}

// InstallAgentApmSingleStep execute install the agent in linux with apm tracer on mode single step
func (d *DatadogWindowsOperation) InstallAgentApmSingleStep(ddSite string, ddApiKey string, ddApmInstrumentationLibraries string) error {
	d.logger.Debug("install agent apm single step", "trace", "agent-os-instance.datadog_windows_operations.InstallAgentApmSingleStep")
	ddApmInstrumentationEnabled := utils.GetApmInstrumentationEnabled("windows")
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService("DatadogAgent")
	if err != nil {
		d.logger.Warn("error in install datadog agent", "error", err)
		if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			version, err := d.getVersionDatadogAgentFromSignal()
			if err != nil {
				d.logger.Error("error in get version datadog agent from signal", "error", err)
				return err
			}
			fileVersionUrl := utils.ChoiceMsiWindowsInstallerFileUrl(version)
			command := fmt.Sprintf(`Start-Process -Wait msiexec -ArgumentList '/qn /i %s APIKEY="%s" SITE="%s" DD_APM_INSTRUMENTATION_ENABLED="%s" DD_APM_INSTRUMENTATION_LIBRARIES="%s"'`, fileVersionUrl, ddApiKey, ddSite, ddApmInstrumentationEnabled, ddApmInstrumentationLibraries)
			out, err := d.program.ExecuteWithOutput("powershell", []string{}, "-Command", command)
			if err != nil {
				d.logger.Error("error in install datadog agent with apm single step start process", "error", err)
				return err
			}
			d.logger.Debug("install agent datadog with apm single step", "output", out)
			if err := d.WriteEnvironmentFile(ddSite, ddApiKey); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	defer s.Close()
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
	oryaFilePath, err := utils.GetWorkDirPath()
	if err != nil {
		return err
	}
	programData := os.Getenv("ProgramData")
	basePath := filepath.Join(programData, "Datadog")
	filteredPath := strings.TrimPrefix(filePath, basePath)
	filePathState := filepath.Join(oryaFilePath, "state", "Datadog", filteredPath)
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
	oryaFilePath, err := utils.GetWorkDirPath()
	if err != nil {
		return err
	}
	programData := os.Getenv("ProgramData")
	basePath := filepath.Join(programData, "Datadog")
	filteredPath := strings.TrimPrefix(filePath, basePath)

	datadogFilePathDir := filepath.Dir(filePath)
	oryaStateDatadogPath := filepath.Join(oryaFilePath, "state", "Datadog", filteredPath)

	if err := os.MkdirAll(datadogFilePathDir, os.ModePerm); err != nil {
		return err
	}

	content, err := os.ReadFile(oryaStateDatadogPath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		return err
	}
	return nil
}

// UpdateConfigFileDatadogContent execute update the config file datadog with content.
// The caller is responsible for restarting the service if needed.
//
// Supported fields:
//   - HostTags: merged with the existing tags in the file (existing tags win on key conflict).
func (d *DatadogWindowsOperation) UpdateConfigFileDatadogContent(filePath string, config dto.DatadogConfigDTO) error {
	d.logger.Debug("update config file datadog content", "trace", "agent-os-instance.datadog_windows_operations.UpdateConfigFileDatadogContent", "filePath", filePath)

	if len(config.HostTags) > 0 {
		if _, err := d.applyHostTags(filePath, config.HostTags); err != nil {
			return err
		}
	}

	return nil
}

// UpdateRepository execute update repository local
func (d *DatadogWindowsOperation) UpdateRepository() error {
	return nil
}

// GetVersion return the version of the datadog agent
func (d *DatadogWindowsOperation) GetVersion() (string, error) {
	m, err := mgr.Connect()
	if err != nil {
		return "", err
	}
	defer m.Disconnect()

	s, err := m.OpenService("DatadogAgent")
	if err != nil {
		if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			return "inactive", nil
		}
		return "", err
	}
	defer s.Close()

	config, err := s.Config()
	if err != nil {
		return "", err
	}

	binaryPath := config.BinaryPathName
	path := strings.Trim(binaryPath, "\"")
	var version string
	buf := new(bytes.Buffer)
	agentExe := filepath.Join(filepath.Dir(path), "agent.exe")
	execCmd := exec.Command(agentExe, "version")
	execCmd.Stdout = buf
	execCmd.Stderr = buf
	if err := execCmd.Run(); err != nil {
		return "", err
	}
	re := regexp.MustCompile(`\d+\.\d+\.\d+`)
	version = re.FindString(buf.String())
	if version == "" {
		return "", utils.ErrDatadogVersionNotFound()
	}
	return version, nil
}

// GetLatestVersion return the latest version of the datadog agent
func (d *DatadogWindowsOperation) GetLatestVersion() (string, error) {
	latestVersion, err := d.utilityService.GetDatadogLastVersionFromGithub()
	if err != nil {
		return "", err
	}
	return latestVersion, nil
}

// UpdateVersion execute update the version of the datadog agent
func (d *DatadogWindowsOperation) UpdateVersion(version string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	applyVersion := version
	if version == "latest" {
		applyVersion, err = d.GetLatestVersion()
		if err != nil {
			return err
		}
	}

	actualVersion, err := d.GetVersion()
	if err != nil {
		return err
	}

	isGreaten, err := utils.IsVersionGreater(actualVersion, applyVersion)
	if err != nil {
		return err
	}

	//verify if execute downgrade
	if isGreaten {
		d.logger.Debug("update version with downgrade version", "actualVersion", actualVersion, "version", version)
		if err := d.UninstallAgent(); err != nil {
			return err
		}
		var datadogYamlDto dto.DatadogYamlDTO
		programDataEnv := os.Getenv("ProgramData")
		datadogYmlPath := filepath.Join(programDataEnv, "Datadog", "datadog.yaml")
		content, err := d.fileSystem.GetFileContent(datadogYmlPath)
		if err != nil {
			return err
		}

		if err := d.ymlClient.Unmarshall(content, &datadogYamlDto); err != nil {
			return err
		}

		if err := d.InstallAgent(datadogYamlDto.ApiKey, datadogYamlDto.Site, version); err != nil {
			return err
		}

		return nil
	}
	d.logger.Debug("connect manager service", "manager", m)
	defer m.Disconnect()
	s, err := m.OpenService("DatadogAgent")
	d.logger.Debug("open service datadog", "service", s)
	if err != nil {
		return err
	}

	defer s.Close()
	fileVersionUrl := utils.ChoiceMsiWindowsInstallerFileUrl(version)
	d.logger.Debug("file version", "fileVersionUrl", fileVersionUrl)
	command := fmt.Sprintf(`Start-Process -Wait msiexec -ArgumentList '/qn /i %s'`, fileVersionUrl)
	out, err := d.program.ExecuteWithOutput("powershell", []string{}, "-Command", command)
	if err != nil {
		d.logger.Error("error in update datadog agent start process", "error", err)
		return err
	}

	d.logger.Debug("update agent datadog", "output", out)
	return nil
}

// RollbackVersion execute rollback the version of the datadog agent
func (d *DatadogWindowsOperation) RollbackVersion(version string) error {
	d.logger.Debug("rollback version", "version", version)
	//execute update version with rollback version
	if err := d.UpdateVersion(version); err != nil {
		return err
	}
	return nil
}

// DPKGConfigure execute configure dpkg
func (d *DatadogWindowsOperation) DPKGConfigure() error {
	// TODO: not implemented windows
	return nil
}

// GetInfos fetch datadog infos
func (d *DatadogWindowsOperation) GetInfos() ([]byte, error) {
	output, err := d.program.ExecuteWithOutput("powershell", []string{}, "-NoProfile", "-Command", `& "$env:ProgramFiles\\Datadog\\Datadog Agent\\bin\\agent.exe"`, "status")
	if err != nil {
		return nil, err
	}
	output = utils.RemoveLinesByPrefix([]string{"ERROR", "Error"}, output)
	datadogInfos := dto.DatadogInfos{
		HostId:                        utils.ParseValueByPrefix(output, "hostId:"),
		Hostname:                      utils.ParseValueByPrefix(output, "hostname:"),
		KernelArch:                    utils.ParseValueByPrefix(output, "kernelArch:"),
		KernelVersion:                 utils.ParseValueByPrefix(output, "kernelVersion:"),
		Os:                            utils.ParseValueByPrefix(output, "os:"),
		Platform:                      utils.ParseValueByPrefix(output, "platform:"),
		PlatformFamily:                utils.ParseValueByPrefix(output, "platformFamily:"),
		PlatformVersion:               utils.ParseValueByPrefix(output, "platformVersion:"),
		AgentVersion:                  utils.ParseValueByPrefix(output, "agent_version:"),
		Flavor:                        utils.ParseValueByPrefix(output, "flavor:"),
		InfrastructureMode:            utils.ParseValueByPrefix(output, "infrastructure_mode:"),
		InstallMethodInstallerVersion: utils.ParseValueByPrefix(output, "install_method_installer_version:"),
		InstallMethodTool:             utils.ParseValueByPrefix(output, "install_method_tool:"),
		InstallMethodToolVersion:      utils.ParseValueByPrefix(output, "install_method_tool_version:"),
	}

	data, err := d.json.Marshall(datadogInfos)
	return data, err
}

// WriteEnvironmentFile write DD_API_KEY and DD_SITE to C:\ProgramData\Datadog\environment
// so credentials persist even when datadog.yaml is overwritten by cloud config
func (d *DatadogWindowsOperation) WriteEnvironmentFile(ddSite, ddApiKey string) error {
	d.logger.Debug("write environment file", "trace", "agent-os-instance.datadog_windows_operations.WriteEnvironmentFile")
	datadogPath, err := d.DiscoverDatadogConfigPath()
	if err != nil {
		return err
	}
	envFilePath := filepath.Join(datadogPath, "environment")
	if err := d.fileSystem.VerifyFileExist(envFilePath); err != nil {
		if errCrt := d.fileSystem.CreateFile(envFilePath); errCrt != nil {
			return errCrt
		}
	}
	content := fmt.Sprintf("DD_API_KEY=%s\nDD_SITE=%s\n", ddApiKey, ddSite)
	command := fmt.Sprintf(`Set-Content -Path '%s' -Value '%s' -Force`, envFilePath, content)
	if _, err := d.program.ExecuteWithOutput("powershell", []string{}, "-NoProfile", "-Command", command); err != nil {
		return err
	}
	return nil
}
