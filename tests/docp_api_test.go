package tests

import (
	"os"
	"testing"

	"github.com/OryaHub/agent-os-instance/api"
	"github.com/OryaHub/agent-os-instance/libs/bdd"
	libutils "github.com/OryaHub/agent-os-instance/libs/utils"
)

func TestNewOryaApi(t *testing.T) {
	bdd.Feature(t, "TestNewOryaApi", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve criar OryaApi sem erro", func(s *bdd.Scenario) {
			var docApi *api.OryaApi
			s.When("NewOryaApi é chamado", func() {
				logger := libutils.NewOryaLoggerJSON(os.Stdout)
				docApi = api.NewOryaApi("3000", logger)
			})
			s.Then("OryaApi não deve ser nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, docApi, "OryaApi deve ser diferente de nil")
			})
		})
	})
}

func TestOryaApiSetup(t *testing.T) {
	bdd.Feature(t, "TestOryaApiSetup", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Setup não deve retornar erro", func(s *bdd.Scenario) {
			var docApi *api.OryaApi
			var err error
			s.When("NewOryaApi é chamado", func() {
				logger := libutils.NewOryaLoggerJSON(os.Stdout)
				docApi = api.NewOryaApi("3000", logger)
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

func TestOryaApiRun(t *testing.T) {
	bdd.Feature(t, "TestOryaApiRun", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Run não deve retornar erro", func(s *bdd.Scenario) {
			var docApi *api.OryaApi
			var err error
			s.When("NewOryaApi é chamado", func() {
				logger := libutils.NewOryaLoggerJSON(os.Stdout)
				docApi = api.NewOryaApi("3000", logger)
			})
			s.Given("OryaApi válido", func() {
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
