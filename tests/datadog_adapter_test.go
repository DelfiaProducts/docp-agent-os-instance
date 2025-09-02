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
			var datadogLinux *adapters.DatadogAdapter
			var logger interfaces.ILogger
			s.When("logger é criado", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.When("NewDatadogAdapter é chamado", func() {
				datadogLinux = adapters.NewDatadogAdapter(logger)
			})
			s.Then("DatadogAdapter não deve ser nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, datadogLinux, "DatadogAdapter deve ser diferente de nil")
			})
		})
	})
}

func TestDatadogAdapterSetup(t *testing.T) {
	bdd.Feature(t, "TestDatadogAdapterSetup", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Setup não deve retornar erro", func(s *bdd.Scenario) {
			var datadogLinux *adapters.DatadogAdapter
			var logger interfaces.ILogger
			var err error
			s.When("logger é criado", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.Given("DatadogAdapter válido", func() {
				datadogLinux = adapters.NewDatadogAdapter(logger)
			})
			s.When("Setup é chamado", func() {
				if datadogLinux != nil {
					err = datadogLinux.Setup()
				}
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
			var datadogLinux *adapters.DatadogAdapter
			var isActive bool
			var logger interfaces.ILogger
			var err error
			s.When("logger é criado", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.Given("DatadogAdapter válido e setup executado", func() {
				datadogLinux = adapters.NewDatadogAdapter(logger)
				err = datadogLinux.Setup()
			})
			s.When("IsActive é chamado", func() {
				isActive, err = datadogLinux.IsActive()
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
			var datadogLinux *adapters.DatadogAdapter
			var logger interfaces.ILogger
			var err error
			s.When("logger é criado", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.Given("DatadogAdapter válido e setup executado", func() {
				datadogLinux = adapters.NewDatadogAdapter(logger)
				err = datadogLinux.Setup()
			})
			s.When("InstallAgent é chamado", func() {
				if err == nil {
					err = datadogLinux.InstallAgent("datadoghq.com", "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
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
			var datadogLinux *adapters.DatadogAdapter
			var logger interfaces.ILogger
			var err error
			s.When("logger é criado", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.Given("DatadogAdapter válido e setup executado", func() {
				datadogLinux = adapters.NewDatadogAdapter(logger)
				err = datadogLinux.Setup()
			})
			s.When("UninstallAgent é chamado", func() {
				if err == nil {
					err = datadogLinux.UninstallAgent()
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
			var datadogLinux *adapters.DatadogAdapter
			var logger interfaces.ILogger
			var filePath string
			var err error
			s.When("logger é criado", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.Given("DatadogAdapter válido e setup executado", func() {
				datadogLinux = adapters.NewDatadogAdapter(logger)
				err = datadogLinux.Setup()
			})
			s.When("DiscoverDatadogConfigPath é chamado", func() {
				filePath, err = datadogLinux.DiscoverDatadogConfigPath()
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
			var datadogLinux *adapters.DatadogAdapter
			var logger interfaces.ILogger
			var decoded []byte
			var err error
			s.When("logger é criado", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.Given("DatadogAdapter válido e setup executado", func() {
				datadogLinux = adapters.NewDatadogAdapter(logger)
				err = datadogLinux.Setup()
			})
			s.When("DecodeBase64 é chamado", func() {
				decoded, err = datadogLinux.DecodeBase64("dGVzdGU=")
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
			var datadogLinux *adapters.DatadogAdapter
			var logger interfaces.ILogger
			var err error
			s.When("logger é criado", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.Given("DatadogAdapter válido e setup executado", func() {
				datadogLinux = adapters.NewDatadogAdapter(logger)
				err = datadogLinux.Setup()

			})
			s.When("UpdateConfigFileDatadog é chamado", func() {
				filePath := "/etc/datadog-agent/conf.d/journald.d/conf.yaml"

				err = datadogLinux.UpdateConfigFileDatadog(filePath)

			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "UpdateConfigFileDatadog não deve retornar erro")
			})
		})
	})
}
