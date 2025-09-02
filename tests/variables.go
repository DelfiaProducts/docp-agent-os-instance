package tests

import (
	"os"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

var logger = utils.NewDocpLoggerJSON(os.Stdout)
