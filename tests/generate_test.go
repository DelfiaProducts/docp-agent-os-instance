package tests

import (
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

func TestGenerateUniqHash(t *testing.T) {
	bdd.Feature(t, "TestGenerateUniqHash", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve gerar hash único sem erro", func(s *bdd.Scenario) {
			var hash string
			s.When("GenerateUniqHash é chamado", func() {
				hash = utils.GenerateUniqHash()
			})
			s.Then("hash não deve ser vazio", func(t *testing.T) {
				if len(hash) == 0 {
					t.Errorf("TestGenerateUniqHash: expect(!nil) - got(%s)\n", hash)
				}
			})
		})
	})
}
