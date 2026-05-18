//go:build windows

package components

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/OryaHub/agent-os-instance/libs/dto"
	"github.com/OryaHub/agent-os-instance/libs/utils"
)

// getVersionDatadogAgentFromSignal return version the datadog agent from signal
func (d *DatadogWindowsOperation) getVersionDatadogAgentFromSignal() (string, error) {
	workDir, err := utils.GetWorkDirPath()
	if err != nil {
		return "", err
	}

	receivedPath := filepath.Join(workDir, "state", "received")
	content, err := d.fileSystem.GetFileContent(receivedPath)
	if err != nil {
		return "", err
	}

	signalDto := dto.StateCheckResponse{}
	if err := json.Unmarshal(content, &signalDto); err != nil {
		return "", err
	}

	return signalDto.Signal.Agents.DatadogAgent.Version, nil
}

// applyHostTags reads the current content of filePath, merges the existing tags
// with newTags (existing tags win on key conflict), and writes the result back.
// It does NOT restart the service — that is the caller's responsibility.
func (d *DatadogWindowsOperation) applyHostTags(filePath string, newTags []string) (string, error) {
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read datadog config file %s: %w", filePath, err)
	}

	newContent, err := utils.MergeTagsInDatadogConfig(string(raw), newTags)
	if err != nil {
		return "", fmt.Errorf("failed to merge tags in datadog config: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(newContent), 0o644); err != nil {
		return "", fmt.Errorf("failed to write datadog config file %s: %w", filePath, err)
	}

	return newContent, nil
}
