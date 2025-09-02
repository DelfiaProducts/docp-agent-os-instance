package builders

import (
	"github.com/DelfiaProducts/docp-agent-os-instance/agents"

	libinterfaces "github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
)

// AgentBuilder return agent by distro
func AgentBuilder() libinterfaces.IAgent {
	return agents.NewDocpAgent(AgentOperatorBuilder())
}
