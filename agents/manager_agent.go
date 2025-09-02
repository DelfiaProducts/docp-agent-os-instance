package agents

import (
	libinterfaces "github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
)

// ManagerAgent is struct for manager agent
type ManagerAgent struct {
	operator libinterfaces.IOperator
}

// NewManagerAgent return instance of manager agent
func NewManagerAgent(operator libinterfaces.IOperator) *ManagerAgent {
	return &ManagerAgent{operator: operator}
}

// Start execute running the manager
func (m *ManagerAgent) Start() error {
	if err := m.operator.Run(); err != nil {
		return err
	}
	return nil
}
