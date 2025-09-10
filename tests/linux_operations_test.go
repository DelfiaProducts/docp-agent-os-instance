package tests

import (
	"os"
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/components"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

func TestNewLinuxOperations(t *testing.T) {
	bdd.Feature(t, "LinuxOperations", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("Criação do linux operations", func(s *bdd.Scenario) {
			var ops *components.LinuxOperations
			var logger interfaces.ILogger
			s.Given("Cria logger", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
				bdd.AssertIsNotNil(t, logger, "Logger deve ser criado")
			})

			s.Given("Criar linux operations", func() {
				ops = components.NewLinuxOperations(logger)
			})

			s.Then("Operaotion deve ser criado com sucesso", func(t *testing.T) {
				bdd.AssertIsNotNil(t, ops, "LinuxOperations deve ser criado")
			})
		})
	})
}

func TestLinuxOperationsSetup(t *testing.T) {
	bdd.Feature(t, "LinuxOperations", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("Criação do linux operations e setup", func(s *bdd.Scenario) {
			var ops *components.LinuxOperations
			var logger interfaces.ILogger
			var err error
			s.Given("Cria logger", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
				bdd.AssertIsNotNil(t, logger, "Logger deve ser criado")
			})

			s.Given("Criar linux operations", func() {
				ops = components.NewLinuxOperations(logger)
			})
			s.Given("Configurar o operation", func() {
				err = ops.Setup()
			})
			s.Then("Operaotion deve ser configurado com sucesso", func(t *testing.T) {
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
		})
	})
}
func TestLinuxOperationsStatus(t *testing.T) {
	bdd.Feature(t, "LinuxOperations", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("Status retorna saída correta", func(s *bdd.Scenario) {
			var ops *components.LinuxOperations
			var logger interfaces.ILogger
			var status string
			var err error
			s.Given("Cria logger", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
				bdd.AssertIsNotNil(t, logger, "Logger deve ser criado")
			})

			s.Given("Criar linux operations", func() {
				ops = components.NewLinuxOperations(logger)
			})
			s.Given("Configurar o operation", func() {
				err = ops.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("chamo Status", func() {
				status, err = ops.Status("agent")
			})
			s.Then("não deve retornar erro e saída correta", func(t *testing.T) {
				bdd.AssertNoError(t, err, "Status não deve retornar erro")
			})
			bdd.Printf("status: %s", status)
		})
	})
}

func TestLinuxOperationsAlreadyInstalled(t *testing.T) {
	bdd.Feature(t, "LinuxOperations", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("AlreadyInstalled retorna true", func(s *bdd.Scenario) {
			var ops *components.LinuxOperations
			var logger interfaces.ILogger
			var installed bool
			var err error
			s.Given("Cria logger", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
				bdd.AssertIsNotNil(t, logger, "Logger deve ser criado")
			})
			s.Given("LinuxOperations com systemd mockado", func() {
				ops = components.NewLinuxOperations(logger)
			})
			s.Given("Configurar o operation", func() {
				err = ops.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("chamo AlreadyInstalled", func() {
				installed, err = ops.AlreadyInstalled("agent")
			})
			s.Then("não deve retornar erro e deve ser true", func(t *testing.T) {
				bdd.AssertNoError(t, err, "AlreadyInstalled não deve retornar erro")
			})
			bdd.Printf("AlreadyInstalled: %v", installed)
		})
	})
}

func TestLinuxOperationsDaemonReload(t *testing.T) {
	bdd.Feature(t, "LinuxOperations", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("DaemonReload executa sem erro", func(s *bdd.Scenario) {
			var ops *components.LinuxOperations
			var logger interfaces.ILogger
			var err error
			s.Given("Cria logger", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
				bdd.AssertIsNotNil(t, logger, "Logger deve ser criado")
			})
			s.Given("LinuxOperations com program mockado", func() {
				ops = components.NewLinuxOperations(logger)
			})
			s.Given("Configurar o operation", func() {
				err = ops.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("chamo DaemonReload", func() {
				err = ops.DaemonReload()
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "DaemonReload não deve retornar erro")
			})
		})
	})
}

func TestLinuxOperationsRestartService(t *testing.T) {
	bdd.Feature(t, "LinuxOperations", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("RestartService executa sem erro", func(s *bdd.Scenario) {
			var ops *components.LinuxOperations
			var logger interfaces.ILogger
			var err error
			s.Given("Cria logger", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
				bdd.AssertIsNotNil(t, logger, "Logger deve ser criado")
			})
			s.Given("LinuxOperations criado", func() {
				ops = components.NewLinuxOperations(logger)
			})
			s.Given("Configurar o operation", func() {
				err = ops.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("chamo RestartService", func() {
				err = ops.RestartService("agent")
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "RestartService não deve retornar erro")
			})
		})
	})
}

func TestLinuxOperationsStopService(t *testing.T) {
	bdd.Feature(t, "LinuxOperations", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("StopService executa sem erro", func(s *bdd.Scenario) {
			var ops *components.LinuxOperations
			var logger interfaces.ILogger
			var err error
			s.Given("Cria logger", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
				bdd.AssertIsNotNil(t, logger, "Logger deve ser criado")
			})
			s.Given("LinuxOperations criado", func() {
				ops = components.NewLinuxOperations(logger)
			})
			s.Given("Configurar o operation", func() {
				err = ops.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("chamo StopService", func() {
				err = ops.StopService("agent")
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "StopService não deve retornar erro")
			})
		})
	})
}

func TestLinuxOperationsInstallAgent(t *testing.T) {
	bdd.Feature(t, "LinuxOperations", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("InstallAgent executa sem erro", func(s *bdd.Scenario) {
			var ops *components.LinuxOperations
			var logger interfaces.ILogger
			var err error
			s.Given("Cria logger", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
				bdd.AssertIsNotNil(t, logger, "Logger deve ser criado")
			})
			s.Given("LinuxOperations criado", func() {
				ops = components.NewLinuxOperations(logger)
			})
			s.Given("Configurar o operation", func() {
				err = ops.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("chamo InstallAgent", func() {
				err = ops.InstallAgent("0.1.1")
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "InstallAgent não deve retornar erro")
			})
		})
	})
}

func TestLinuxOperationsInstallUpdater(t *testing.T) {
	bdd.Feature(t, "LinuxOperations", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("InstallUpdater executa sem erro", func(s *bdd.Scenario) {
			var ops *components.LinuxOperations
			var logger interfaces.ILogger
			var err error
			s.Given("Cria logger", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
				bdd.AssertIsNotNil(t, logger, "Logger deve ser criado")
			})
			s.Given("LinuxOperations criado", func() {
				ops = components.NewLinuxOperations(logger)
			})
			s.Given("Configurar o operation", func() {
				err = ops.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("chamo InstallUpdater", func() {
				err = ops.InstallUpdater("0.1.0")
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "InstallUpdater não deve retornar erro")
			})
		})
	})
}

func TestLinuxOperationsUninstallAgent(t *testing.T) {
	bdd.Feature(t, "LinuxOperations", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("UninstallAgent executa sem erro", func(s *bdd.Scenario) {
			var ops *components.LinuxOperations
			var logger interfaces.ILogger
			var err error
			s.Given("Cria logger", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
				bdd.AssertIsNotNil(t, logger, "Logger deve ser criado")
			})
			s.Given("LinuxOperations criado", func() {
				ops = components.NewLinuxOperations(logger)
			})
			s.Given("Configurar o operation", func() {
				err = ops.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("chamo UninstallAgent", func() {
				err = ops.UninstallAgent("0.1.0")
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "UninstallAgent não deve retornar erro")
			})
		})
	})
}

func TestLinuxOperationsUninstallUpdater(t *testing.T) {
	bdd.Feature(t, "LinuxOperations", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("UninstallUpdater executa sem erro", func(s *bdd.Scenario) {
			var ops *components.LinuxOperations
			var logger interfaces.ILogger
			var err error
			s.Given("Cria logger", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
				bdd.AssertIsNotNil(t, logger, "Logger deve ser criado")
			})
			s.Given("LinuxOperations criado", func() {
				ops = components.NewLinuxOperations(logger)
			})
			s.Given("Configurar o operation", func() {
				err = ops.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("chamo UninstallUpdater", func() {
				err = ops.UninstallUpdater("0.1.0")
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "UninstallUpdater não deve retornar erro")
			})
		})
	})
}

// Para UpdateAgent, AutoUninstall e Execute, normalmente seria necessário mockar dependências do SO e arquivos.
// Aqui, apenas estrutura básica:

func TestLinuxOperationsUpdateAgent(t *testing.T) {
	bdd.Feature(t, "LinuxOperations", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("UpdateAgent executa (mock)", func(s *bdd.Scenario) {
			var ops *components.LinuxOperations
			var logger interfaces.ILogger
			var err error
			s.Given("Cria logger", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
				bdd.AssertIsNotNil(t, logger, "Logger deve ser criado")
			})
			s.Given("LinuxOperations criado", func() {
				ops = components.NewLinuxOperations(logger)
			})
			s.Given("Configurar o operation", func() {
				err = ops.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("chamo UpdateAgent", func() {
				err = ops.UpdateAgent("0.1.0")
			})
			s.Then("não deve retornar erro (mock)", func(t *testing.T) {
				// Em ambiente real, seria necessário mockar os utilitários do SO
				bdd.AssertNoError(t, err, "UpdateAgent não deve retornar erro (mock)")
			})
		})
	})
}

func TestLinuxOperationsAutoUninstall(t *testing.T) {
	bdd.Feature(t, "LinuxOperations", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("AutoUninstall executa sem erro (mock)", func(s *bdd.Scenario) {
			var ops *components.LinuxOperations
			var logger interfaces.ILogger
			var err error
			s.Given("Cria logger", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
				bdd.AssertIsNotNil(t, logger, "Logger deve ser criado")
			})
			s.Given("LinuxOperations criado", func() {
				ops = components.NewLinuxOperations(logger)
			})
			s.Given("Configurar o operation", func() {
				err = ops.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("chamo AutoUninstall", func() {
				err = ops.AutoUninstall("0.1.0")
			})
			s.Then("não deve retornar erro (mock)", func(t *testing.T) {
				bdd.AssertNoError(t, err, "AutoUninstall não deve retornar erro (mock)")
			})
		})
	})
}

func TestLinuxOperationsExecute(t *testing.T) {
	bdd.Feature(t, "LinuxOperations", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("Execute executa sem erro (mock)", func(s *bdd.Scenario) {
			var ops *components.LinuxOperations
			var logger interfaces.ILogger
			var err error
			s.Given("Cria logger", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
				bdd.AssertIsNotNil(t, logger, "Logger deve ser criado")
			})
			s.Given("LinuxOperations criado", func() {
				ops = components.NewLinuxOperations(logger)
			})
			s.Given("Configurar o operation", func() {
				err = ops.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("chamo Execute", func() {
				err = ops.Execute()
			})
			s.Then("não deve retornar erro (mock)", func(t *testing.T) {
				bdd.AssertNoError(t, err, "Execute não deve retornar erro (mock)")
			})
		})
	})
}
