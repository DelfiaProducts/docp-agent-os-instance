package builders

import (
	libinterfaces "github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/DelfiaProducts/docp-agent-os-instance/operators"
)

// AgentOperatorBuilder return operator agent by distro
func AgentOperatorBuilder() libinterfaces.IOperator {
	return operators.NewAgentOperator()
}
