package builders

import (
	"github.com/DelfiaProducts/docp-agent-os-instance/agents"
	libinterfaces "github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
)

// ManagerBuilder return manager by distro
func ManagerBuilder() libinterfaces.IManager {
	return agents.NewManagerAgent(ManagerOperatorBuilder())
}
