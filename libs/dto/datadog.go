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

type DatadogInfos struct {
	HostId                        string `yaml:"hostId" json:"host_id"`
	Hostname                      string `yaml:"hostname" json:"hostname"`
	KernelArch                    string `yaml:"kernelArch" json:"kernel_arch"`
	KernelVersion                 string `yaml:"kernelVersion" json:"kernel_version"`
	Os                            string `yaml:"os" json:"os"`
	Platform                      string `yaml:"platform" json:"platform"`
	PlatformFamily                string `yaml:"platformFamily" json:"platform_family"`
	PlatformVersion               string `yaml:"platformVersion" json:"platform_version"`
	AgentVersion                  string `yaml:"agent_version" json:"agent_version"`
	Flavor                        string `yaml:"flavor" json:"flavor"`
	InfrastructureMode            string `yaml:"infrastructure_mode" json:"infrastructure_mode"`
	InstallMethodInstallerVersion string `yaml:"install_method_installer_version" json:"install_method_installer_version"`
	InstallMethodTool             string `yaml:"install_method_tool" json:"install_method_tool"`
	InstallMethodToolVersion      string `yaml:"install_method_tool_version" json:"install_method_tool_version"`
}

type DatadogConfigDTO struct {
	HostTags []string `json:"host_tags"`
	Hostname string   `json:"hostname,omitempty"`
}
