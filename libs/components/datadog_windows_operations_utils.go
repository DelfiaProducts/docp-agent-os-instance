//go:build windows

package components

import (
	"encoding/json"
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
