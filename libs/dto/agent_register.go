package dto

// AgentRegisterData is struct for data the service agente register
type AgentRegisterData struct {
	ClientInfo ClientInfo `json:"client_info"`
	Metadata   Metadata   `json:"linux_metadata"`
}

// RegisterAgent is struct for agents
type RegisterAgent struct {
	ClientInfo ClientInfo `json:"client_info"`
	LinuxAgent LinuxAgent `json:"linux_agent"`
}

// AgentRegisterDataCreate is struct for data the create initial metadata
// in service agente register
type AgentRegisterDataCreate struct {
	NoGroupAssociation bool     `json:"no_group_association,omitempty"`
	Tags               []string `json:"tags"`
	VMName             string   `json:"vm_name"`
	Metadata           Metadata `json:"metadata"`
}

// AgentRegisterDataResponseSuccess is struct for response success the metadata
// in service agente register
type AgentRegisterDataResponseSuccess struct {
	AccessToken string `json:"access_token"`
}

// AgentRegisterDataResponseError is struct for response error from metadata
// in service agente register
type AgentRegisterDataResponseError struct {
	Status  string      `json:"status,omitempty"`
	Code    string      `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// AgentRegisterDataUpdate is struct for data the update metadata
// in service agente register
type AgentRegisterDataUpdate struct {
	VMName   string   `json:"vm_name"`
	Tags     []string `json:"tags"`
	Metadata Metadata `json:"metadata"`
}
