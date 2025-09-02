package tests

import (
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/operators"
)

func TestNewAgentOperator(t *testing.T) {
	bdd.Feature(t, "Criar novo AgentOperator", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Operador não deve ser nulo", func(s *bdd.Scenario) {
			s.Given("um AgentOperator é criado", func() {
				linux := operators.NewAgentOperator()
				s.Then("o operador não deve ser nulo", func(t *testing.T) {
					if linux == nil {
						t.Errorf("Esperado operador não nulo")
					}
				})
			})
		})
	})
}

func TestAgentOperatorSetup(t *testing.T) {
	bdd.Feature(t, "Setup do AgentOperator", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Setup não deve retornar erro", func(s *bdd.Scenario) {
			var linux *operators.AgentOperator
			s.Given("um AgentOperator válido", func() {
				linux = operators.NewAgentOperator()
			})
			s.When("Setup é chamado", func() {
				err := linux.Setup()
				s.Then("não deve retornar erro", func(t *testing.T) {
					if err != nil {
						t.Errorf("Esperado erro nulo, obtido: %v", err)
					}
				})
			})
		})
	})
}

func TestFeatureAgentOperatorRun(t *testing.T) {
	bdd.Feature(t, "Executar Run do AgentOperator", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Run e Setup não devem retornar erro", func(s *bdd.Scenario) {
			var linux *operators.AgentOperator
			s.Given("um AgentOperator válido", func() {
				linux = operators.NewAgentOperator()
			})
			s.When("Setup e Run são chamados", func() {
				errSetup := linux.Setup()
				errRun := linux.Run()
				s.Then("nenhum dos métodos deve retornar erro", func(t *testing.T) {
					if errSetup != nil {
						t.Errorf("Esperado erro nulo no Setup, obtido: %v", errSetup)
					}
					if errRun != nil {
						t.Errorf("Esperado erro nulo no Run, obtido: %v", errRun)
					}
				})
			})
		})
	})
}
