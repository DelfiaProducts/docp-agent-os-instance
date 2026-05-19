package components

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/OryaHub/agent-os-instance/libs/utils"
)

// applyHostTags reads the current content of filePath as dd-agent, merges
// the existing tags with newTags (existing tags win on key conflict), writes
// the result back as dd-agent using a temp file, and returns the new content.
// It does NOT restart the service — that is the caller's responsibility.
func (d *DatadogLinuxOperation) applyHostTags(filePath string, newTags []string) (string, error) {
	currentContent, err := d.program.ExecuteWithOutput("sudo", []string{}, "-u", "dd-agent", "cat", filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read datadog config file %s: %w", filePath, err)
	}

	newContent, err := utils.MergeTagsInDatadogConfig(currentContent, newTags)
	if err != nil {
		return "", fmt.Errorf("failed to merge tags in datadog config: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "orya-dd-tags-*.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create temp config file: %w", err)
	}
	tmpFilePath := tmpFile.Name()
	defer os.Remove(tmpFilePath)

	if _, err := tmpFile.WriteString(newContent); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write temp config file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp config file: %w", err)
	}
	if err := os.Chmod(tmpFilePath, 0644); err != nil {
		return "", fmt.Errorf("failed to chmod temp config file: %w", err)
	}

	if err := d.program.Execute("sudo", []string{}, "-u", "dd-agent", "bash", "-c",
		fmt.Sprintf("cat %s | tee %s > /dev/null", tmpFilePath, filePath)); err != nil {
		return "", fmt.Errorf("failed to write datadog config file %s: %w", filePath, err)
	}

	return newContent, nil
}

// prepareEnvs return envs the datadog
func (d *DatadogLinuxOperation) prepareEnvs(ddSite, ddApiKey string) []string {
	var envs []string
	envs = append(envs, fmt.Sprintf("DD_API_KEY=%s", ddApiKey))
	envs = append(envs, fmt.Sprintf("DD_SITE=%s", ddSite))
	return envs
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
		if strings.Contains(line, "Installed") || strings.Contains(line, "Candidate") {
			continue
		}
		if strings.Contains(line, version) {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				return fields[0], nil
			}
		}
	}

	return "", utils.ErrDatadogVersionNotFound()
}
