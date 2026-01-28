package tests

import (
	"os"
	"testing"

	"github.com/OryaHub/agent-os-instance/libs/bdd"
	"github.com/OryaHub/agent-os-instance/libs/components"
	"github.com/OryaHub/agent-os-instance/libs/interfaces"
	"github.com/OryaHub/agent-os-instance/libs/utils"
)

func TestNewDatadogOperation(t *testing.T) {
	bdd.Feature(t, "TestNewDatadogOperation", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Instanciar datadog opertions", func(s *bdd.Scenario) {
			var datadogOperation interfaces.IDatadogOperation
			var logger interfaces.ILogger
			var err error
			s.When("logger é criado", func() {
				logger = utils.NewOryaLoggerText(os.Stdout)
			})
			s.When("NewDatadogOperation é chamado", func() {
				datadogOperation, err = components.DatadogOperation(logger)
			})
			s.Then("nenhum erro deve ocorrer", func(t *testing.T) {
				bdd.AssertNoError(t, err, "não deve retornar erro")
			})
			s.Then("o datadogOperation deve ser diferente de nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, datadogOperation, "datadogOperation deve ser diferente de nil")
			})
		})
	})
}

func TestNewDatadogOperationSetup(t *testing.T) {
	bdd.Feature(t, "TestNewDatadogOperationSetup", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Instanciar datadog opertions", func(s *bdd.Scenario) {
			var datadogOperation interfaces.IDatadogOperation
			var logger interfaces.ILogger
			var err error
			s.When("logger é criado", func() {
				logger = utils.NewOryaLoggerText(os.Stdout)
			})
			s.When("NewDatadogOperation é chamado", func() {
				datadogOperation, err = components.DatadogOperation(logger)
			})
			s.Then("nenhum erro deve ocorrer", func(t *testing.T) {
				bdd.AssertNoError(t, err, "não deve retornar erro")
			})
			s.Then("o datadogOperation deve ser diferente de nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, datadogOperation, "datadogOperation deve ser diferente de nil")
			})
			s.When("Executar configuração do datadogOperation", func() {
				err = datadogOperation.Setup()
			})
			s.Then("nenhum erro deve ocorrer na configuração", func(t *testing.T) {
				bdd.AssertNoError(t, err, "não deve retornar erro")
			})
		})
	})
}

func TestNewDatadogOperationGetInfos(t *testing.T) {
	bdd.Feature(t, "TestNewDatadogOperationGetInfos", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Instanciar datadog opertions", func(s *bdd.Scenario) {
			var datadogOperation interfaces.IDatadogOperation
			var logger interfaces.ILogger
			var data []byte
			var err error
			s.When("logger é criado", func() {
				logger = utils.NewOryaLoggerText(os.Stdout)
			})
			s.When("NewDatadogOperation é chamado", func() {
				datadogOperation, err = components.DatadogOperation(logger)
			})
			s.Then("nenhum erro deve ocorrer", func(t *testing.T) {
				bdd.AssertNoError(t, err, "não deve retornar erro")
			})
			s.Then("o datadogOperation deve ser diferente de nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, datadogOperation, "datadogOperation deve ser diferente de nil")
			})
			s.When("Executar configuração do datadogOperation", func() {
				err = datadogOperation.Setup()
			})
			s.Then("nenhum erro deve ocorrer na configuração", func(t *testing.T) {
				bdd.AssertNoError(t, err, "não deve retornar erro")
			})
			s.When("quando busco as informações do datadog", func() {
				data, err = datadogOperation.GetInfos()
			})
			s.Then("nenhum erro deve ocorrer na busca das informações", func(t *testing.T) {
				bdd.AssertNoError(t, err, "não deve retornar erro")
			})
			s.Then("os dados retornados não devem ser nulos", func(t *testing.T) {
				bdd.AssertIsNotNil(t, data, "dados retornados não devem ser nulos")
			})
			bdd.Printf("data: %v\n", string(data))
		})
	})
}
