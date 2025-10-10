//go:build windows

package components

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/OryaHub/agent-os-instance/libs/interfaces"
	"github.com/OryaHub/agent-os-instance/libs/pkg"
	"github.com/OryaHub/agent-os-instance/libs/utils"
)

// DatadogWindowsAPMTracer is struct for datadog windows apm tracer
type DatadogWindowsAPMTracer struct {
	logger     interfaces.ILogger
	program    *pkg.ExecProgram
	fileSystem *pkg.FileSystem
}

// NewDatadogWindowsAPMTracer return new instance of datadog windows apm tracer
func NewDatadogWindowsAPMTracer() *DatadogWindowsAPMTracer {
	logger := utils.NewOryaLoggerJSON(os.Stdout)
	return &DatadogWindowsAPMTracer{
		logger:     logger,
		program:    pkg.NewExecProgram(),
		fileSystem: pkg.NewFileSystem(),
	}
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
	d.logger.Debug("install library", "trace", "agent-os-instance.datadog_windows_apm_tracer.InstallLibrary", "language", languageName, "pathTracer", pathTracer, "version", version)
	switch languageName {
	case "net_core":
		return d.installLibraryNetCoreWindows(languageName, pathTracer, version)
	default:
		return errors.New("not implemented language")
	}
}
