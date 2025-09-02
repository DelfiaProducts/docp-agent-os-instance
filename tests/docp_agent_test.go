package tests

import (
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/agents"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/operators"
)

func TestNewDocpAgent(t *testing.T) {
	bdd.Feature(t, "TestNewDocpAgent", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve criar DocpAgent sem erro", func(s *bdd.Scenario) {
			var operator *operators.AgentOperator
			var agent *agents.DocpAgent
			s.When("operator é criado", func() {
				operator = operators.NewAgentOperator()
			})
			s.When("NewDocpAgent é chamado", func() {
				agent = agents.NewDocpAgent(operator)
			})
			s.Then("DocpAgent não deve ser nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, agent, "DocpAgent deve ser diferente de nil")
			})
		})
	})
}
