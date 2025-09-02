package tests

import (
	"os"
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/controllers"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

func TestNewDatadogHttpController(t *testing.T) {
	bdd.Feature(t, "TestNewDatadogHttpController", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve criar DatadogHttpController sem erro", func(s *bdd.Scenario) {
			var datadogController *controllers.DatadogHttpController
			var logger interfaces.ILogger
			s.When("logger é criado", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.When("NewDatadogHttpController é chamado", func() {
				datadogController = controllers.NewDatadogHttpController(logger)
			})
			s.Then("DatadogHttpController não deve ser nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, datadogController, "DatadogHttpController deve ser diferente de nil")
			})
		})
	})
}
