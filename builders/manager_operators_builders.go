package builders

import (
	libinterfaces "github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/DelfiaProducts/docp-agent-os-instance/operators"
)

// ManagerOperatorBuilder return operator agent by distro
func ManagerOperatorBuilder() libinterfaces.IOperator {
	return operators.NewManagerOperator()
}
