package tests

import (
	"log"
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/builders"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
)

func TestAgentBuilder(t *testing.T) {
	bdd.Feature(t, "TestAgentBuilder", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve criar um agent sem erro", func(s *bdd.Scenario) {
			var agent any
			s.Given("nenhum agente criado", func() {})
			s.When("AgentBuilder é chamado", func() {
				agent = builders.AgentBuilder()
			})
			s.Then("agent não deve ser nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, agent, "Agent deve ser diferente de nil")
				log.Printf("agent: %+v\n", agent)
			})
		})
	})
}
