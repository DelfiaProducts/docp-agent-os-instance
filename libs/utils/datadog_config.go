package utils

import (
	"fmt"
	"regexp"
	"strings"
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

	// No tags section found — create it at the end of the file with newTags.
	if tagsStartIdx == -1 {
		if len(newTags) == 0 {
			return content, nil
		}
		block := []string{"tags:"}
		for _, t := range newTags {
			block = append(block, fmt.Sprintf(`  - "%s"`, t))
		}
		return content + "\n" + strings.Join(block, "\n"), nil
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

// hostnameRe matches a hostname line in datadog.yaml:
//
//	hostname: my-host              (active)
//	# hostname: my-host            (commented)
//	#hostname: my-host             (no space after #)
var hostnameRe = regexp.MustCompile(`^(#?)\s*hostname:\s*(.*)$`)

// ApplyHostnameInDatadogConfig sets or updates the hostname field in a
// datadog.yaml content string. Rules:
//   - If the hostname line is commented out, it is uncommented and updated.
//   - If the hostname line is active, it is overwritten with the new value.
//   - If no hostname line exists, one is appended at the end.
//   - The rest of the file is preserved byte-for-byte.
func ApplyHostnameInDatadogConfig(content string, hostname string) string {
	if hostname == "" {
		return content
	}

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if hostnameRe.MatchString(line) {
			lines[i] = fmt.Sprintf("hostname: %s", hostname)
			return strings.Join(lines, "\n")
		}
	}

	// Not found — append at the end.
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return content + fmt.Sprintf("hostname: %s\n", hostname)
}

// ExtractHostnameFromDatadogConfig parses a datadog.yaml content string and
// returns the value of the hostname field, or empty string if not found.
func ExtractHostnameFromDatadogConfig(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		matches := hostnameRe.FindStringSubmatch(line)
		if matches != nil && len(matches) >= 3 {
			// matches[1] is the comment marker (#), matches[2] is the value
			value := strings.TrimSpace(matches[2])
			if value != "" {
				// Strip surrounding quotes if present
				value = strings.Trim(value, `"'`)
				return value
			}
		}
	}
	return ""
}

// ExtractTagKey returns the key portion of a "key:value" Datadog tag.
// For tags that do not contain ":", the whole tag string is treated as the key.
func ExtractTagKey(tag string) string {
	if idx := strings.Index(tag, ":"); idx != -1 {
		return tag[:idx]
	}
	return tag
}

// simpleKeyRe builds a regex that matches an active or commented line like:
//
//	api_key: "xxx"       (active)
//	# api_key: "xxx"     (commented)
//	#api_key: "xxx"      (no space after #)
func simpleKeyRe(key string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`^(#?)\s*%s:\s*(.*)$`, regexp.QuoteMeta(key)))
}

// applySimpleConfigField is a generic helper that sets or updates a simple
// key: value field in a YAML content string. Rules:
//   - If the line is commented out, it is uncommented and updated.
//   - If the line is active, it is overwritten with the new value.
//   - If no such line exists, one is appended at the end.
//   - The rest of the file is preserved byte-for-byte.
func applySimpleConfigField(content, key, value string) string {
	if value == "" {
		return content
	}
	re := simpleKeyRe(key)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if re.MatchString(line) {
			lines[i] = fmt.Sprintf("%s: %s", key, value)
			return strings.Join(lines, "\n")
		}
	}
	// Not found — append at the end.
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return content + fmt.Sprintf("%s: %s\n", key, value)
}

// extractSimpleConfigField extracts the value of a simple key: value field
// from a YAML content string. Returns empty string if not found.
func extractSimpleConfigField(content, key string) string {
	re := simpleKeyRe(key)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if matches != nil && len(matches) >= 3 {
			value := strings.TrimSpace(matches[2])
			if value != "" {
				value = strings.Trim(value, `"'`)
				return value
			}
		}
	}
	return ""
}

// ApplyApiKeyInDatadogConfig sets or updates the api_key field in a
// datadog.yaml content string.
func ApplyApiKeyInDatadogConfig(content, apiKey string) string {
	return applySimpleConfigField(content, "api_key", apiKey)
}

// ApplyAppKeyInDatadogConfig sets or updates the app_key field in a
// datadog.yaml content string.
func ApplyAppKeyInDatadogConfig(content, appKey string) string {
	return applySimpleConfigField(content, "app_key", appKey)
}

// ApplySiteInDatadogConfig sets or updates the site field in a
// datadog.yaml content string.
func ApplySiteInDatadogConfig(content, site string) string {
	return applySimpleConfigField(content, "site", site)
}

// ExtractApiKeyFromDatadogConfig extracts the api_key value from a
// datadog.yaml content string.
func ExtractApiKeyFromDatadogConfig(content string) string {
	return extractSimpleConfigField(content, "api_key")
}

// ExtractAppKeyFromDatadogConfig extracts the app_key value from a
// datadog.yaml content string.
func ExtractAppKeyFromDatadogConfig(content string) string {
	return extractSimpleConfigField(content, "app_key")
}

// ExtractSiteFromDatadogConfig extracts the site value from a
// datadog.yaml content string.
func ExtractSiteFromDatadogConfig(content string) string {
	return extractSimpleConfigField(content, "site")
}
