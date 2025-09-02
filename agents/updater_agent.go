package agents

import (
	libinterfaces "github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
)

// UpdaterAgent is struct for updater agent
type UpdaterAgent struct {
	operator libinterfaces.IOperator
}

// NewUpdaterAgent return instance of updater agent
func NewUpdaterAgent(operator libinterfaces.IOperator) *UpdaterAgent {
	return &UpdaterAgent{operator: operator}
}

// Start execute running the updater
func (m *UpdaterAgent) Start() error {
	if err := m.operator.Run(); err != nil {
		return err
	}
	return nil
}
