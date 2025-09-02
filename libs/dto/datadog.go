package dto

// DatadogResponse is dto for response datadog
type DatadogResponse struct {
	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type DatadogEnvVars struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// DatadogInstallDTO is struct for payload the install datadog
type DatadogInstallDTO struct {
	DDSite    string           `json:"dd_site"`
	DDApiKey  string           `json:"dd_api_key"`
	Mode      string           `json:"mode"`
	Component string           `json:"component"`
	EnvVars   []DatadogEnvVars `json:"env_vars,omitempty"`
}
