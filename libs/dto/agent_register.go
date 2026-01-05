package dto

// AgentRegisterDataCreateOrUpdate is struct for data the create or update metadata
// in service agente register
type AgentRegisterDataCreateOrUpdate struct {
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
