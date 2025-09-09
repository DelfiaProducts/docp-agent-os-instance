package components

import (
	"bufio"
	"bytes"
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
)

const (
	CURL_INSTALL_SH         = "curl -L https://install.datadoghq.com/scripts/install_script_agent7.sh | bash"
	UNINSTALL_AGENT_COMMAND = "sudo apt remove --purge datadog-agent -y"
	REMOVE_FILES_CONFIGS    = "sudo rm -rf /etc/datadog-agent"
	REMOVE_FILES_LOGS       = "sudo rm -rf /var/log/datadog"
)

type DatadogLinuxOperation struct {
	logger           interfaces.ILogger
	program          *pkg.ExecProgram
	hostStats        *pkg.HostStats
	stateCheck       *services.StateCheckService
	fileSystem       *pkg.FileSystem
	datadogApmTracer *DatadogAPMTracer
}

func NewDatadogLinuxOperation(logger interfaces.ILogger) *DatadogLinuxOperation {
	return &DatadogLinuxOperation{
		logger: logger,
	}
}

// prepareEnvs return envs the datadog
func (d *DatadogLinuxOperation) prepareEnvs(ddSite, ddApiKey string) []string {
	var envs []string
	envs = append(envs, fmt.Sprintf("DD_API_KEY=%s", ddApiKey))
	envs = append(envs, fmt.Sprintf("DD_SITE=%s", ddSite))
	return envs
}

// getApmEnvVarsSingleStep get envs apm datadog in mode single step
func (d *DatadogLinuxOperation) getApmEnvVarsSingleStep(envs []dto.DatadogEnvVars) (string, string, string) {
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

// parseDatadogAgentVersion parses the installed version from the output of `apt-cache policy datadog-agent`
func (d *DatadogLinuxOperation) parseDatadogAgentVersion(output string, prefix string) (string, error) {
	lines := strings.Split(output, "\n")
	re := regexp.MustCompile(fmt.Sprintf(`%s\s*([0-9]+:)?([0-9]+\.[0-9]+\.[0-9]+)-[0-9]+`, prefix))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 3 {
				return matches[2], nil
			}
		}
	}
	return "", errors.New("installed version not found")
}

func (d *DatadogLinuxOperation) getVersionFromOutput(output []byte, version string) (string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, version) {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				return fields[0], nil
			}
		}
	}

	return "", utils.ErrDatadogVersionNotFound()
}

func (d *DatadogLinuxOperation) Setup() error {
	execProgram := pkg.NewExecProgram()
	d.program = execProgram
	hostStats := pkg.NewHostStats()
	d.hostStats = hostStats
	stateCheck := services.NewStateCheckService(d.logger)
	if err := stateCheck.Setup(); err != nil {
		return err
	}
	d.stateCheck = stateCheck
	datadogApmTracer := NewDatadogAPMTracer()
	d.datadogApmTracer = datadogApmTracer
	fileSystem := pkg.NewFileSystem()
	d.fileSystem = fileSystem
	return nil
}

// InstallAgent execute install the agent in linux
func (d *DatadogLinuxOperation) InstallAgent(ddSite, ddApiKey string) error {
	envs := d.prepareEnvs(ddSite, ddApiKey)
	aptOrDpkgIsRunning, err := d.hostStats.AptOrDpkgIsRunning()
	if err != nil {
		return err
	}
	if !aptOrDpkgIsRunning {
		d.logger.Debug("install agent", "trace", "docp-agent-os-instance.datadog_linux_operations.InstallAgent", "aptOrDpkgIsRunning", aptOrDpkgIsRunning)
		if err := d.program.Execute("bash", envs, "-c", fmt.Sprintf("export DD_API_KEY=%s;export DD_SITE=%s;%s", ddApiKey, ddSite, CURL_INSTALL_SH)); err != nil {
			return err
		}
	} else {
		d.logger.Debug("install agent", "trace", "docp-agent-os-instance.datadog_linux_operations.InstallAgent", "aptOrDpkgIsRunning", aptOrDpkgIsRunning)
	}

	return nil
}

// InstallAgentApmSingleStep execute install the agent in linux with apm tracer on mode single step
func (d *DatadogLinuxOperation) InstallAgentApmSingleStep(ddSite string, ddApiKey string, datadogEnvVars []dto.DatadogEnvVars) error {
	envs := d.prepareEnvs(ddSite, ddApiKey)
	aptOrDpkgIsRunning, err := d.hostStats.AptOrDpkgIsRunning()
	if err != nil {
		return err
	}
	ddApmInstrumentationEnabled, ddEnv, ddApmInstrumentationLibraries := d.getApmEnvVarsSingleStep(datadogEnvVars)
	if !aptOrDpkgIsRunning {
		d.logger.Debug("install agent apm single step", "trace", "docp-agent-os-instance.datadog_linux_operations.InstallAgentApmSingleStep", "aptOrDpkgIsRunning", aptOrDpkgIsRunning)
		if err := d.program.Execute("bash", envs, "-c", fmt.Sprintf("export DD_API_KEY=%s;export DD_SITE=%s;export DD_APM_INSTRUMENTATION_ENABLED=%s;export DD_ENV=%s;export DD_APM_INSTRUMENTATION_LIBRARIES=%s;%s", ddApiKey, ddSite, ddApmInstrumentationEnabled, ddEnv, ddApmInstrumentationLibraries, CURL_INSTALL_SH)); err != nil {
			return err
		}
	} else {
		d.logger.Debug("install agent apm single step", "trace", "docp-agent-os-instance.datadog_linux_operations.InstallAgentApmSingleStep", "aptOrDpkgIsRunning", aptOrDpkgIsRunning)
	}

	return nil
}

// InstallAgentApmTracingLibrary execute install the agent in linux with apm tracer on mode tracing library
func (d *DatadogLinuxOperation) InstallAgentApmTracingLibrary(languageName, pathTracer, version string) error {
	if err := d.datadogApmTracer.InstallLibrary(languageName, pathTracer, version); err != nil {
		return err
	}

	return nil
}

// UninstallAgent execute uninstall the agent in linux
func (d *DatadogLinuxOperation) UninstallAgent() error {
	aptOrDpkgIsRunning, err := d.hostStats.AptOrDpkgIsRunning()
	if err != nil {
		return err
	}
	if !aptOrDpkgIsRunning {
		d.logger.Debug("uninstall agent", "trace", "docp-agent-os-instance.datadog_linux_operations.UninstallAgent", "aptOrDpkgIsRunning", aptOrDpkgIsRunning)
		if err := d.program.Execute("bash", []string{}, "-c", UNINSTALL_AGENT_COMMAND); err != nil {
			return err
		}
		if err := d.program.Execute("bash", []string{}, "-c", REMOVE_FILES_CONFIGS); err != nil {
			return err
		}
		if err := d.program.Execute("bash", []string{}, "-c", REMOVE_FILES_LOGS); err != nil {
			return err
		}
	} else {
		d.logger.Debug("uninstall agent", "trace", "docp-agent-os-instance.datadog_linux_operations.UninstallAgent", "aptOrDpkgIsRunning", aptOrDpkgIsRunning)
	}

	return nil
}

// DiscoverDatadogConfigPath return file path the datadog config
func (d *DatadogLinuxOperation) DiscoverDatadogConfigPath() (string, error) {
	ddConfPathEnv := os.Getenv("DD_CONF_PATH")
	if len(ddConfPathEnv) > 0 {
		return ddConfPathEnv, nil
	}
	exist, err := d.fileSystem.VerifyDirExist("/etc/datadog-agent")
	if err != nil {
		return "", err
	}
	if exist {
		return "/etc/datadog-agent", nil
	}
	return "", errors.New("datadog config path not found")
}

// DatadogAddPermitionGroupFilePath add permition for file path the datadog
func (d *DatadogLinuxOperation) DatadogAddPermitionGroupFilePath(filePath string) error {
	if err := d.program.Execute("sudo", []string{}, "chmod", "-R", "g+rw", filePath); err != nil {
		return err
	}
	return nil
}

// DatadogAddPermitionUser add permition for directory the datadog
func (d *DatadogLinuxOperation) DatadogAddPermitionUser() error {
	if err := d.program.Execute("sudo", []string{}, "usermod", "-aG", "dd-agent", "docp-agent"); err != nil {
		return err
	}
	return nil
}

// BackupConfigFileDatadog execute backup the current config file datadog
func (d *DatadogLinuxOperation) BackupConfigFileDatadog(filePath string, content []byte) error {
	docpFilePath, err := utils.GetWorkDirPath()
	if err != nil {
		return err
	}
	filePathState := filepath.Join(docpFilePath, "state", "datadog", filePath)
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
func (d *DatadogLinuxOperation) UpdateConfigFileDatadog(filePath string) error {
	docpFilePath, err := utils.GetWorkDirPath()
	if err != nil {
		return err
	}
	datadogFilePathDir := filepath.Dir(filePath)
	docpStateDatadogPath := filepath.Join(docpFilePath, "state", "datadog", filePath)
	if err := d.program.Execute("sudo", []string{}, "-u", "dd-agent", "bash", "-c", fmt.Sprintf("mkdir -p %s", datadogFilePathDir)); err != nil {
		return err
	}
	if err := d.program.Execute("sudo", []string{}, "-u", "dd-agent", "bash", "-c", fmt.Sprintf("cat %s | tee %s > /dev/null", docpStateDatadogPath, filePath)); err != nil {
		return err
	}
	return nil
}

// UpdateRepository execute update repository local
func (d *DatadogLinuxOperation) UpdateRepository() error {
	if err := d.program.Execute("sudo", []string{}, "bash", "-c", "apt-get update"); err != nil {
		return err
	}
	return nil
}

// GetVersion return the version of the datadog agent
func (d *DatadogLinuxOperation) GetVersion() (string, error) {
	output, err := d.program.ExecuteWithOutput("apt-cache", []string{}, "policy", "datadog-agent")
	if err != nil {
		return "", err
	}
	return d.parseDatadogAgentVersion(output, "Installed:")
}

// GetLatestVersion return the latest version of the datadog agent
func (d *DatadogLinuxOperation) GetLatestVersion() (string, error) {
	output, err := d.program.ExecuteWithOutput("apt-cache", []string{}, "policy", "datadog-agent")
	if err != nil {
		return "", err
	}
	return d.parseDatadogAgentVersion(output, "Candidate:")
}

// UpdateVersion execute update the version of the datadog agent
func (d *DatadogLinuxOperation) UpdateVersion(version string) error {
	output, err := d.program.ExecuteWithOutput("apt-cache", []string{}, "policy", "datadog-agent")
	if err != nil {
		return err
	}
	datadogVersion, err := d.getVersionFromOutput([]byte(output), version)
	if err != nil {
		return err
	}
	d.logger.Debug("update version", "datadogVersion", datadogVersion)
	if err := d.program.Execute("sudo", []string{}, "apt-get", "install", "-y", "--allow-downgrades", fmt.Sprintf("datadog-agent=%s", datadogVersion)); err != nil {
		return err
	}
	return nil
}

// RollbackVersion execute rollback the version of the datadog agent
func (d *DatadogLinuxOperation) RollbackVersion(version string) error {
	if err := d.UpdateVersion(version); err != nil {
		return err
	}
	return nil
}

// DPKGConfigure execute configure dpkg
func (d *DatadogLinuxOperation) DPKGConfigure() error {
	if err := d.program.Execute("sudo", []string{}, "dpkg", "--configure", "-a"); err != nil {
		return err
	}
	return nil
}
