package tests

import (
	"log"
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/builders"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
)

func TestManagerBuilder(t *testing.T) {
	bdd.Feature(t, "TestManagerBuilder", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve criar um manager sem erro", func(s *bdd.Scenario) {
			var agent any
			s.Given("nenhum manager criado", func() {})
			s.When("ManagerBuilder é chamado", func() {
				agent = builders.ManagerBuilder()
			})
			s.Then("manager não deve ser nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, agent, "Manager deve ser diferente de nil")
				log.Printf("agent: %+v\n", agent)
			})
		})
	})
}
