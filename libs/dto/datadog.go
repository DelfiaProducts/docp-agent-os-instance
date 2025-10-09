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
	DDSite                        string           `json:"dd_site"`
	DDApiKey                      string           `json:"dd_api_key"`
	Mode                          string           `json:"mode"`
	Version                       string           `json:"version"`
	Component                     string           `json:"component"`
	DDApmInstrumentationLibraries string           `json:"dd_apm_instrumentation_libraries"`
	EnvVars                       []DatadogEnvVars `json:"env_vars,omitempty"`
}

// DatadogUpdateVersionDTO is struct for payload the update version datadog
type DatadogUpdateVersionDTO struct {
	Version string `json:"version"`
}

// DatadogYamlDTO is struct for dto the yaml file configuration datadog
type DatadogYamlDTO struct {
	ApiKey string `yaml:"api_key"`
	Site   string `yaml:"site"`
}
