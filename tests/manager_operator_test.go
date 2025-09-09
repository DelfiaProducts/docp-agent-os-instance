package tests

import (
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/operators"
)

func TestNewManagerOperator(t *testing.T) {
	bdd.Feature(t, "Criar novo ManagerOperator", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		var linux *operators.ManagerOperator
		Scenario("Operador não deve ser nulo", func(s *bdd.Scenario) {
			s.Given("um ManagerOperator é criado", func() {
				linux = operators.NewManagerOperator()

			})
			s.Then("o operador não deve ser nulo", func(t *testing.T) {
				bdd.AssertIsNotNil(t, linux, "operator não pode ser nulo")
			})
		})
	})
}

func TestManagerOperatorSetup(t *testing.T) {
	bdd.Feature(t, "Setup do ManagerOperator", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Setup não deve retornar erro", func(s *bdd.Scenario) {
			var linux *operators.ManagerOperator
			var err error
			s.Given("um ManagerOperator válido", func() {
				linux = operators.NewManagerOperator()
			})
			s.When("Setup é chamado", func() {
				err = linux.Setup()
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
		})
	})
}

func TestManagerOperatorRun(t *testing.T) {
	bdd.Feature(t, "Executar Run do ManagerOperator", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Run e Setup não devem retornar erro", func(s *bdd.Scenario) {
			var linux *operators.ManagerOperator
			var err error
			s.Given("um ManagerOperator válido", func() {
				linux = operators.NewManagerOperator()
			})
			s.When("Setup é chamado", func() {
				err = linux.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("Run é chamado", func() {
				err = linux.Run()
			})
			s.Then("Run não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "Run não deve retornar erro")
			})
		})
	})
}

func TestManagerOperatorStart(t *testing.T) {
	bdd.Feature(t, "Start do ManagerOperator", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Start não deve retornar erro no Setup", func(s *bdd.Scenario) {
			var linux *operators.ManagerOperator
			var err error
			s.Given("um ManagerOperator válido", func() {
				linux = operators.NewManagerOperator()
			})
			s.When("Setup é chamado", func() {
				err = linux.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("Start é chamado", func() {
				linux.Start()
			})
			s.Then("Start não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "Start não deve retornar erro")
			})
		})
	})
}

func TestManagerOperatorGetActions(t *testing.T) {
	bdd.Feature(t, "GetActions do ManagerOperator", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("GetActions não deve retornar erro e actions deve ser retornado", func(s *bdd.Scenario) {
			var linux *operators.ManagerOperator
			var actions []byte
			var err error
			s.Given("um ManagerOperator válido e setup executado", func() {
				linux = operators.NewManagerOperator()
			})
			s.When("Setup é chamado", func() {
				err = linux.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("GetActions é chamado", func() {
				actions, err = linux.GetActions()
			})
			s.Then("não deve retornar erro e actions deve ser retornado", func(t *testing.T) {
				bdd.AssertNoError(t, err, "GetActions não deve retornar erro")
			})
			bdd.Printf("actions: %s\n", string(actions))
		})
	})
}

func TestManagerOperatorCheckHealth(t *testing.T) {
	bdd.Feature(t, "CheckHealth do ManagerOperator", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("CheckHealth não deve causar panic ou erro", func(s *bdd.Scenario) {
			var linux *operators.ManagerOperator
			var err error
			s.Given("um ManagerOperator válido", func() {
				linux = operators.NewManagerOperator()
			})
			s.When("Setup é chamado", func() {
				err = linux.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("CheckHealth é chamado", func() {
				linux.CheckHealth()
			})
		})
	})
}

func TestManagerOperatorUpdateAgent(t *testing.T) {
	bdd.Feature(t, "UpdateAgent do ManagerOperator", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("UpdateAgent não deve causar panic ou erro", func(s *bdd.Scenario) {
			var linux *operators.ManagerOperator
			var err error

			s.Given("um ManagerOperator válido", func() {
				linux = operators.NewManagerOperator()
			})

			s.When("Setup é chamado", func() {
				err = linux.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})

			s.When("UpdateAgent é chamado", func() {
				linux.WaitGroupAdd(1)
				err = linux.UpdateAgent("0.1.0")
				linux.WaitGroupWait()
			})

			s.Then("não deve retornar erro no update agent", func(t *testing.T) {
				bdd.AssertNoError(t, err, "UpdateAgent não deve retornar erro")
			})
		})
	})
}

func TestManagerOperatorAutoUpdateAgentVersion(t *testing.T) {
	bdd.Feature(t, "AutoUpdateAgentVersion do ManagerOperator", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("AutoUpdateAgentVersion não deve causar panic ou erro", func(s *bdd.Scenario) {
			var linux *operators.ManagerOperator
			var err error
			s.Given("um ManagerOperator válido", func() {
				linux = operators.NewManagerOperator()
			})
			s.When("Setup é chamado", func() {
				err = linux.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("AutoUpdateAgentVersion é chamado", func() {
				err = linux.AutoUpdateAgentVersion()
			})
			s.Then("não deve retornar erro no auto update agent version", func(t *testing.T) {
				bdd.AssertNoError(t, err, "AutoUpdateAgentVersion não deve retornar erro")
			})
		})
	})
}

func TestManagerOperatorUpdateAgentVersionDatadog(t *testing.T) {
	bdd.Feature(t, "UpdateAgentVersionDatadog do ManagerOperator", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("UpdateAgentVersionDatadog não deve causar panic ou erro", func(s *bdd.Scenario) {
			var linux *operators.ManagerOperator
			var err error
			s.Given("um ManagerOperator válido", func() {
				linux = operators.NewManagerOperator()
			})
			s.When("Setup é chamado", func() {
				err = linux.Setup()
				bdd.AssertNoError(t, err, "Setup não deve retornar erro")
			})
			s.When("UpdateAgentVersionDatadog é chamado", func() {
				linux.WaitGroupAdd(1)
				err = linux.UpdateAgentVersionDatadog("7.69.4")
				linux.WaitGroupWait()
			})
			s.Then("não deve retornar erro no update agent version datadog", func(t *testing.T) {
				bdd.AssertNoError(t, err, "UpdateAgentVersionDatadog não deve retornar erro")
			})
		})
	})
}
