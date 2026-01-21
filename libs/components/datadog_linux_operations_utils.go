package components

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/OryaHub/agent-os-instance/libs/utils"
)

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

func (d *DatadogLinuxOperation) parseValueByPrefix(output string, prefix string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, prefix) {
			value := strings.TrimPrefix(trimmedLine, prefix)
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func (d *DatadogLinuxOperation) removeLinesByPrefix(prefixs []string, output string) string {
	lines := strings.Split(output, "\n")
	var filteredLines []string
	for _, line := range lines {
		for _, prefix := range prefixs {
			if !strings.Contains(line, prefix) {
				filteredLines = append(filteredLines, line)
			}
		}
	}
	output = strings.Join(filteredLines, "\n")
	return output
}
