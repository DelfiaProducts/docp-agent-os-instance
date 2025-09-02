package tests

import (
	"os"
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"

	"github.com/DelfiaProducts/docp-agent-os-instance/api"
)

func TestNewDatadogRoutes(t *testing.T) {

	bdd.Feature(t, "TestNewDatadogRoutes", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve criar DatadogRoutes sem erro", func(s *bdd.Scenario) {
			var datadogRouter *api.DatadogRoutes
			var logger interfaces.ILogger
			s.When("logger é criado", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.When("NewDatadogRoutes é chamado", func() {
				datadogRouter = api.NewDatadogRoutes(logger)
			})
			s.Then("DatadogRoutes não deve ser nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, datadogRouter, "DatadogRoutes deve ser diferente de nil")
			})
		})
	})
}

func TestDatadogRoutesSetup(t *testing.T) {
	bdd.Feature(t, "TestDatadogRoutesSetup", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Setup não deve retornar erro", func(s *bdd.Scenario) {
			var datadogRouter *api.DatadogRoutes
			var logger interfaces.ILogger
			var err error
			s.When("logger é criado", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.Given("DatadogRoutes válido", func() {
				datadogRouter = api.NewDatadogRoutes(logger)
			})
			s.When("Setup é chamado", func() {
				if datadogRouter != nil {
					err = datadogRouter.Setup()
				}
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
		})
	})
}
