package builders

import (
	libinterfaces "github.com/OryaHub/agent-os-instance/libs/interfaces"
	"github.com/OryaHub/agent-os-instance/operators"
)

// AgentOperatorBuilder return operator agent by distro
func AgentOperatorBuilder() libinterfaces.IOperator {
	return operators.NewAgentOperator()
}
