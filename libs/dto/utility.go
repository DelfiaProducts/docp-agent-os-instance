package dto

// AgentVersions struct for agent versions
type AgentVersions struct {
	LatestVersion string   `json:"latest"`
	Versions      []string `json:"versions"`
}
