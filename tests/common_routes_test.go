package tests

import (
	"os"
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/api"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	libutils "github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

func TestNewCommonRoutes(t *testing.T) {
	bdd.Feature(t, "TestNewCommonRoutes", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve criar CommonRoutes sem erro", func(s *bdd.Scenario) {
			var commonRouter any
			s.Given("um logger válido", func() {})
			s.When("NewCommonRoutes é chamado", func() {
				logger := libutils.NewDocpLoggerJSON(os.Stdout)
				commonRouter = api.NewCommonRoutes(logger)
			})
			s.Then("CommonRoutes não deve ser nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, commonRouter, "CommonRoutes deve ser diferente de nil")
			})
		})
	})
}

func TestCommonRoutesSetup(t *testing.T) {
	bdd.Feature(t, "TestCommonRoutesSetup", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Setup não deve retornar erro", func(s *bdd.Scenario) {
			var commonRouter any
			var err error
			s.Given("um logger válido", func() {})
			s.When("NewCommonRoutes é chamado", func() {
				logger := libutils.NewDocpLoggerJSON(os.Stdout)
				commonRouter = api.NewCommonRoutes(logger)
			})
			s.When("Setup é chamado", func() {
				if commonRouter != nil {
					err = commonRouter.(*api.CommonRoutes).Setup()
				}
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
		})
	})
}
