package components

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/pkg"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

var (
	DEFAULT_LIBRARY_FILE_PATH            = "/opt/docp-agent/shared"
	DATADOG_PHP_FILE_NAME                = "datadog-setup.php"
	DATADOG_JAVA_FILE_NAME               = "dd-java-agent.jar"
	DEFAULT_VERSION_DOT_NET_TRACER       = "3.7.0"
	DATADOG_LIBRARY_JAVA_URL_INSTALLER   = "https://dtdg.co/latest-java-tracer"
	DATADOG_LIBRARY_PHP_URL_INSTALLER    = fmt.Sprintf("https://github.com/DataDog/dd-trace-php/releases/latest/download/%s", DATADOG_PHP_FILE_NAME)
	DATADOG_LIBRARY_DOTNET_URL_INSTALLER = "https://github.com/DataDog/dd-trace-dotnet/releases/download"
)

// DatadogAPMTracer is struct for apm tracer libraries
type DatadogAPMTracer struct {
	logger     interfaces.ILogger
	program    *pkg.ExecProgram
	fileSystem *pkg.FileSystem
}

// NewDatadogAPMTracer return new instance of datadog apm tracer
func NewDatadogAPMTracer() *DatadogAPMTracer {
	logger := utils.NewDocpLoggerJSON(os.Stdout)
	return &DatadogAPMTracer{
		logger:     logger,
		program:    pkg.NewExecProgram(),
		fileSystem: pkg.NewFileSystem(),
	}
}

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
	} else { // not exist path tracer, setting /opt/docp-agent/shared
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
	d.logger.Debug("install library java", "trace", "docp-agent-os-instance.datadog_apm_tracer.installLibraryJava", "language", languageName, "pathTracer", pathTracer)
	if err := d.verifyUtilityExist("curl"); err != nil {
		return err
	}
	pathFormated, err := d.formatPathTracer(pathTracer, DATADOG_JAVA_FILE_NAME)
	d.logger.Debug("install library java", "trace", "docp-agent-os-instance.datadog_apm_tracer.installLibraryJava", "pathFormated", pathFormated)
	if err != nil {
		return err
	}
	output, err := d.program.ExecuteWithOutput("sudo", []string{}, "bash", "-c", fmt.Sprintf("curl -kLo %s '%s'", pathFormated, DATADOG_LIBRARY_JAVA_URL_INSTALLER))
	if err != nil {
		return err
	}
	d.logger.Debug("install library java", "trace", "docp-agent-os-instance.datadog_apm_tracer.installLibraryJava", "language", languageName, "output", output)
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
	d.logger.Debug("install library php all", "trace", "docp-agent-os-instance.datadog_apm_tracer.installLibraryPhpAll", "language", languageName)
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
	d.logger.Debug("install library php all", "trace", "docp-agent-os-instance.datadog_apm_tracer.installLibraryPhpAll", "language", languageName, "output", output)
	return nil
}

// installLibraryPhpApmAsm install php with apm and asm
func (d *DatadogAPMTracer) installLibraryPhpApmAsm(languageName, pathTracer string) error {
	d.logger.Debug("install library php apm asm", "trace", "docp-agent-os-instance.datadog_apm_tracer.installLibraryPhpApmAsm", "language", languageName, "pathTracer", pathTracer)
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
	d.logger.Debug("install library php apm asm", "trace", "docp-agent-os-instance.datadog_apm_tracer.installLibraryPhpApmAsm", "language", languageName, "pathTracer", pathTracer, "output", output)
	return nil
}

// installLibraryPhpApmProfiling install php with apm and profiling
func (d *DatadogAPMTracer) installLibraryPhpApmProfiling(languageName, pathTracer string) error {
	d.logger.Debug("install library php apm profiling", "trace", "docp-agent-os-instance.datadog_apm_tracer.installLibraryPhpApmProfiling", "language", languageName, "pathTracer", pathTracer)
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
	d.logger.Debug("install library php apm profiling", "trace", "docp-agent-os-instance.datadog_apm_tracer.installLibraryPhpApmProfiling", "language", languageName, "output", output)
	return nil
}

// installLibraryPhpApmOnly install php with apm only
func (d *DatadogAPMTracer) installLibraryPhpApmOnly(languageName, pathTracer string) error {
	d.logger.Debug("install library php apm only", "trace", "docp-agent-os-instance.datadog_apm_tracer.installLibraryPhpApmOnly", "language", languageName, "pathTracer", pathTracer)
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
	d.logger.Debug("install library php apm only", "trace", "docp-agent-os-instance.datadog_apm_tracer.installLibraryPhpApmOnly", "language", languageName, "output", output)
	return nil
}

// installLibraryNetCore install dotnet core
func (d *DatadogAPMTracer) installLibraryNetCore(languageName, pathTracer, version string) error {
	d.logger.Debug("install library net core", "trace", "docp-agent-os-instance.datadog_apm_tracer.installLibraryNetCore", "language", languageName, "pathTracer", pathTracer, "version", version)
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

	d.logger.Debug("install library net core", "trace", "docp-agent-os-instance.datadog_apm_tracer.installLibraryNetCore", "language", languageName, "outputDpkg", outputDpkg)
	output, err := d.program.ExecuteWithOutput("sudo", []string{}, "bash", "-c", "/opt/datadog/createLogPath.sh")
	if err != nil {
		return err
	}
	d.logger.Debug("install library net core", "trace", "docp-agent-os-instance.datadog_apm_tracer.installLibraryNetCore", "language", languageName, "output", output)

	return nil
}

// InstallLibrary execute install the library of language
func (d *DatadogAPMTracer) InstallLibrary(languageName, pathTracer, version string) error {
	d.logger.Debug("install library", "trace", "docp-agent-os-instance.datadog_apm_tracer.InstallLibrary", "language", languageName, "pathTracer", pathTracer, "version", version)
	switch languageName {
	case "java":
		return d.installLibraryJava(languageName, pathTracer)
	case "php_full":
		return d.installLibraryPhpAll(languageName, pathTracer)
	case "php_apm_only":
		return d.installLibraryPhpApmOnly(languageName, pathTracer)
	case "php_apm_asm":
		return d.installLibraryPhpApmAsm(languageName, pathTracer)
	case "php_apm_profiling":
		return d.installLibraryPhpApmProfiling(languageName, pathTracer)
	case "net_core":
		return d.installLibraryNetCore(languageName, pathTracer, version)
	default:
		return errors.New("not implemented language")
	}
}
