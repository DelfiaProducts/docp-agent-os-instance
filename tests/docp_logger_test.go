package tests

import (
	"os"
	"testing"

	"github.com/OryaHub/agent-os-instance/libs/bdd"
	"github.com/OryaHub/agent-os-instance/libs/utils"
)

func TestNewOryaLoggerJSON(t *testing.T) {
	bdd.Feature(t, "TestNewOryaLoggerJSON", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve criar logger JSON sem erro", func(s *bdd.Scenario) {
			var oryaLogger any
			s.Given("um stdout válido", func() {})
			s.When("NewOryaLoggerJSON é chamado", func() {
				oryaLogger = utils.NewOryaLoggerJSON(os.Stdout)
			})
			s.Then("logger não deve ser nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, oryaLogger, "Logger deve ser diferente de nil")
			})
		})
	})
}

func TestOryaLoggerText(t *testing.T) {
	bdd.Feature(t, "TestOryaLoggerText", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve criar logger Text sem erro", func(s *bdd.Scenario) {
			var oryaLogger any
			s.Given("um stdout válido", func() {})
			s.When("NewOryaLoggerText é chamado", func() {
				oryaLogger = utils.NewOryaLoggerText(os.Stdout)
			})
			s.Then("logger não deve ser nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, oryaLogger, "Logger deve ser diferente de nil")
			})
		})
	})
}

func TestOryaLoggerLogs(t *testing.T) {
	bdd.Feature(t, "TestOryaLoggerLogs", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve logar mensagens nos dois formatos", func(s *bdd.Scenario) {
			var oryaLoggerJson, oryaLoggerText any
			s.Given("um logger JSON e um logger Text válidos", func() {
				oryaLoggerJson = utils.NewOryaLoggerJSON(os.Stdout)
				oryaLoggerText = utils.NewOryaLoggerText(os.Stdout)
			})
			s.When("chamo os métodos de log", func() {
				if oryaLoggerJson != nil {
					oryaLoggerJson.(*utils.OryaLogger).Info("test orya info", "api_key", "123rg389gh", "srv", "agent", "code", 123, "active", true)
					oryaLoggerJson.(*utils.OryaLogger).Warn("test orya warn", "api_key", "123rg389gh", "srv", "agent", "code", 123, "active", true)
					oryaLoggerJson.(*utils.OryaLogger).Error("test orya error", "api_key", "123rg389gh", "srv", "agent", "code", 123, "active", true)
				}
				if oryaLoggerText != nil {
					oryaLoggerText.(*utils.OryaLogger).Info("test orya info", "api_key", "123rg389gh", "srv", "agent", "code", 123, "active", true)
					oryaLoggerText.(*utils.OryaLogger).Warn("test orya warn", "api_key", "123rg389gh", "srv", "agent", "code", 123, "active", true)
					oryaLoggerText.(*utils.OryaLogger).Error("test orya error", "api_key", "123rg389gh", "srv", "agent", "code", 123, "active", true)
				}
			})
			s.Then("não deve ocorrer panic ao logar", func(t *testing.T) {
				// Se não ocorrer panic, considera sucesso
			})
		})
	})
}
