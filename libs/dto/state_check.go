package dto

// StateCheckResponse is struct for response the state check
type StateCheckResponse struct {
	Signal StateCheckSignal `json:"signal"`
}

// StateCheckSignal is struct for signal
type StateCheckSignal struct {
	TypeSignal         string           `json:"type"`
	Agents             StateCheckAgents `json:"agents"`
	Duration           string           `json:"duration"`
	RemoveOtherVendors []string         `json:"remove_other_vendors"`
}

// StateCheckAgents is struct for agents payload
type StateCheckAgents struct {
	DocpAgent               StateCheckDocpAgent               `json:"docp-agent"`
	DatadogAgent            StateCheckDatadogAgent            `json:"datadog-agent"`
	DatadogTracerLibrary    StateCheckDatadogTracerLibrary    `json:"datadog-tracer-library"`
	DatadogTracerSingleStep StateCheckDatadogTracerSingleStep `json:"datadog-tracer-single-step"`
}

// StateCheckDocpAgent is component for docp agents
type StateCheckDocpAgent struct {
	Version string `json:"version"`
}

// StateCheckDatadogAgent is component for datadog agent
type StateCheckDatadogAgent struct {
	Version        string                          `json:"version"`
	Enabled        *bool                           `json:"enabled,omitempty"`
	ApiKey         string                          `json:"api-key"`
	AppKey         string                          `json:"app-key"`
	Site           string                          `json:"site"`
	Configurations StateCheckDatadogConfigurations `json:"configurations,omitempty"`
}

// StateCheckDatadogConfigurations is struct for configurations
type StateCheckDatadogConfigurations struct {
	Files []StateCheckFiles `json:"files,omitempty"`
}

// StateCheckFiles is struct for files path
type StateCheckFiles struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
}

// StateCheckDatadogTracerLibrary is component for datadog tracer library
type StateCheckDatadogTracerLibrary struct {
	Version    string `json:"version"`
	Language   string `json:"language"`
	PathTracer string `json:"path_tracer"`
}

// StateCheckDatadogTracerSingleStep is component for datadog tracer single step
type StateCheckDatadogTracerSingleStep struct {
	Version            string              `json:"version"`
	InstallWithEnvVars []StateCheckEnvVars `json:"install_with_env_vars"`
}

type StateCheckEnvVars struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// CustomInstallVars is custom vars
type CustomInstallVars struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// ctxKey is type for key context
type ctxKey string

// ContextTransactionStatus is const for context transaction
const ContextTransactionStatus ctxKey = "transactionStatus"

// TransactionStatus is struct for transaction status
type TransactionStatus struct {
	ID        string `json:"id"`
	UlidEvent string `json:"ulid_event"`
	TypeEvent string `json:"type"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}
