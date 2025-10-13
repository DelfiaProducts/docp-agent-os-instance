//go:build windows

package components

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

var (
	DEFAULT_VERSION_DOT_NET_TRACER_WINDOWS       = "3.14.2"
	DATADOG_LIBRARY_DOTNET_URL_INSTALLER_WINDOWS = "https://github.com/DataDog/dd-trace-dotnet/releases/download"
)

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
	d.logger.Debug("install library net core windows", "trace", "agent-os-instance.datadog_windows_apm_tracer.installLibraryNetCore", "language", languageName, "pathTracer", pathTracer, "version", version)
	url, _ := d.prepareDotNetNameInstallerWindows(version)
	command := fmt.Sprintf(`Start-Process -Wait msiexec -ArgumentList '/qn /i %s'`, url)
	out, err := d.program.ExecuteWithOutput("powershell", []string{}, "-Command", command)
	if err != nil {
		d.logger.Error("error in install datadog agent start process", "error", err)
		return err
	}
	d.logger.Debug("install library net core windows", "trace", "agent-os-instance.datadog_windows_apm_tracer.installLibraryNetCore", "output", out)
	if err := d.setEnvVars(version); err != nil {
		d.logger.Error("error in set envs datadog tracing library dot net windows", "error", err)
		return err
	}
	return nil
}
