package tests

import (
	"log"
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

func TestGetBinary(t *testing.T) {
	bdd.Feature(t, "TestGetBinary", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve baixar o binário sem erro", func(s *bdd.Scenario) {
			var resp any
			var statusCode int
			var err error
			s.Given("uma URL válida de binário", func() {})
			s.When("GetBinary é chamado", func() {
				url := "https://test-docp-agent-data.s3.amazonaws.com/manager/latest/linux_amd64"
				resp, statusCode, err = utils.GetBinary(url)
			})
			s.Then("não deve retornar erro e deve retornar resposta e statusCode", func(t *testing.T) {
				bdd.AssertNoError(t, err, "GetBinary não deve retornar erro")
				log.Printf("resp: %v\n", resp)
				log.Printf("statusCode: %d\n", statusCode)
			})
		})
	})
}
