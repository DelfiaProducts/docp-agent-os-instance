package dto

// AgentVersions struct for agent versions
type AgentVersions struct {
	LatestVersion string   `json:"latest"`
	Versions      []string `json:"versions"`
}

// DatadogApiGithubVersions struct for datadog api github versions
type DatadogApiGithubVersions struct {
	Name    string `json:"name"`
	TagName string `json:"tag_name"`
}
