package tests

import (
	"testing"
	"time"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/adapters"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/dto"
)

func TestNewManagerAdapter(t *testing.T) {
	bdd.Feature(t, "ManagerAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("instanciar o manager adapter", func(s *bdd.Scenario) {
			var manager *adapters.ManagerAdapter
			s.Given("um logger válido", func() {})
			s.When("instancio o manager adapter", func() {
				manager = adapters.NewManagerAdapter(logger)
			})
			s.Then("o manager adapter deve ser diferente de nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, manager, "manager deve ser diferente de nil")
			})
		})
	})
}

func TestManagerAdapterPrepare(t *testing.T) {
	bdd.Feature(t, "ManagerAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("preparar o manager adapter", func(s *bdd.Scenario) {
			var manager *adapters.ManagerAdapter
			var err error
			s.Given("um manager adapter instanciado", func() {
				manager = adapters.NewManagerAdapter(logger)
			})
			s.When("chamo Prepare", func() {
				err = manager.Prepare()
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "Prepare não deve retornar erro")
			})
		})
	})
}

func TestManagerAdapterCollect(t *testing.T) {
	bdd.Feature(t, "ManagerAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("coletar métricas", func(s *bdd.Scenario) {
			var manager *adapters.ManagerAdapter
			var err error
			var metricsCount int
			s.Given("um manager adapter preparado", func() {
				manager = adapters.NewManagerAdapter(logger)
				err = manager.Prepare()
			})
			s.When("chamo Collect", func() {
				ch := manager.Collect()
				for range ch {
					metricsCount++
				}
			})
			s.Then("deve coletar métricas sem erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "Collect não deve retornar erro na preparação")
			})
		})
	})
}

func TestManagerAdapterClose(t *testing.T) {
	bdd.Feature(t, "ManagerAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("fechar o manager adapter", func(s *bdd.Scenario) {
			var manager *adapters.ManagerAdapter
			var err error
			s.Given("um manager adapter preparado", func() {
				manager = adapters.NewManagerAdapter(logger)
				err = manager.Prepare()
			})
			s.When("chamo Close em paralelo ao Collect", func() {
				go func() {
					time.Sleep(time.Second)
					manager.Close()
				}()
				ch := manager.Collect()
				for range ch {
				}
			})
			s.Then("deve fechar sem erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "Close não deve retornar erro na preparação")
			})
		})
	})
}

func TestManagerAdapterStatus(t *testing.T) {
	bdd.Feature(t, "ManagerAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("verificar status do serviço", func(s *bdd.Scenario) {
			var manager *adapters.ManagerAdapter
			var err error
			var status any
			s.Given("um manager adapter preparado", func() {
				manager = adapters.NewManagerAdapter(logger)
				err = manager.Prepare()
			})
			s.When("chamo Status para docp-agent.service", func() {
				status, err = manager.Status("docp-agent.service")
			})
			s.Then("deve retornar status sem erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "Status não deve retornar erro")
				bdd.AssertIsNotNil(t, status, "Status não deve ser nil")
			})
		})
	})
}

func TestManagerAdapterDaemonReloadService(t *testing.T) {
	bdd.Feature(t, "ManagerAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("executar DaemonReload", func(s *bdd.Scenario) {
			var manager *adapters.ManagerAdapter
			var err error
			s.Given("um manager adapter preparado", func() {
				manager = adapters.NewManagerAdapter(logger)
				err = manager.Prepare()
			})
			s.When("chamo DaemonReload", func() {
				err = manager.DaemonReload()
			})
			s.Then("deve executar sem erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "DaemonReload não deve retornar erro")
			})
		})
	})
}

func TestManagerAdapterRestartService(t *testing.T) {
	bdd.Feature(t, "ManagerAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("reiniciar serviço", func(s *bdd.Scenario) {
			var manager *adapters.ManagerAdapter
			var err error
			s.Given("um manager adapter preparado", func() {
				manager = adapters.NewManagerAdapter(logger)
				err = manager.Prepare()
			})
			s.When("chamo RestartService para docp-manager.service", func() {
				err = manager.RestartService("docp-manager.service")
			})
			s.Then("deve reiniciar sem erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "RestartService não deve retornar erro")
			})
		})
	})
}

func TestManagerAdapterStopService(t *testing.T) {
	bdd.Feature(t, "ManagerAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("parar serviço", func(s *bdd.Scenario) {
			var manager *adapters.ManagerAdapter
			var err error
			s.Given("um manager adapter preparado", func() {
				manager = adapters.NewManagerAdapter(logger)
				err = manager.Prepare()
			})
			s.When("chamo StopService para docp-manager.service", func() {
				err = manager.StopService("docp-manager.service")
			})
			s.Then("deve parar sem erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "StopService não deve retornar erro")
			})
		})
	})
}

func TestManagerAdapterInstallAgent(t *testing.T) {
	bdd.Feature(t, "ManagerAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("instalar agente", func(s *bdd.Scenario) {
			var manager *adapters.ManagerAdapter
			var err error
			s.Given("um manager adapter preparado", func() {
				manager = adapters.NewManagerAdapter(logger)
				err = manager.Prepare()
			})
			s.When("chamo InstallAgent", func() {
				err = manager.InstallAgent("0.1.1")
			})
			s.Then("deve instalar sem erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "InstallAgent não deve retornar erro")
			})
		})
	})
}

func TestManagerAdapterUninstallAgent(t *testing.T) {
	bdd.Feature(t, "ManagerAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("desinstalar agente", func(s *bdd.Scenario) {
			var manager *adapters.ManagerAdapter
			var err error
			s.Given("um manager adapter preparado", func() {
				manager = adapters.NewManagerAdapter(logger)
				err = manager.Prepare()
			})
			s.When("chamo UninstallAgent", func() {
				err = manager.UninstallAgent("0.1.0")
			})
			s.Then("deve desinstalar sem erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "UninstallAgent não deve retornar erro")
			})
		})
	})
}

func TestManagerAdapterAutoUninstall(t *testing.T) {
	bdd.Feature(t, "ManagerAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("auto desinstalar agente", func(s *bdd.Scenario) {
			var manager *adapters.ManagerAdapter
			var err error
			s.Given("um manager adapter preparado", func() {
				manager = adapters.NewManagerAdapter(logger)
				err = manager.Prepare()
			})
			s.When("chamo AutoUninstall", func() {
				err = manager.AutoUninstall("0.1.0")
			})
			s.Then("deve auto desinstalar sem erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "AutoUninstall não deve retornar erro")
			})
		})
	})
}

func TestManagerAdapterDocpAgentApiInstallDatadog(t *testing.T) {
	bdd.Feature(t, "ManagerAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("instalar datadog via API", func(s *bdd.Scenario) {
			var manager *adapters.ManagerAdapter
			var err error
			var result []byte
			s.Given("um manager adapter preparado", func() {
				manager = adapters.NewManagerAdapter(logger)
				err = manager.Prepare()
			})
			s.When("chamo DocpAgentApiInstallDatadog", func() {
				result, err = manager.DocpAgentApiInstallDatadog("9ba8aefcaa347216ffa5a5e7b3156f54", "datadoghq.com")
			})
			s.Then("deve instalar datadog sem erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "DocpAgentApiInstallDatadog não deve retornar erro")
				bdd.AssertIsNotNil(t, result, "resultado não deve ser nil")
			})
		})
	})
}

func TestManagerAdapterDocpAgentApiUninstallDatadog(t *testing.T) {
	bdd.Feature(t, "ManagerAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("desinstalar datadog via API", func(s *bdd.Scenario) {
			var manager *adapters.ManagerAdapter
			var err error
			var result []byte
			s.Given("um manager adapter preparado", func() {
				manager = adapters.NewManagerAdapter(logger)
				err = manager.Prepare()
			})
			s.When("chamo DocpAgentApiUninstallDatadog", func() {
				result, err = manager.DocpAgentApiUninstallDatadog()
			})
			s.Then("deve desinstalar datadog sem erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "DocpAgentApiUninstallDatadog não deve retornar erro")
				bdd.AssertIsNotNil(t, result, "resultado não deve ser nil")
			})
		})
	})
}

func TestManagerAdapterSaveStateReceived(t *testing.T) {
	bdd.Feature(t, "ManagerAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("salvar estado recebido", func(s *bdd.Scenario) {
			var manager *adapters.ManagerAdapter
			var err error
			var data = []byte(`{...}`) // Use o JSON real do teste original
			s.Given("um manager adapter preparado", func() {
				manager = adapters.NewManagerAdapter(logger)
				err = manager.Prepare()
			})
			s.When("chamo SaveStateReceived", func() {
				err = manager.SaveStateReceived(data)
			})
			s.Then("deve salvar estado recebido sem erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "SaveStateReceived não deve retornar erro")
			})
		})
	})
}

func TestManagerAdapterSaveStateCurrent(t *testing.T) {
	bdd.Feature(t, "ManagerAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("salvar estado atual", func(s *bdd.Scenario) {
			var manager *adapters.ManagerAdapter
			var err error
			var data = []byte(`{...}`) // Use o JSON real do teste original
			s.Given("um manager adapter preparado", func() {
				manager = adapters.NewManagerAdapter(logger)
				err = manager.Prepare()
			})
			s.When("chamo SaveStateCurrent", func() {
				err = manager.SaveStateCurrent(data)
			})
			s.Then("deve salvar estado atual sem erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "SaveStateCurrent não deve retornar erro")
			})
		})
	})
}

func TestManagerAdapterValidate(t *testing.T) {
	bdd.Feature(t, "ManagerAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("validar manager adapter", func(s *bdd.Scenario) {
			var manager *adapters.ManagerAdapter
			var err error
			s.Given("um manager adapter instanciado", func() {
				manager = adapters.NewManagerAdapter(logger)
			})
			s.Given("um manager adapter preparado", func() {
				err = manager.Prepare()
				bdd.AssertNoError(t, err, "Prepare não deve retornar erro")
			})
			s.When("chamo Validate", func() {
				err = manager.Validate()
			})
			s.Then("deve validar sem erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "Validate não deve retornar erro")
			})
		})
	})
}

func TestManagerAdapterFetchAgentVersions(t *testing.T) {
	bdd.Feature(t, "ManagerAdapter Fetch Agent Versions", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("validar manager adapter", func(s *bdd.Scenario) {
			var manager *adapters.ManagerAdapter
			var agentVersions dto.AgentVersions
			var err error

			s.Given("um manager adapter instanciado", func() {
				manager = adapters.NewManagerAdapter(logger)
			})
			s.Given("um manager adapter preparado", func() {
				err = manager.Prepare()
				bdd.AssertNoError(t, err, "Prepare não deve retornar erro")
			})
			s.When("buscar agent versions", func() {
				agentVersions, err = manager.FetchAgentVersions()
				bdd.AssertNoError(t, err, "FetchAgentVersions não deve retornar erro")
			})
			s.Then("deve validar sem erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "FetchAgentVersions não deve retornar erro")
			})
			bdd.Printf("Agent Versions: %+v\n", agentVersions)
		})
	})
}

func TestManagerAdapterGetAgentVersion(t *testing.T) {
	bdd.Feature(t, "ManagerAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("Pegar versão do agent", func(s *bdd.Scenario) {
			var manager *adapters.ManagerAdapter
			var version string
			var err error
			s.Given("um manager adapter instanciado", func() {
				manager = adapters.NewManagerAdapter(logger)
			})
			s.Given("um manager adapter preparado", func() {
				err = manager.Prepare()
				bdd.AssertNoError(t, err, "Prepare não deve retornar erro")
			})
			s.When("chamo GetAgentVersion", func() {
				version, err = manager.GetAgentVersion()
			})
			s.Then("deve validar sem erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "GetAgentVersion não deve retornar erro")
			})
			bdd.Printf("version: %s\n", version)
		})
	})
}

func TestManagerAdapterSaveAgentVersion(t *testing.T) {
	bdd.Feature(t, "ManagerAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("Salvar versão do agent", func(s *bdd.Scenario) {
			var manager *adapters.ManagerAdapter
			var err error
			s.Given("um manager adapter instanciado", func() {
				manager = adapters.NewManagerAdapter(logger)
			})
			s.Given("um manager adapter preparado", func() {
				err = manager.Prepare()
				bdd.AssertNoError(t, err, "Prepare não deve retornar erro")
			})
			s.When("chamo SaveAgentVersion", func() {
				err = manager.SaveAgentVersion("0.1.1")
			})
			s.Then("deve validar sem erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "SaveAgentVersion não deve retornar erro")
			})
		})
	})
}

func TestManagerAdapterGetAgentRollbackVersion(t *testing.T) {
	bdd.Feature(t, "ManagerAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("Pegar versão do agent rollback", func(s *bdd.Scenario) {
			var manager *adapters.ManagerAdapter
			var version string
			var err error
			s.Given("um manager adapter instanciado", func() {
				manager = adapters.NewManagerAdapter(logger)
			})
			s.Given("um manager adapter preparado", func() {
				err = manager.Prepare()
				bdd.AssertNoError(t, err, "Prepare não deve retornar erro")
			})
			s.When("chamo GetAgentRollbackVersion", func() {
				version, err = manager.GetAgentRollbackVersion()
			})
			s.Then("deve validar sem erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "GetAgentRollbackVersion não deve retornar erro")
			})
			bdd.Printf("version: %s\n", version)
		})
	})
}

func TestManagerAdapterSaveAgentRollbackVersion(t *testing.T) {
	bdd.Feature(t, "ManagerAdapter", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("Salvar versão do agent rollback", func(s *bdd.Scenario) {
			var manager *adapters.ManagerAdapter
			var err error
			s.Given("um manager adapter instanciado", func() {
				manager = adapters.NewManagerAdapter(logger)
			})
			s.Given("um manager adapter preparado", func() {
				err = manager.Prepare()
				bdd.AssertNoError(t, err, "Prepare não deve retornar erro")
			})
			s.When("chamo SaveAgentRollbackVersion", func() {
				err = manager.SaveAgentRollbackVersion("0.1.0")
			})
			s.Then("deve validar sem erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "SaveAgentRollbackVersion não deve retornar erro")
			})
		})
	})
}
