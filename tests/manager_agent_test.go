package tests

import (
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/agents"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/operators"
)

func TestManagerAgentStart(t *testing.T) {
	bdd.Feature(t, "TestManagerAgentStart", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve iniciar o ManagerAgent sem erro", func(s *bdd.Scenario) {
			var err error
			s.When("Start é chamado", func() {
				operator := operators.NewManagerOperator()
				managerAgent := agents.NewManagerAgent(operator)
				err = managerAgent.Start()
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				if err != nil {
					t.Errorf("TestManagerAgentStart: expect(nil) - got(%s)\n", err.Error())
				}
			})
		})
	})
}
