package tests

import (
	"log"
	"os"
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/adapters"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

func TestNewDatadogAdapter(t *testing.T) {
	bdd.Feature(t, "TestNewDatadogAdapter", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve criar DatadogAdapter sem erro", func(s *bdd.Scenario) {
			var datadogAdapter *adapters.DatadogAdapter
			var logger interfaces.ILogger
			s.When("logger é criado", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.When("NewDatadogAdapter é chamado", func() {
				datadogAdapter = adapters.NewDatadogAdapter(logger)
			})
			s.Then("DatadogAdapter não deve ser nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, datadogAdapter, "DatadogAdapter deve ser diferente de nil")
			})
		})
	})
}

func TestDatadogAdapterSetup(t *testing.T) {
	bdd.Feature(t, "TestDatadogAdapterSetup", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Setup não deve retornar erro", func(s *bdd.Scenario) {
			var datadogAdapter *adapters.DatadogAdapter
			var logger interfaces.ILogger
			var err error
			s.When("logger é criado", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.Given("DatadogAdapter válido", func() {
				datadogAdapter = adapters.NewDatadogAdapter(logger)
			})
			s.When("Setup é chamado", func() {
				err = datadogAdapter.Setup()
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
		})
	})
}

func TestDatadogAdapterIsActive(t *testing.T) {
	bdd.Feature(t, "TestDatadogAdapterIsActive", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("IsActive não deve retornar erro", func(s *bdd.Scenario) {
			var datadogAdapter *adapters.DatadogAdapter
			var isActive bool
			var logger interfaces.ILogger
			var err error
			s.When("logger é criado", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.Given("DatadogAdapter válido e setup executado", func() {
				datadogAdapter = adapters.NewDatadogAdapter(logger)
				err = datadogAdapter.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("IsActive é chamado", func() {
				isActive, err = datadogAdapter.IsActive()
			})
			s.Then("não deve retornar erro e deve retornar isActive", func(t *testing.T) {
				bdd.AssertNoError(t, err, "IsActive não deve retornar erro")
				log.Printf("isActive: %v\n", isActive)
			})
		})
	})
}

func TestDatadogAdapterInstallAgent(t *testing.T) {
	bdd.Feature(t, "TestDatadogAdapterInstallAgent", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("InstallAgent não deve retornar erro", func(s *bdd.Scenario) {
			var datadogAdapter *adapters.DatadogAdapter
			var logger interfaces.ILogger
			var err error
			s.When("logger é criado", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.Given("DatadogAdapter válido e setup executado", func() {
				datadogAdapter = adapters.NewDatadogAdapter(logger)
				err = datadogAdapter.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("InstallAgent é chamado", func() {
				if err == nil {
					err = datadogAdapter.InstallAgent("datadoghq.com", "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
				}
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "InstallAgent não deve retornar erro")
			})
		})
	})
}

func TestDatadogAdapterUninstallAgent(t *testing.T) {
	bdd.Feature(t, "TestDatadogAdapterUninstallAgent", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("UninstallAgent não deve retornar erro", func(s *bdd.Scenario) {
			var datadogAdapter *adapters.DatadogAdapter
			var logger interfaces.ILogger
			var err error
			s.When("logger é criado", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.Given("DatadogAdapter válido e setup executado", func() {
				datadogAdapter = adapters.NewDatadogAdapter(logger)
				err = datadogAdapter.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("UninstallAgent é chamado", func() {
				if err == nil {
					err = datadogAdapter.UninstallAgent()
				}
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "UninstallAgent não deve retornar erro")
			})
		})
	})
}

func TestDatadogAdapterDiscoverDatadogConfigPath(t *testing.T) {
	bdd.Feature(t, "TestDatadogAdapterDiscoverDatadogConfigPath", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("DiscoverDatadogConfigPath não deve retornar erro", func(s *bdd.Scenario) {
			var datadogAdapter *adapters.DatadogAdapter
			var logger interfaces.ILogger
			var filePath string
			var err error
			s.When("logger é criado", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.Given("DatadogAdapter válido e setup executado", func() {
				datadogAdapter = adapters.NewDatadogAdapter(logger)
				err = datadogAdapter.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("DiscoverDatadogConfigPath é chamado", func() {
				filePath, err = datadogAdapter.DiscoverDatadogConfigPath()
			})
			s.Then("não deve retornar erro e deve retornar filePath", func(t *testing.T) {
				bdd.AssertNoError(t, err, "DiscoverDatadogConfigPath não deve retornar erro")
				log.Printf("filePath: %s\n", filePath)
			})
		})
	})
}

func TestDatadogAdapterDecodeBase64(t *testing.T) {
	bdd.Feature(t, "TestDatadogAdapterDecodeBase64", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("DecodeBase64 não deve retornar erro", func(s *bdd.Scenario) {
			var datadogAdapter *adapters.DatadogAdapter
			var logger interfaces.ILogger
			var decoded []byte
			var err error
			s.When("logger é criado", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.Given("DatadogAdapter válido e setup executado", func() {
				datadogAdapter = adapters.NewDatadogAdapter(logger)
				err = datadogAdapter.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("DecodeBase64 é chamado", func() {
				decoded, err = datadogAdapter.DecodeBase64("dGVzdGU=")
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "DecodeBase64 não deve retornar erro")
			})
			bdd.Printf("decoded: %s\n", string(decoded))
		})
	})
}

func TestDatadogAdapterUpdateConfigFileDatadog(t *testing.T) {

	bdd.Feature(t, "TestDatadogAdapterUpdateConfigFileDatadog", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("UpdateConfigFileDatadog não deve retornar erro", func(s *bdd.Scenario) {
			var datadogAdapter *adapters.DatadogAdapter
			var logger interfaces.ILogger
			var err error
			s.When("logger é criado", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.Given("DatadogAdapter válido e setup executado", func() {
				datadogAdapter = adapters.NewDatadogAdapter(logger)
				err = datadogAdapter.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")

			})
			s.When("UpdateConfigFileDatadog é chamado", func() {
				filePath := "/etc/datadog-agent/conf.d/journald.d/conf.yaml"

				err = datadogAdapter.UpdateConfigFileDatadog(filePath)

			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "UpdateConfigFileDatadog não deve retornar erro")
			})
		})
	})
}

func TestDatadogAdapterUpdateRepository(t *testing.T) {

	bdd.Feature(t, "TestDatadogAdapterUpdateRepository", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("UpdateRepository não deve retornar erro", func(s *bdd.Scenario) {
			var datadogAdapter *adapters.DatadogAdapter
			var logger interfaces.ILogger
			var err error
			s.When("logger é criado", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.Given("DatadogAdapter válido e setup executado", func() {
				datadogAdapter = adapters.NewDatadogAdapter(logger)
				err = datadogAdapter.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")

			})
			s.When("UpdateRepository é chamado", func() {
				err = datadogAdapter.UpdateRepository()

			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "UpdateRepository não deve retornar erro")
			})
		})
	})
}

func TestDatadogAdapterGetVersion(t *testing.T) {
	bdd.Feature(t, "TestDatadogAdapterGetVersion", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("GetVersion não deve retornar erro", func(s *bdd.Scenario) {
			var datadogAdapter *adapters.DatadogAdapter
			var logger interfaces.ILogger
			var version string
			var err error
			s.When("logger é criado", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.Given("DatadogAdapter válido e setup executado", func() {
				datadogAdapter = adapters.NewDatadogAdapter(logger)
				err = datadogAdapter.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")

			})
			s.When("GetVersion é chamado", func() {
				version, err = datadogAdapter.GetVersion()

			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "GetVersion não deve retornar erro")
			})
			bdd.Printf("Versão do Datadog: %s\n", version)
		})
	})
}
