package dto

const (
	ErrLevelHigh = iota
	ErrLevelMedium
	ErrLevelLow
)

type HealthResponse struct {
	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ManagerData is struct for manage data
type ManagerData struct{}

// ManagerStateAction is struct for actions
type ManagerStateAction struct {
	Type          string                    `json:"type"`
	Action        string                    `json:"action"`
	Mode          string                    `json:"mode"`
	Version       string                    `json:"version"`
	Component     string                    `json:"component"`
	ComponentEnvs []ManagerStateActionEnvs  `json:"component_envs,omitempty"`
	Envs          []ManagerStateActionEnvs  `json:"envs,omitempty"`
	Files         []ManagerStateActionFiles `json:"files,omitempty"`
}

// ManagerStateActionEnvs is struct for envs the actions
type ManagerStateActionEnvs struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// ManagerStateActionFiles is struct for files the actions
type ManagerStateActionFiles struct {
	FilePath string `json:"file_path,omitempty"`
	Content  string `json:"content,omitempty"`
}

type ManagerChanErrors struct {
	From     string
	Priority int
	Err      error
}
