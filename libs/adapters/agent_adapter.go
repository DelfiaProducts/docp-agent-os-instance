package adapters

import (
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/components"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
)

// AgentAdapter is struct for adapter agent
type AgentAdapter struct {
	logger      interfaces.ILogger
	osOperation interfaces.IOSOperation
}

// NewAgentAdapter return instance of agent adapter
func NewAgentAdapter(logger interfaces.ILogger) *AgentAdapter {
	return &AgentAdapter{
		logger: logger,
	}
}

// Prepare configure agent adapter
func (a *AgentAdapter) Prepare() error {
	osOperation, err := components.SystemOperation(a.logger)
	if err != nil {
		return err
	}
	if err := osOperation.Setup(); err != nil {
		return err
	}
	a.osOperation = osOperation
	return nil
}

// HandlerSCMManager execute handler for scm manager
func (a *AgentAdapter) HandlerSCMManager() error {
	if err := a.osOperation.Execute(); err != nil {
		return err
	}
	return nil
}
