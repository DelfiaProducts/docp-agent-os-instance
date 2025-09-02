package tests

import (
	"os"
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/dto"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/services"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

func TestNewUtilityService(t *testing.T) {
	bdd.Feature(t, "Criar serviço de utilitário", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("instanciar serviço de utilitário com logger", func(s *bdd.Scenario) {
			var logger interfaces.ILogger
			var utilityService *services.UtilityService
			s.Given("que eu tenho um logger", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
			})
			s.When("eu instancio o serviço de utilitário", func() {
				utilityService = services.NewUtilityService(logger)
			})
			s.Then("o serviço de utilitário deve ser criado com sucesso", func(t *testing.T) {
				bdd.AssertIsNotNil(t, utilityService, "o serviço de utilitário deve ser diferente de nil")
			})
		})
	})
}

func TestUtilityServiceSetup(t *testing.T) {
	bdd.Feature(t, "Configurar serviço de utilitário", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("configuração bem-sucedida", func(s *bdd.Scenario) {
			var err error
			var logger interfaces.ILogger
			var utilityService *services.UtilityService
			s.Given("eu crio o utility service", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
				utilityService = services.NewUtilityService(logger)
			})
			s.When("eu configuro o utility service", func() {
				err = utilityService.Setup()
			})
			s.Then("a configuração deve ser executada sem erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "configuração do serviço deve ser executada sem erro")
			})
		})
	})
}

func TestUtilityServiceFetchAgentVersions(t *testing.T) {
	bdd.Feature(t, "Buscar versões do agente", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("busca bem-sucedida", func(s *bdd.Scenario) {
			var err error
			var logger interfaces.ILogger
			var utilityService *services.UtilityService
			var versions dto.AgentVersions
			s.Given("eu crio o utility service", func() {
				logger = utils.NewDocpLoggerText(os.Stdout)
				utilityService = services.NewUtilityService(logger)
			})
			s.When("eu configuro o utility service", func() {
				err = utilityService.Setup()
				bdd.AssertNoError(t, err, "configuração do serviço deve ser executada sem erro")
			})
			s.When("eu busco as versions", func() {
				versions, err = utilityService.FetchAgentVersions()
			})
			s.Then("as versões devem ser retornadas com sucesso", func(t *testing.T) {
				bdd.AssertNoError(t, err, "busca das versões do agente deve ser executada sem erro")
				bdd.AssertIsNotNil(t, versions, "versões do agente devem ser diferentes de nil")
			})
			bdd.Printf("Versões do agente: %+v\n", versions)
		})
	})
}
