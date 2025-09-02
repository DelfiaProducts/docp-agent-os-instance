package tests

import (
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/controllers"
)

func TestNewCommonHttpController(t *testing.T) {
	bdd.Feature(t, "TestNewCommonHttpController", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve criar CommonHttpController sem erro", func(s *bdd.Scenario) {
			var commonController any
			s.Given("nenhum controller criado", func() {})
			s.When("NewCommonHttpController é chamado", func() {
				commonController = controllers.NewCommonHttpController(logger)
			})
			s.Then("controller não deve ser nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, commonController, "CommonHttpController deve ser diferente de nil")
			})
		})
	})
}
