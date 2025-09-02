//go:build windows

package components

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/pkg"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
	"golang.org/x/sys/windows/registry"
)

var (
	DEFAULT_VERSION_DOT_NET_TRACER_WINDOWS       = "3.14.2"
	DATADOG_LIBRARY_DOTNET_URL_INSTALLER_WINDOWS = "https://github.com/DataDog/dd-trace-dotnet/releases/download"
)

// DatadogWindowsAPMTracer is struct for datadog windows apm tracer
type DatadogWindowsAPMTracer struct {
	logger     interfaces.ILogger
	program    *pkg.ExecProgram
	fileSystem *pkg.FileSystem
}

// NewDatadogWindowsAPMTracer return new instance of datadog windows apm tracer
func NewDatadogWindowsAPMTracer() *DatadogWindowsAPMTracer {
	logger := utils.NewDocpLoggerJSON(os.Stdout)
	return &DatadogWindowsAPMTracer{
		logger:     logger,
		program:    pkg.NewExecProgram(),
		fileSystem: pkg.NewFileSystem(),
	}
}

// prepareDotNetNameInstallerWindows prepare installer for dot net
func (d *DatadogWindowsAPMTracer) prepareDotNetNameInstallerWindows(version string) (string, string) {
	arch := "x64"
	if len(version) == 0 {
		version = DEFAULT_VERSION_DOT_NET_TRACER_WINDOWS
	}
	url := fmt.Sprintf("%s/v%s/datadog-dotnet-apm-%s-%s.msi", DATADOG_LIBRARY_DOTNET_URL_INSTALLER_WINDOWS, version, version, arch)
	filename := fmt.Sprintf("datadog-dotnet-apm-%s-%s.msi", version, arch)
	return url, filename
}

func setEnvForService(name, value string) error {
	k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	if err := k.SetStringValue(name, value); err != nil {
		return err
	}

	return nil
}

// setEnvForService set env for scm registry windows
func (d *DatadogWindowsAPMTracer) setEnvForService(name, value string) error {
	k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	if err := k.SetStringValue(name, value); err != nil {
		return err
	}

	return nil
}

func (d *DatadogWindowsAPMTracer) setEnvVars(version string) error {
	vars := map[string]string{
		"CORECLR_ENABLE_PROFILING": "1",
		"CORECLR_PROFILER":         "{846F5F1C-F9AE-4B07-969E-05C26BC060D8}",
		"COR_ENABLE_PROFILING":     "1",
		"COR_PROFILER":             "{846F5F1C-F9AE-4B07-969E-05C26BC060D8}",
		"DD_VERSION":               version,
		"DD_LOGS_INJECTION":        "true",
	}

	for key, value := range vars {
		if err := d.setEnvForService(key, value); err != nil {
			return err
		}
	}
	return nil
}

// installLibraryNetCoreWindows install dotnet core
func (d *DatadogWindowsAPMTracer) installLibraryNetCoreWindows(languageName, pathTracer, version string) error {
	d.logger.Debug("install library net core windows", "trace", "docp-agent-os-instance.datadog_windows_apm_tracer.installLibraryNetCore", "language", languageName, "pathTracer", pathTracer, "version", version)
	url, _ := d.prepareDotNetNameInstallerWindows(version)
	command := fmt.Sprintf(`Start-Process -Wait msiexec -ArgumentList '/qn /i %s'`, url)
	out, err := d.program.ExecuteWithOutput("powershell", []string{}, "-Command", command)
	if err != nil {
		d.logger.Error("error in install datadog agent start process", "error", err)
		return err
	}
	d.logger.Debug("install library net core windows", "trace", "docp-agent-os-instance.datadog_windows_apm_tracer.installLibraryNetCore", "output", out)
	if err := d.setEnvVars(version); err != nil {
		d.logger.Error("error in set envs datadog tracing library dot net windows", "error", err)
		return err
	}
	return nil
}

// UpdateConfiApmTracer update configuration for apm tracer
func (d *DatadogWindowsAPMTracer) UpdateConfiApmTracer(path string, enable bool) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	newLines := []string{}
	updated := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trim := strings.TrimSpace(line)

		if strings.HasPrefix(trim, "apm_enabled:") || strings.HasPrefix(trim, "#apm_enabled:") {
			newLines = append(newLines, "apm_enabled: "+fmt.Sprintf("%t", enable))
			updated = true
			continue
		}

		if strings.HasPrefix(trim, "apm_config:") || strings.HasPrefix(trim, "#apm_config:") {
			newLines = append(newLines, "apm_config:")
			i++
			if i < len(lines) && strings.Contains(lines[i], "enabled") {
				newLines = append(newLines, "  enabled: "+fmt.Sprintf("%t", enable))
			} else {
				newLines = append(newLines, "  enabled: "+fmt.Sprintf("%t", enable))
				i--
			}
			updated = true
			continue
		}

		newLines = append(newLines, line)
	}

	if !updated {
		newLines = append(newLines, "apm_enabled: "+fmt.Sprintf("%t", enable))
	}

	if err := os.WriteFile(path, []byte(strings.Join(newLines, "\n")), 0o644); err != nil {
		return err
	}
	return nil
}

// InstallLibrary execute install the library of language
func (d *DatadogWindowsAPMTracer) InstallLibrary(languageName, pathTracer, version string) error {
	d.logger.Debug("install library", "trace", "docp-agent-os-instance.datadog_windows_apm_tracer.InstallLibrary", "language", languageName, "pathTracer", pathTracer, "version", version)
	switch languageName {
	case "net_core":
		return d.installLibraryNetCoreWindows(languageName, pathTracer, version)
	default:
		return errors.New("not implemented language")
	}
}
