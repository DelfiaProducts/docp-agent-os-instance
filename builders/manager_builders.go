package builders

import (
	"github.com/OryaHub/agent-os-instance/agents"
	libinterfaces "github.com/OryaHub/agent-os-instance/libs/interfaces"
)

// ManagerBuilder return manager by distro
func ManagerBuilder() libinterfaces.IManager {
	return agents.NewManagerAgent(ManagerOperatorBuilder())
}
