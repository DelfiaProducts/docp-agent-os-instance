package tests

import (
	"os"
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

func TestNewDocpLoggerJSON(t *testing.T) {
	bdd.Feature(t, "TestNewDocpLoggerJSON", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve criar logger JSON sem erro", func(s *bdd.Scenario) {
			var docpLogger any
			s.Given("um stdout válido", func() {})
			s.When("NewDocpLoggerJSON é chamado", func() {
				docpLogger = utils.NewDocpLoggerJSON(os.Stdout)
			})
			s.Then("logger não deve ser nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, docpLogger, "Logger deve ser diferente de nil")
			})
		})
	})
}

func TestDocpLoggerText(t *testing.T) {
	bdd.Feature(t, "TestDocpLoggerText", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve criar logger Text sem erro", func(s *bdd.Scenario) {
			var docpLogger any
			s.Given("um stdout válido", func() {})
			s.When("NewDocpLoggerText é chamado", func() {
				docpLogger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.Then("logger não deve ser nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, docpLogger, "Logger deve ser diferente de nil")
			})
		})
	})
}

func TestDocpLoggerLogs(t *testing.T) {
	bdd.Feature(t, "TestDocpLoggerLogs", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve logar mensagens nos dois formatos", func(s *bdd.Scenario) {
			var docpLoggerJson, docpLoggerText any
			s.Given("um logger JSON e um logger Text válidos", func() {
				docpLoggerJson = utils.NewDocpLoggerJSON(os.Stdout)
				docpLoggerText = utils.NewDocpLoggerText(os.Stdout)
			})
			s.When("chamo os métodos de log", func() {
				if docpLoggerJson != nil {
					docpLoggerJson.(*utils.DocpLogger).Info("test docp info", "api_key", "123rg389gh", "srv", "agent", "code", 123, "active", true)
					docpLoggerJson.(*utils.DocpLogger).Warn("test docp warn", "api_key", "123rg389gh", "srv", "agent", "code", 123, "active", true)
					docpLoggerJson.(*utils.DocpLogger).Error("test docp error", "api_key", "123rg389gh", "srv", "agent", "code", 123, "active", true)
				}
				if docpLoggerText != nil {
					docpLoggerText.(*utils.DocpLogger).Info("test docp info", "api_key", "123rg389gh", "srv", "agent", "code", 123, "active", true)
					docpLoggerText.(*utils.DocpLogger).Warn("test docp warn", "api_key", "123rg389gh", "srv", "agent", "code", 123, "active", true)
					docpLoggerText.(*utils.DocpLogger).Error("test docp error", "api_key", "123rg389gh", "srv", "agent", "code", 123, "active", true)
				}
			})
			s.Then("não deve ocorrer panic ao logar", func(t *testing.T) {
				// Se não ocorrer panic, considera sucesso
			})
		})
	})
}
