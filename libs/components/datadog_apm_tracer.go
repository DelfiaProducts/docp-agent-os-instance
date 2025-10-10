package components

import (
	"errors"
	"fmt"
	"os"

	"github.com/OryaHub/agent-os-instance/libs/interfaces"
	"github.com/OryaHub/agent-os-instance/libs/pkg"
	"github.com/OryaHub/agent-os-instance/libs/utils"
)

var (
	DEFAULT_LIBRARY_FILE_PATH            = "/opt/orya-agent/shared"
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
	logger := utils.NewOryaLoggerJSON(os.Stdout)
	return &DatadogAPMTracer{
		logger:     logger,
		program:    pkg.NewExecProgram(),
		fileSystem: pkg.NewFileSystem(),
	}
}

// InstallLibrary execute install the library of language
func (d *DatadogAPMTracer) InstallLibrary(languageName, pathTracer, version string) error {
	d.logger.Debug("install library", "trace", "agent-os-instance.datadog_apm_tracer.InstallLibrary", "language", languageName, "pathTracer", pathTracer, "version", version)
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
