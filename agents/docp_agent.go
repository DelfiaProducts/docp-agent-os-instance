package agents

import (
	libinterfaces "github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
)

// DocpAgent is struct for agent docp
type DocpAgent struct {
	operator libinterfaces.IOperator
}

// NewDocpAgent return instance of docp agent
func NewDocpAgent(operator libinterfaces.IOperator) *DocpAgent {
	return &DocpAgent{operator: operator}
}

// Start execute running the agent
func (d *DocpAgent) Start() error {
	if err := d.operator.Run(); err != nil {
		return err
	}
	return nil
}
