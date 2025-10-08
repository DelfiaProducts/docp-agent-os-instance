package builders

import (
	"github.com/OryaHub/agent-os-instance/agents"
	libinterfaces "github.com/OryaHub/agent-os-instance/libs/interfaces"
)

// UpdaterBuilder return updater by distro
func UpdaterBuilder() libinterfaces.IUpdater {
	return agents.NewUpdaterAgent(UpdaterOperatorBuilder())
}
