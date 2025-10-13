package components

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// verifyUtilityExist verify if utility already installed
func (d *DatadogAPMTracer) verifyUtilityExist(name string) error {
	_, err := exec.LookPath(name)
	return err
}

// formatPathTracer execute format and verify the path tracer
func (d *DatadogAPMTracer) formatPathTracer(pathTracer, file string) (string, error) {
	if len(pathTracer) > 0 { // exist path tracer
		isDir := d.fileSystem.IsDirectory(pathTracer)
		if isDir {
			formatedPath := d.fileSystem.JoinPaths([]string{pathTracer, file})
			return formatedPath, nil
		} else {
			return "", errors.New("path tracer is not directory")
		}
	} else { // not exist path tracer, setting /opt/orya-agent/shared
		if err := d.fileSystem.VerifyDirExistAndCreate(DEFAULT_LIBRARY_FILE_PATH); err != nil {
			return "", err
		}
		formatedPath := d.fileSystem.JoinPaths([]string{DEFAULT_LIBRARY_FILE_PATH, file})
		return formatedPath, nil
	}
}

// prepareDotNetNameInstaller prepare installer for dot net
func (d *DatadogAPMTracer) prepareDotNetNameInstaller(version string) (string, string) {
	var arch string
	if strings.Contains(runtime.GOARCH, "arm") {
		arch = "arm64"
	} else {
		arch = "amd64"
	}
	if len(version) == 0 {
		version = DEFAULT_VERSION_DOT_NET_TRACER
	}
	//./datadog-dotnet-apm_3.7.0_amd64.deb
	url := fmt.Sprintf("%s/v%s/datadog-dotnet-apm_%s_%s.deb", DATADOG_LIBRARY_DOTNET_URL_INSTALLER, version, version, arch)
	filename := fmt.Sprintf("datadog-dotnet-apm_%s_%s.deb", version, arch)
	return url, filename
}

// installLibraryJava install java library
func (d *DatadogAPMTracer) installLibraryJava(languageName, pathTracer string) error {
	d.logger.Debug("install library java", "trace", "agent-os-instance.datadog_apm_tracer.installLibraryJava", "language", languageName, "pathTracer", pathTracer)
	if err := d.verifyUtilityExist("curl"); err != nil {
		return err
	}
	pathFormated, err := d.formatPathTracer(pathTracer, DATADOG_JAVA_FILE_NAME)
	d.logger.Debug("install library java", "trace", "agent-os-instance.datadog_apm_tracer.installLibraryJava", "pathFormated", pathFormated)
	if err != nil {
		return err
	}
	output, err := d.program.ExecuteWithOutput("sudo", []string{}, "bash", "-c", fmt.Sprintf("curl -kLo %s '%s'", pathFormated, DATADOG_LIBRARY_JAVA_URL_INSTALLER))
	if err != nil {
		return err
	}
	d.logger.Debug("install library java", "trace", "agent-os-instance.datadog_apm_tracer.installLibraryJava", "language", languageName, "output", output)
	return nil
}

// verifyAndDownloadPhpLibrary execute verify and download library
func (d *DatadogAPMTracer) verifyAndDownloadPhpLibrary(pathTracer string) error {
	if err := d.verifyUtilityExist("curl"); err != nil {
		return err
	}
	if err := d.verifyUtilityExist("php"); err != nil {
		return err
	}
	formatedPath, err := d.formatPathTracer(pathTracer, DATADOG_PHP_FILE_NAME)
	if err != nil {
		return err
	}
	if err := d.program.Execute("sudo", []string{}, "bash", "-c", fmt.Sprintf("curl -kLo %s %s", formatedPath, DATADOG_LIBRARY_PHP_URL_INSTALLER)); err != nil {
		return err
	}
	return nil
}

// verifyAndDownloadDotNetLibrary execute verify and download library
func (d *DatadogAPMTracer) verifyAndDownloadDotNetLibrary(pathTracer, version string) error {
	if err := d.verifyUtilityExist("curl"); err != nil {
		return err
	}
	url, filename := d.prepareDotNetNameInstaller(version)
	formatedPath, err := d.formatPathTracer(pathTracer, filename)
	if err != nil {
		return err
	}
	if err := d.program.Execute("sudo", []string{}, "bash", "-c", fmt.Sprintf("curl -kLo %s %s", formatedPath, url)); err != nil {
		return err
	}
	return nil
}

// installLibraryPhpAll install php with apm all
func (d *DatadogAPMTracer) installLibraryPhpAll(languageName, pathTracer string) error {
	d.logger.Debug("install library php all", "trace", "agent-os-instance.datadog_apm_tracer.installLibraryPhpAll", "language", languageName)
	if err := d.verifyAndDownloadPhpLibrary(pathTracer); err != nil {
		return err
	}
	formatedPath, err := d.formatPathTracer(pathTracer, DATADOG_PHP_FILE_NAME)
	if err != nil {
		return err
	}
	output, err := d.program.ExecuteWithOutput("sudo", []string{}, "bash", "-c", fmt.Sprintf("php %s --php-bin=all --enable-appsec --enable-profiling", formatedPath))
	if err != nil {
		return err
	}
	d.logger.Debug("install library php all", "trace", "agent-os-instance.datadog_apm_tracer.installLibraryPhpAll", "language", languageName, "output", output)
	return nil
}

// installLibraryPhpApmAsm install php with apm and asm
func (d *DatadogAPMTracer) installLibraryPhpApmAsm(languageName, pathTracer string) error {
	d.logger.Debug("install library php apm asm", "trace", "agent-os-instance.datadog_apm_tracer.installLibraryPhpApmAsm", "language", languageName, "pathTracer", pathTracer)
	if err := d.verifyAndDownloadPhpLibrary(pathTracer); err != nil {
		return err
	}
	formatedPath, err := d.formatPathTracer(pathTracer, DATADOG_PHP_FILE_NAME)
	if err != nil {
		return err
	}
	output, err := d.program.ExecuteWithOutput("sudo", []string{}, "bash", "-c", fmt.Sprintf("php %s --php-bin=all --enable-appsec", formatedPath))
	if err != nil {
		return err
	}
	d.logger.Debug("install library php apm asm", "trace", "agent-os-instance.datadog_apm_tracer.installLibraryPhpApmAsm", "language", languageName, "pathTracer", pathTracer, "output", output)
	return nil
}

// installLibraryPhpApmProfiling install php with apm and profiling
func (d *DatadogAPMTracer) installLibraryPhpApmProfiling(languageName, pathTracer string) error {
	d.logger.Debug("install library php apm profiling", "trace", "agent-os-instance.datadog_apm_tracer.installLibraryPhpApmProfiling", "language", languageName, "pathTracer", pathTracer)
	if err := d.verifyAndDownloadPhpLibrary(pathTracer); err != nil {
		return err
	}
	formatedPath, err := d.formatPathTracer(pathTracer, DATADOG_PHP_FILE_NAME)
	if err != nil {
		return err
	}
	output, err := d.program.ExecuteWithOutput("sudo", []string{}, "bash", "-c", fmt.Sprintf("php %s --php-bin=all --enable-profiling", formatedPath))
	if err != nil {
		return err
	}
	d.logger.Debug("install library php apm profiling", "trace", "agent-os-instance.datadog_apm_tracer.installLibraryPhpApmProfiling", "language", languageName, "output", output)
	return nil
}

// installLibraryPhpApmOnly install php with apm only
func (d *DatadogAPMTracer) installLibraryPhpApmOnly(languageName, pathTracer string) error {
	d.logger.Debug("install library php apm only", "trace", "agent-os-instance.datadog_apm_tracer.installLibraryPhpApmOnly", "language", languageName, "pathTracer", pathTracer)
	if err := d.verifyAndDownloadPhpLibrary(pathTracer); err != nil {
		return err
	}
	formatedPath, err := d.formatPathTracer(pathTracer, DATADOG_PHP_FILE_NAME)
	if err != nil {
		return err
	}
	output, err := d.program.ExecuteWithOutput("sudo", []string{}, "bash", "-c", fmt.Sprintf("php %s --php-bin=all", formatedPath))
	if err != nil {
		return err
	}
	d.logger.Debug("install library php apm only", "trace", "agent-os-instance.datadog_apm_tracer.installLibraryPhpApmOnly", "language", languageName, "output", output)
	return nil
}

// installLibraryNetCore install dotnet core
func (d *DatadogAPMTracer) installLibraryNetCore(languageName, pathTracer, version string) error {
	d.logger.Debug("install library net core", "trace", "agent-os-instance.datadog_apm_tracer.installLibraryNetCore", "language", languageName, "pathTracer", pathTracer, "version", version)
	fmt.Printf("INSTALL DATADOG LIBRARY LANGUAGE: [%s]\n", languageName)
	if err := d.verifyUtilityExist("dpkg"); err != nil {
		return err
	}
	// prepare datadog dotnet
	if err := d.verifyAndDownloadDotNetLibrary(pathTracer, version); err != nil {
		return err
	}
	_, filename := d.prepareDotNetNameInstaller(version)
	formatedPath, err := d.formatPathTracer(pathTracer, filename)
	if err != nil {
		return err
	}
	outputDpkg, err := d.program.ExecuteWithOutput("sudo", []string{}, "bash", "-c", fmt.Sprintf("dpkg -i %s", formatedPath))
	if err != nil {
		return err
	}

	d.logger.Debug("install library net core", "trace", "agent-os-instance.datadog_apm_tracer.installLibraryNetCore", "language", languageName, "outputDpkg", outputDpkg)
	output, err := d.program.ExecuteWithOutput("sudo", []string{}, "bash", "-c", "/opt/datadog/createLogPath.sh")
	if err != nil {
		return err
	}
	d.logger.Debug("install library net core", "trace", "agent-os-instance.datadog_apm_tracer.installLibraryNetCore", "language", languageName, "output", output)

	return nil
}
