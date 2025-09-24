package tests

import (
	"testing"

	"github.com/OryaHub/agent-os-instance/agents"
	"github.com/OryaHub/agent-os-instance/libs/bdd"
	"github.com/OryaHub/agent-os-instance/operators"
)

func TestNewOryaAgent(t *testing.T) {
	bdd.Feature(t, "TestNewOryaAgent", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve criar OryaAgent sem erro", func(s *bdd.Scenario) {
			var operator *operators.AgentOperator
			var agent *agents.OryaAgent
			s.When("operator é criado", func() {
				operator = operators.NewAgentOperator()
			})
			s.When("NewOryaAgent é chamado", func() {
				agent = agents.NewOryaAgent(operator)
			})
			s.Then("OryaAgent não deve ser nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, agent, "OryaAgent deve ser diferente de nil")
			})
		})
	})
}
