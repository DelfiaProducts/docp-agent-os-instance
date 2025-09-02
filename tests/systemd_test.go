package tests

import (
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/pkg"
)

func TestNewSystemdClient(t *testing.T) {
	bdd.Feature(t, "TestNewSystemdClient", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve criar SystemdClient sem erro", func(s *bdd.Scenario) {
			var systemdClient *pkg.SystemdClient
			s.When("NewSystemdClient é chamado", func() {
				systemdClient = pkg.NewSystemdClient()
			})
			s.Then("SystemdClient não deve ser nil", func(t *testing.T) {
				if systemdClient == nil {
					t.Errorf("TestNewSystemdClient: expect(!nil) - got(%v)\n", systemdClient)
				}
			})
		})
	})
}

func TestSystemdClientStatus(t *testing.T) {
	bdd.Feature(t, "TestSystemdClientStatus", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve obter status do serviço sem erro", func(s *bdd.Scenario) {
			var systemdClient *pkg.SystemdClient
			var status interface{}
			var err error
			s.When("NewSystemdClient é chamado", func() {
				systemdClient = pkg.NewSystemdClient()
			})
			s.Given("SystemdClient válido", func() {
				if systemdClient == nil {
					t.Errorf("TestSystemdClientStatus: expect(!nil) - got(%v)\n", systemdClient)
				}
			})
			s.When("Status é chamado", func() {
				if systemdClient != nil {
					status, err = systemdClient.Status("datadog-agent.service")
				}
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "Status não deve retornar erro")
			})
			s.Then("status deve ser diferente de nil", func(t *testing.T) {
				if status == nil {
					t.Errorf("TestSystemdClientStatus: expect(!nil) - got(%v)\n", status)
				}
			})
		})
	})
}
