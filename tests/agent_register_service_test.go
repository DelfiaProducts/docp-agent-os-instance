package tests

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/dto"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/services"
)

func TestNewAgentRegisterService(t *testing.T) {
	bdd.Feature(t, "Criar novo AgentRegisterService", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Service não deve ser nulo", func(s *bdd.Scenario) {
			s.Given("um AgentRegisterService é criado", func() {
				agRegister := services.NewAgentRegisterService(logger)
				s.Then("o service não deve ser nulo", func(t *testing.T) {
					bdd.AssertIsNotNil(t, agRegister, "Service deve ser diferente de nil")
				})
			})
		})
	})
}

func TestAgentRegisterServiceSetup(t *testing.T) {
	bdd.Feature(t, "Setup do AgentRegisterService", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Setup não deve retornar erro", func(s *bdd.Scenario) {
			var agRegister *services.AgentRegisterService
			s.Given("um AgentRegisterService válido", func() {
				agRegister = services.NewAgentRegisterService(logger)
			})
			s.When("Setup é chamado", func() {
				err := agRegister.Setup()
				s.Then("não deve retornar erro", func(t *testing.T) {
					bdd.AssertNoError(t, err, "Setup não deve retornar erro")
				})
			})
		})
	})
}

func TestAgentRegisterServiceGetConfigFileContent(t *testing.T) {
	bdd.Feature(t, "GetConfigFileContent do AgentRegisterService", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve obter conteúdo do arquivo de configuração sem erro", func(s *bdd.Scenario) {
			var agRegister *services.AgentRegisterService
			var configFileContent []byte
			var err error
			s.Given("um AgentRegisterService válido e setup executado", func() {
				agRegister = services.NewAgentRegisterService(logger)
				err = agRegister.Setup()
			})
			s.When("GetConfigFileContent é chamado", func() {
				configFileContent, err = agRegister.GetConfigFileContent("/opt/docp-agent/config.yml")
			})
			s.Then("não deve retornar erro e deve retornar conteúdo", func(t *testing.T) {
				bdd.AssertNoError(t, err, "GetConfigFileContent não deve retornar erro")
				bdd.AssertIsNotNil(t, configFileContent, "Conteúdo não deve ser nil")
				log.Printf("configFileContent: %+v\n", string(configFileContent))
			})
		})
	})
}

func TestAgentRegisterServiceInjectClientInfo(t *testing.T) {
	bdd.Feature(t, "InjectClientInfo do AgentRegisterService", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve injetar informações do cliente sem erro", func(s *bdd.Scenario) {
			var agRegister *services.AgentRegisterService
			var contentBytes []byte
			var err error
			s.Given("um AgentRegisterService válido e setup executado", func() {
				agRegister = services.NewAgentRegisterService(logger)
				err = agRegister.Setup()
			})
			s.When("InjectClientInfoUpdate é chamado", func() {
				linuxMetadata := dto.Metadata{
					ComputeInfo: dto.ComputeInfo{
						Computename:     "my-machine",
						PlatformVersion: "v0.0.1",
					},
				}
				bytesLinuxMetadata, errMarshal := json.Marshal(linuxMetadata)
				if errMarshal != nil {
					err = errMarshal
					return
				}
				f, errOpen := os.OpenFile("/opt/docp-agent/config.yml", os.O_RDONLY, os.ModePerm)
				if errOpen != nil {
					err = errOpen
					return
				}
				defer f.Close()
				configFileContent, errRead := io.ReadAll(f)
				if errRead != nil {
					err = errRead
					return
				}
				contentBytes, _, err = agRegister.InjectClientInfoUpdate(configFileContent, bytesLinuxMetadata)
			})
			s.Then("não deve retornar erro e deve retornar conteúdo", func(t *testing.T) {
				bdd.AssertNoError(t, err, "InjectClientInfoUpdate não deve retornar erro")
				bdd.AssertIsNotNil(t, contentBytes, "Conteúdo não deve ser nil")
				log.Printf("contentBytes: %+v\n", string(contentBytes))
			})
		})
	})
}

func TestAgentRegisterServiceSendMetadataCreate(t *testing.T) {
	bdd.Feature(t, "SendMetadataCreate do AgentRegisterService", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve enviar metadados de criação sem erro", func(s *bdd.Scenario) {
			var agRegister *services.AgentRegisterService
			var resp any
			var err error
			s.Given("um AgentRegisterService válido e setup executado", func() {
				agRegister = services.NewAgentRegisterService(logger)
				err = agRegister.Setup()
			})
			s.When("SendMetadataCreate é chamado", func() {
				data := make(map[string]interface{})
				data["hostname"] = "my-machine"
				data["version"] = "v0.0.1"
				bytesData, errMarshal := json.Marshal(data)
				if errMarshal != nil {
					err = errMarshal
					return
				}
				resp, _, err = agRegister.SendMetadataCreate(bytesData)
			})
			s.Then("não deve retornar erro e deve retornar resposta", func(t *testing.T) {
				bdd.AssertNoError(t, err, "SendMetadataCreate não deve retornar erro")
				bdd.AssertIsNotNil(t, resp, "Resposta não deve ser nil")
				log.Printf("resp: %+v\n", resp)
			})
		})
	})
}

func TestAgentRegisterServiceSendMetadataUpdate(t *testing.T) {
	bdd.Feature(t, "SendMetadataUpdate do AgentRegisterService", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve enviar metadados de atualização sem erro", func(s *bdd.Scenario) {
			var agRegister *services.AgentRegisterService
			var resp any
			var err error
			s.Given("um AgentRegisterService válido e setup executado", func() {
				agRegister = services.NewAgentRegisterService(logger)
				err = agRegister.Setup()
			})
			s.When("SendMetadataUpdate é chamado", func() {
				data := make(map[string]interface{})
				data["hostname"] = "my-machine"
				data["version"] = "v0.0.1"
				bytesData, errMarshal := json.Marshal(data)
				if errMarshal != nil {
					err = errMarshal
					return
				}
				resp, _, err = agRegister.SendMetadataUpdate(bytesData)
			})
			s.Then("não deve retornar erro e deve retornar resposta", func(t *testing.T) {
				bdd.AssertNoError(t, err, "SendMetadataUpdate não deve retornar erro")
				bdd.AssertIsNotNil(t, resp, "Resposta não deve ser nil")
				log.Printf("resp: %+v\n", resp)
			})
		})
	})
}
