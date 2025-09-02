package tests

import (
	"os"
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/adapters"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/dto"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

func TestNewUpdaterAdapter(t *testing.T) {
	bdd.Feature(t, "UpdaterAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("instanciar o updater adapter", func(s *bdd.Scenario) {
			var updater *adapters.UpdaterAdapter
			s.Given("um logger válido", func() {})
			s.When("instancio o updater adapter", func() {
				updater = adapters.NewUpdaterAdapter(logger)
			})
			s.Then("o updater adapter deve ser diferente de nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, updater, "updater deve ser diferente de nil")
			})
		})
	})
}

func TestUpdaterAdapterPrepare(t *testing.T) {
	bdd.Feature(t, "UpdaterAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("preparar o updater adapter", func(s *bdd.Scenario) {
			var updater *adapters.UpdaterAdapter
			var err error
			s.Given("um updater adapter instanciado", func() {
				updater = adapters.NewUpdaterAdapter(logger)
			})
			s.When("chamo Prepare", func() {
				err = updater.Prepare()
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "Prepare não deve retornar erro")
			})
		})
	})
}

func TestUpdaterAdapterGetAgentVersion(t *testing.T) {
	bdd.Feature(t, "UpdaterAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("obter a versão do agente", func(s *bdd.Scenario) {
			var updater *adapters.UpdaterAdapter
			var version string
			var err error
			s.Given("um updater adapter instanciado", func() {
				updater = adapters.NewUpdaterAdapter(logger)
			})
			s.When("chamo Prepare", func() {
				err = updater.Prepare()
			})
			s.When("chamo GetAgentVersion", func() {
				version, err = updater.GetAgentVersion()
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "GetAgentVersion não deve retornar erro")
			})
			bdd.Printf("version: %s", version)
		})
	})
}

func TestUpdaterAdapterGetContentReceived(t *testing.T) {
	bdd.Feature(t, "UpdaterAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("obter o conteúdo recebido", func(s *bdd.Scenario) {
			var updater *adapters.UpdaterAdapter
			var content []byte
			var err error
			s.Given("um updater adapter instanciado", func() {
				updater = adapters.NewUpdaterAdapter(logger)
			})
			s.When("chamo Prepare", func() {
				err = updater.Prepare()
			})
			s.When("chamo GetContentReceived", func() {
				content, err = updater.GetContentReceived()
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "GetContentReceived não deve retornar erro")
			})
			bdd.Printf("content: %s", content)
		})
	})
}

func TestUpdaterAdapterGetAgentVersionFromSignal(t *testing.T) {
	bdd.Feature(t, "UpdaterAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("obter a versão do agente a partir do sinal", func(s *bdd.Scenario) {
			var updater *adapters.UpdaterAdapter
			var version string
			var response []byte
			var err error
			s.Given("um updater adapter instanciado", func() {
				updater = adapters.NewUpdaterAdapter(logger)
			})
			s.When("chamo Prepare", func() {
				err = updater.Prepare()
			})
			s.When("preparo o signal", func() {
				response = []byte(`{"signal":{"type":"update","agents":{"docp-agent":{"version":"0.1.0"}}}}`)
			})
			s.When("chamo GetAgentVersion", func() {
				version, err = updater.GetAgentVersionFromSignal(response)
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "GetAgentVersion não deve retornar erro")
			})
			bdd.Printf("version: %s", version)
		})
	})
}

func TestUpdaterAdapterGetAgentRollbackVersion(t *testing.T) {
	bdd.Feature(t, "UpdaterAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("obter a versão de rollback do agente", func(s *bdd.Scenario) {
			var updater *adapters.UpdaterAdapter
			var version string
			var err error
			s.Given("um updater adapter instanciado", func() {
				updater = adapters.NewUpdaterAdapter(logger)
			})
			s.When("chamo Prepare", func() {
				err = updater.Prepare()
			})
			s.When("chamo GetAgentRollbackVersion", func() {
				version, err = updater.GetAgentRollbackVersion()
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "GetAgentRollbackVersion não deve retornar erro")
			})
			bdd.Printf("version: %s", version)
		})
	})
}

func TestUpdaterAdapterExecuteUpdateVersion(t *testing.T) {
	bdd.Feature(t, "UpdaterAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("ExecuteUpdateVersion executa (mock)", func(s *bdd.Scenario) {
			var updater *adapters.UpdaterAdapter
			var err error
			s.Given("Cria logger", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
				bdd.AssertIsNotNil(t, logger, "Logger deve ser criado")
			})
			s.Given("UpdaterAdapter criado", func() {
				updater = adapters.NewUpdaterAdapter(logger)
			})
			s.Given("Configurar o updater adapter", func() {
				err = updater.Prepare()
				bdd.AssertNoError(t, err, "Prepare não deve retornar erro")
			})
			s.When("chamo ExecuteUpdateVersion", func() {
				err = updater.ExecuteUpdateVersion("0.1.0")
			})
			s.Then("não deve retornar erro (mock)", func(t *testing.T) {
				// Em ambiente real, seria necessário mockar os utilitários do SO
				bdd.AssertNoError(t, err, "ExecuteUpdateVersion não deve retornar erro (mock)")
			})
		})
	})
}

func TestUpdaterAdapterExecuteRollbackVersion(t *testing.T) {
	bdd.Feature(t, "UpdaterAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("ExecuteRollbackVersion executa (mock)", func(s *bdd.Scenario) {
			var updater *adapters.UpdaterAdapter
			var err error
			s.Given("Cria logger", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
				bdd.AssertIsNotNil(t, logger, "Logger deve ser criado")
			})
			s.Given("UpdaterAdapter criado", func() {
				updater = adapters.NewUpdaterAdapter(logger)
			})
			s.Given("Configurar o updater adapter", func() {
				err = updater.Prepare()
				bdd.AssertNoError(t, err, "Prepare não deve retornar erro")
			})
			s.When("chamo ExecuteRollbackVersion", func() {
				err = updater.ExecuteRollbackVersion("0.1.1")
			})
			s.Then("não deve retornar erro (mock)", func(t *testing.T) {
				// Em ambiente real, seria necessário mockar os utilitários do SO
				bdd.AssertNoError(t, err, "ExecuteRollbackVersion não deve retornar erro")
			})
		})
	})
}

func TestUpdaterAdapterFetchAgentVersions(t *testing.T) {
	bdd.Feature(t, "UpdaterAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("FetchAgentVersions executa (mock)", func(s *bdd.Scenario) {
			var updater *adapters.UpdaterAdapter
			var agentVersions dto.AgentVersions
			var err error
			s.Given("Cria logger", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
				bdd.AssertIsNotNil(t, logger, "Logger deve ser criado")
			})
			s.Given("UpdaterAdapter criado", func() {
				updater = adapters.NewUpdaterAdapter(logger)
			})
			s.Given("Configurar o updater adapter", func() {
				err = updater.Prepare()
				bdd.AssertNoError(t, err, "Prepare não deve retornar erro")
			})
			s.When("chamo FetchAgentVersions", func() {
				agentVersions, err = updater.FetchAgentVersions()
			})
			s.Then("não deve retornar erro (mock)", func(t *testing.T) {
				bdd.AssertNoError(t, err, "FetchAgentVersions não deve retornar erro")
			})
			bdd.Printf("agentVersions: %+v", agentVersions)
		})
	})
}

func TestUpdaterAdapterValidateSuccessUpdated(t *testing.T) {
	bdd.Feature(t, "UpdaterAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("validar se a atualização foi bem-sucedida", func(s *bdd.Scenario) {
			var updater *adapters.UpdaterAdapter
			var success bool
			var err error
			s.Given("um updater adapter instanciado", func() {
				updater = adapters.NewUpdaterAdapter(logger)
			})
			s.When("chamo Prepare", func() {
				err = updater.Prepare()
			})
			s.When("valido se o manager e agent estao ok", func() {
				success, err = updater.ValidateSuccessUpdated()
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "ValidateSuccessUpdated não deve retornar erro")
			})
			bdd.Printf("success: %t", success)
		})
	})
}

func TestUpdaterAdapterUpdaterUninstall(t *testing.T) {
	bdd.Feature(t, "UpdaterAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("UpdaterUninstall executa sem erro (mock)", func(s *bdd.Scenario) {
			var updater *adapters.UpdaterAdapter
			var err error
			s.Given("Cria logger", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
				bdd.AssertIsNotNil(t, logger, "Logger deve ser criado")
			})
			s.Given("um updater adapter instanciado", func() {
				updater = adapters.NewUpdaterAdapter(logger)
			})
			s.When("chamo Prepare", func() {
				err = updater.Prepare()
			})
			s.When("chamo UpdaterUninstall", func() {
				err = updater.UpdaterUninstall()
			})
			s.Then("não deve retornar erro (mock)", func(t *testing.T) {
				bdd.AssertNoError(t, err, "UpdaterUninstall não deve retornar erro (mock)")
			})
		})
	})
}
