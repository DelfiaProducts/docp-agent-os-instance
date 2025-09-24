package builders

import (
	"github.com/OryaHub/agent-os-instance/agents"

	libinterfaces "github.com/OryaHub/agent-os-instance/libs/interfaces"
)

// AgentBuilder return agent by distro
func AgentBuilder() libinterfaces.IAgent {
	return agents.NewDocpAgent(AgentOperatorBuilder())
}
