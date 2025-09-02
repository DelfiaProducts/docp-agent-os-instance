package tests

import (
	"testing"
	"time"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

func TestGetCollectInterval(t *testing.T) {
	bdd.Feature(t, "TestGetCollectInterval", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve obter o intervalo de coleta sem erro", func(s *bdd.Scenario) {
			var interval time.Duration
			var err error
			s.When("GetCollectInterval é chamado", func() {
				interval, err = utils.GetCollectInterval()
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "GetCollectInterval não deve retornar erro")
			})
			s.Then("intervalo deve ser maior que zero", func(t *testing.T) {
				if interval <= 0 {
					t.Errorf("intervalo esperado > 0, obtido %d", interval)
				}
			})
		})
	})
}

func TestGetStateCheckUrl(t *testing.T) {
	bdd.Feature(t, "TestGetStateCheckUrl", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve obter a URL de verificação de estado sem erro", func(s *bdd.Scenario) {
			var stateCheckUrl string
			var err error
			s.When("GetDomainUrl é chamado", func() {
				stateCheckUrl, err = utils.GetDomainUrl()
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "GetDomainUrl não deve retornar erro")
			})
			s.Then("URL deve ser diferente de vazio", func(t *testing.T) {
				if stateCheckUrl == "" {
					t.Errorf("URL esperada não deve ser vazia")
				}
			})
		})
	})
}

func TestGetPortAgentApi(t *testing.T) {
	bdd.Feature(t, "TestGetPortAgentApi", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve obter a porta da API do agente sem erro", func(s *bdd.Scenario) {
			var port string
			var err error
			s.When("GetPortAgentApi é chamado", func() {
				port, err = utils.GetPortAgentApi()
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "GetPortAgentApi não deve retornar erro")
			})
			bdd.Printf("port: %v\n", port)
		})
	})
}
