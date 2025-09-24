package builders

import (
	libinterfaces "github.com/OryaHub/agent-os-instance/libs/interfaces"
	"github.com/OryaHub/agent-os-instance/operators"
)

// ManagerOperatorBuilder return operator agent by distro
func ManagerOperatorBuilder() libinterfaces.IOperator {
	return operators.NewManagerOperator()
}
