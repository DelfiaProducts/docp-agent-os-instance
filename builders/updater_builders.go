package builders

import (
	"github.com/DelfiaProducts/docp-agent-os-instance/agents"
	libinterfaces "github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
)

// UpdaterBuilder return updater by distro
func UpdaterBuilder() libinterfaces.IUpdater {
	return agents.NewUpdaterAgent(UpdaterOperatorBuilder())
}
