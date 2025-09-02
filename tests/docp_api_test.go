package tests

import (
	"os"
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/api"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	libutils "github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

func TestNewDocpApi(t *testing.T) {
	bdd.Feature(t, "TestNewDocpApi", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve criar DocpApi sem erro", func(s *bdd.Scenario) {
			var docApi *api.DocpApi
			s.When("NewDocpApi é chamado", func() {
				logger := libutils.NewDocpLoggerJSON(os.Stdout)
				docApi = api.NewDocpApi("3000", logger)
			})
			s.Then("DocpApi não deve ser nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, docApi, "DocpApi deve ser diferente de nil")
			})
		})
	})
}

func TestDocpApiSetup(t *testing.T) {
	bdd.Feature(t, "TestDocpApiSetup", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Setup não deve retornar erro", func(s *bdd.Scenario) {
			var docApi *api.DocpApi
			var err error
			s.When("NewDocpApi é chamado", func() {
				logger := libutils.NewDocpLoggerJSON(os.Stdout)
				docApi = api.NewDocpApi("3000", logger)
			})
			s.When("Setup é chamado", func() {
				if docApi != nil {
					err = docApi.Setup()
				}
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
		})
	})
}

func TestDocpApiRun(t *testing.T) {
	bdd.Feature(t, "TestDocpApiRun", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Run não deve retornar erro", func(s *bdd.Scenario) {
			var docApi *api.DocpApi
			var err error
			s.When("NewDocpApi é chamado", func() {
				logger := libutils.NewDocpLoggerJSON(os.Stdout)
				docApi = api.NewDocpApi("3000", logger)
			})
			s.Given("DocpApi válido", func() {
				if docApi != nil {
					err = docApi.Setup()
				}
			})
			s.When("Run é chamado", func() {
				if docApi != nil {
					err = docApi.Run()
				}
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "Run não deve retornar erro")
			})
		})
	})
}
