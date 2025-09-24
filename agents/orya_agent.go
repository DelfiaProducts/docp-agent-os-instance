package agents

import (
	libinterfaces "github.com/OryaHub/agent-os-instance/libs/interfaces"
)

// OryaAgent is struct for agent orya
type OryaAgent struct {
	operator libinterfaces.IOperator
}

// NewOryaAgent return instance of orya agent
func NewOryaAgent(operator libinterfaces.IOperator) *OryaAgent {
	return &OryaAgent{operator: operator}
}

// Start execute running the agent
func (d *OryaAgent) Start() error {
	if err := d.operator.Run(); err != nil {
		return err
	}
	return nil
}
