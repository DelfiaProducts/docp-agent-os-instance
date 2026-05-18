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

// MergeTagsInDatadogConfig parses the datadog-agent.yaml content and merges the
// existing host_tags section with newTags. Rules:
//   - If the tags section is active (uncommented), existing tags take priority:
//     a tag with the same key already in the file is kept and the incoming one is discarded.
//   - If the tags section is commented out, it is treated as absent: newTags are used as-is.
//   - The tags section is always written back uncommented.
//   - The rest of the file is preserved byte-for-byte.
//   - Both YAML tag formats are supported:
//     inline  →  tags: ["key1:val1", "key2:val2"]
//     block   →  tags:\n  - "key1:val1"\n  - "key2:val2"
func MergeTagsInDatadogConfig(content string, newTags []string) (string, error) {
	lines := strings.Split(content, "\n")

	// Patterns that match both the commented and active variants of the tags header.
	// Inline  example: tags: ["env:prod"]  or  # tags: ["env:prod"]
	// Block   example: tags:               or  # tags:
	inlineTagsRe := regexp.MustCompile(`^#?\s*tags:\s*\[`)
	blockTagsRe := regexp.MustCompile(`^#?\s*tags:\s*$`)
	// A YAML list item, with or without a leading comment marker.
	// Example active:    - "env:prod"
	// Example commented: #   - "env:prod"
	blockItemRe := regexp.MustCompile(`^#?\s*-\s+\S`)

	tagsStartIdx := -1
	tagsEndIdx := -1
	isInline := false
	existingTags := []string{}

	for i, line := range lines {
		stripped := strings.TrimSpace(line)
		if stripped == "" {
			continue
		}

		if inlineTagsRe.MatchString(stripped) {
			tagsStartIdx = i
			tagsEndIdx = i
			isInline = true
			// Only collect existing tags when the section is active (not commented).
			if !strings.HasPrefix(stripped, "#") {
				existingTags = ParseInlineTagLine(line)
			}
			break
		}

		if blockTagsRe.MatchString(stripped) {
			tagsStartIdx = i
			isInline = false
			sectionCommented := strings.HasPrefix(stripped, "#")
			j := i + 1
			for j < len(lines) {
				nextStripped := strings.TrimSpace(lines[j])
				// An empty line or a line that is not a list item terminates the block.
				if nextStripped == "" || !blockItemRe.MatchString(nextStripped) {
					break
				}
				// Only collect existing tags when the section is active (not commented).
				if !sectionCommented {
					tag := ParseBlockTagItemLine(lines[j])
					if tag != "" {
						existingTags = append(existingTags, tag)
					}
				}
				j++
			}
			// tagsEndIdx points at the last line that belongs to the tags section.
			// When there are no items j == i+1, so tagsEndIdx == i (the header itself).
			tagsEndIdx = j - 1
			break
		}
	}

	// No tags section found — nothing to change.
	if tagsStartIdx == -1 {
		return content, nil
	}

	// Merge: existing tags win on key conflict only when the section was active.
	mergedTags := MergeTagsByKeyPriority(existingTags, newTags)

	// Build the replacement lines (always uncommented).
	var replacement []string
	if isInline {
		quoted := make([]string, len(mergedTags))
		for i, t := range mergedTags {
			quoted[i] = fmt.Sprintf(`"%s"`, t)
		}
		replacement = []string{fmt.Sprintf("tags: [%s]", strings.Join(quoted, ", "))}
	} else {
		replacement = []string{"tags:"}
		for _, t := range mergedTags {
			replacement = append(replacement, fmt.Sprintf(`  - "%s"`, t))
		}
	}

	// Rebuild the file replacing only the tags section.
	result := make([]string, 0, len(lines))
	result = append(result, lines[:tagsStartIdx]...)
	result = append(result, replacement...)
	result = append(result, lines[tagsEndIdx+1:]...)

	return strings.Join(result, "\n"), nil
}

// ParseInlineTagLine extracts the tag strings from an inline tags line.
// Handles both active and commented variants:
//
//	tags: ["env:prod", "service:web"]
//	# tags: ["env:prod", "service:web"]
func ParseInlineTagLine(line string) []string {
	start := strings.Index(line, "[")
	end := strings.LastIndex(line, "]")
	if start == -1 || end == -1 || end <= start {
		return nil
	}
	inner := line[start+1 : end]
	if strings.TrimSpace(inner) == "" {
		return nil
	}
	parts := strings.Split(inner, ",")
	tags := make([]string, 0, len(parts))
	for _, p := range parts {
		tag := strings.Trim(strings.TrimSpace(p), `"'`)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

// ParseBlockTagItemLine extracts the tag value from a YAML block list item.
// Handles both active and commented variants:
//
//   - "env:prod"
//     #   - "env:prod"
func ParseBlockTagItemLine(line string) string {
	stripped := strings.TrimSpace(line)
	// Remove leading comment marker.
	if strings.HasPrefix(stripped, "#") {
		stripped = strings.TrimSpace(stripped[1:])
	}
	// Remove the YAML list marker "- ".
	if idx := strings.Index(stripped, "- "); idx == 0 {
		stripped = strings.TrimSpace(stripped[2:])
	}
	return strings.Trim(stripped, `"'`)
}

// MergeTagsByKeyPriority merges two tag slices. Each tag follows the Datadog
// "key:value" convention. When an existing tag and an incoming tag share the
// same key, the existing tag is kept and the incoming one is discarded.
func MergeTagsByKeyPriority(existing, incoming []string) []string {
	existingKeys := make(map[string]bool, len(existing))
	for _, tag := range existing {
		existingKeys[ExtractTagKey(tag)] = true
	}

	merged := make([]string, len(existing))
	copy(merged, existing)

	for _, tag := range incoming {
		key := ExtractTagKey(tag)
		if !existingKeys[key] {
			merged = append(merged, tag)
			existingKeys[key] = true
		}
	}
	return merged
}

// ExtractTagKey returns the key portion of a "key:value" Datadog tag.
// For tags that do not contain ":", the whole tag string is treated as the key.
func ExtractTagKey(tag string) string {
	if idx := strings.Index(tag, ":"); idx != -1 {
		return tag[:idx]
	}
	return tag
}

// applyHostTags reads the current content of filePath as dd-agent, merges
// the existing tags with newTags (existing tags win on key conflict), writes
// the result back as dd-agent using a temp file, and returns the new content.
// It does NOT restart the service — that is the caller's responsibility.
func (d *DatadogLinuxOperation) applyHostTags(filePath string, newTags []string) (string, error) {
	currentContent, err := d.program.ExecuteWithOutput("sudo", []string{}, "-u", "dd-agent", "cat", filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read datadog config file %s: %w", filePath, err)
	}

	newContent, err := MergeTagsInDatadogConfig(currentContent, newTags)
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
