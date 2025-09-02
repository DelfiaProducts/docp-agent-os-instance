package tests

import (
	"log"
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/services"
)

func TestNewStateCheckService(t *testing.T) {
	bdd.Feature(t, "TestNewStateCheckService", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve criar StateCheckService sem erro", func(s *bdd.Scenario) {
			var stateCheck any
			s.Given("nenhum service criado", func() {})
			s.When("NewStateCheckService é chamado", func() {
				stateCheck = services.NewStateCheckService(logger)
			})
			s.Then("service não deve ser nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, stateCheck, "Service deve ser diferente de nil")
			})
		})
	})
}

func TestStateCheckServiceSetup(t *testing.T) {
	bdd.Feature(t, "TestStateCheckServiceSetup", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Setup não deve retornar erro", func(s *bdd.Scenario) {
			var stateCheck *services.StateCheckService
			s.Given("um StateCheckService válido", func() {
				stateCheck = services.NewStateCheckService(logger)
			})
			s.When("Setup é chamado", func() {
				err := stateCheck.Setup()
				s.Then("não deve retornar erro", func(t *testing.T) {
					bdd.AssertNoError(t, err, "Setup não deve retornar erro")
				})
			})
		})
	})
}

func TestStateCheckServicePreparePayload(t *testing.T) {
	bdd.Feature(t, "TestStateCheckServicePreparePayload", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("PreparePayload não deve retornar erro", func(s *bdd.Scenario) {
			var stateCheck *services.StateCheckService
			var payload []byte
			var apiKey string
			var err error
			s.Given("um StateCheckService válido e setup executado", func() {
				stateCheck = services.NewStateCheckService(logger)
				err = stateCheck.Setup()
			})
			s.When("PreparePayload é chamado", func() {
				payload, apiKey, err = stateCheck.PreparePayload()
			})
			s.Then("não deve retornar erro e deve retornar payload e apiKey", func(t *testing.T) {
				bdd.AssertNoError(t, err, "PreparePayload não deve retornar erro")
				log.Printf("payload: %s\napiKey: %s\n", string(payload), apiKey)
			})
		})
	})
}

func TestStateCheckServiceGetState(t *testing.T) {
	bdd.Feature(t, "TestStateCheckServiceGetState", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("GetState não deve retornar erro", func(s *bdd.Scenario) {
			var stateCheck *services.StateCheckService
			var result []byte
			var statusCode int
			var err error
			s.Given("um StateCheckService válido e setup executado", func() {
				stateCheck = services.NewStateCheckService(logger)
				err = stateCheck.Setup()
			})
			s.When("GetState é chamado", func() {
				result, statusCode, err = stateCheck.GetState()
			})
			s.Then("não deve retornar erro e deve retornar resultado e statusCode", func(t *testing.T) {
				bdd.AssertNoError(t, err, "GetState não deve retornar erro")
				log.Printf("result: %s\n", string(result))
				log.Printf("statusCode: %d\n", statusCode)
			})
		})
	})
}
