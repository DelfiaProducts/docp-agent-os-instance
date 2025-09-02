package builders

import (
	libinterfaces "github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/DelfiaProducts/docp-agent-os-instance/operators"
)

// UpdaterOperatorBuilder return updater operator by distro
func UpdaterOperatorBuilder() libinterfaces.IOperator {
	return operators.NewUpdaterOperator()
}
