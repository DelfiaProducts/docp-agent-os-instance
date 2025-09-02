package tests

import (
	"io"
	"log"
	"os"
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/dto"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/pkg"
)

func TestNewYmlClient(t *testing.T) {
	bdd.Feature(t, "TestNewYmlClient", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve criar YmlClient sem erro", func(s *bdd.Scenario) {
			var ymlClient any
			s.Given("nenhum YmlClient criado", func() {})
			s.When("NewYmlClient é chamado", func() {
				ymlClient = pkg.NewYmlClient()
			})
			s.Then("YmlClient não deve ser nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, ymlClient, "YmlClient deve ser diferente de nil")
			})
		})
	})
}

func TestYmlClientUnmarshal(t *testing.T) {
	bdd.Feature(t, "TestYmlClientUnmarshal", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve fazer unmarshal sem erro", func(s *bdd.Scenario) {
			var ymlClient *pkg.YmlClient
			var config dto.ConfigAgent
			var content []byte
			var err error
			s.Given("um YmlClient válido", func() {
				ymlClient = pkg.NewYmlClient()
			})
			s.When("abro o arquivo de config e leio o conteúdo", func() {
				f, errOpen := os.OpenFile("/opt/docp-agent/config.yml", os.O_RDONLY, os.ModePerm)
				if errOpen != nil {
					err = errOpen
					return
				}
				defer f.Close()
				content, err = io.ReadAll(f)
			})
			s.When("faço o unmarshal", func() {
				if err == nil {
					err = ymlClient.Unmarshall(content, &config)
				}
			})
			s.Then("não deve retornar erro e deve popular config", func(t *testing.T) {
				bdd.AssertNoError(t, err, "Unmarshall não deve retornar erro")
				log.Printf("config: %+v\n", config)
			})
		})
	})
}

func TestYmlClientMarshal(t *testing.T) {
	bdd.Feature(t, "TestYmlClientMarshal", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve fazer marshal sem erro", func(s *bdd.Scenario) {
			var ymlClient *pkg.YmlClient
			var data []byte
			var err error
			s.Given("um YmlClient válido", func() {
				ymlClient = pkg.NewYmlClient()
			})
			s.When("faço marshal de um config válido", func() {
				config := dto.ConfigAgent{
					Version: "0.0.1",
					Agent: dto.Agent{
						ApiKey: "xptom",
						Tags: map[string]interface{}{
							"group": "dev",
							"app":   1,
							"host":  "machine-1",
						},
					},
				}
				data, err = ymlClient.Marshall(&config)
			})
			s.Then("não deve retornar erro e deve retornar dados", func(t *testing.T) {
				bdd.AssertNoError(t, err, "Marshall não deve retornar erro")
				log.Printf("data: %+v\n", string(data))
			})
		})
	})
}
