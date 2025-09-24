package builders

import (
	libinterfaces "github.com/OryaHub/agent-os-instance/libs/interfaces"
	"github.com/OryaHub/agent-os-instance/operators"
)

// UpdaterOperatorBuilder return updater operator by distro
func UpdaterOperatorBuilder() libinterfaces.IOperator {
	return operators.NewUpdaterOperator()
}
