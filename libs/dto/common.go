package dto

const (
	ErrLevelHigh = iota
	ErrLevelMedium
	ErrLevelLow
)

// Metadata is struct for metadata the host
type Metadata struct {
	ComputeInfo  ComputeInfo   `json:"compute_info"`
	CPUInfo      []CPUInfo     `json:"cpus_info"`
	MemoryInfo   MemoryInfo    `json:"memory_info"`
	DiskInfo     DiskInfo      `json:"disk_info"`
	ProcessInfos []ProcessInfo `json:"process_infos"`
}

// LinuxAgent is struct for agent
type LinuxAgent struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// ConfigAgent is struct for config file agent
type ConfigAgent struct {
	Version            string   `yaml:"version"`
	RollbackVersion    string   `yaml:"rollback_version"`
	AlreadyCreated     bool     `yaml:"already_created,omitempty"`
	AlreadyTracer      bool     `yaml:"already_tracer"`
	TracerLanguages    []string `yaml:"tracer_languages"`
	NoGroupAssociation bool     `yaml:"no_group_association,omitempty"`
	Agent              Agent    `yaml:"agent"`
	AccessToken        string   `json:"access_token"`
	ComputeId          string   `json:"compute_id"`
	DocpOrgId          int      `json:"docp_org_id"`
}

// Agent is struct for config file agent
type Agent struct {
	ApiKey string                 `yaml:"apiKey"`
	Tags   map[string]interface{} `yaml:"tags"`
}

// StateAction is struct for actions
type StateAction struct {
	Type          string             `json:"type"`
	Action        string             `json:"action"`
	Version       string             `json:"version"`
	Mode          string             `json:"mode,omitempty"`
	Component     string             `json:"component"`
	ComponentEnvs []StateActionEnvs  `json:"component_envs,omitempty"`
	Envs          []StateActionEnvs  `json:"envs,omitempty"`
	Files         []StateActionFiles `json:"files,omitempty"`
}

// StateActionEnvs is struct for envs the actions
type StateActionEnvs struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// StateActionFiles is struct for files the actions
type StateActionFiles struct {
	FilePath string `json:"file_path,omitempty"`
	Content  string `json:"content,omitempty"`
}

// AuthTokenClaims is struct for auth token claims
type AuthTokenClaims struct {
	DocpOrgId int    `json:"docp_org_id"`
	ComputeId string `json:"compute_id"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type CommonChanErrors struct {
	From     string
	Priority int
	Err      error
}
